package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/sysjob"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/sysjoblog"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// maxCacheKeys 缓存列表最多扫描的键数量，避免阻塞 Redis。
const maxCacheKeys = 1000

// maxCacheValueItems 集合类键值（list/set/hash/zset）预览的最大元素数量。
const maxCacheValueItems = 50

// maxCacheValueLen 键值预览的最大长度（字节），超长截断避免响应过大。
const maxCacheValueLen = 500

// recentJobRunLimit 定时任务监控展示的最近执行记录条数。
const recentJobRunLimit = 10

var _ biz.MonitorRepo = (*monitorRepo)(nil)

type monitorRepo struct {
	data *Data
	log  *log.Helper
}

func NewMonitorRepo(data *Data, logger log.Logger) biz.MonitorRepo {
	return &monitorRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// CacheInfo 解析 Redis INFO 获取缓存运行状态。
func (r *monitorRepo) CacheInfo(ctx context.Context) (*biz.CacheInfo, error) {
	resp, err := r.data.redis.Info(ctx).Result()
	if err != nil {
		r.log.Errorf("redis info error: %v", err)
		return nil, err
	}
	kv := parseRedisInfo(resp)

	totalKeys, err := r.data.redis.DBSize(ctx).Result()
	if err != nil {
		r.log.Errorf("redis dbsize error: %v", err)
	}

	hits := atoi64(kv["keyspace_hits"])
	misses := atoi64(kv["keyspace_misses"])
	hitRate := 0.0
	if hits+misses > 0 {
		hitRate = float64(hits) / float64(hits+misses) * 100
	}

	info := &biz.CacheInfo{
		Version:          kv["redis_version"],
		Mode:             kv["redis_mode"],
		Role:             kv["role"],
		UptimeSeconds:    atoi64(kv["uptime_in_seconds"]),
		ConnectedClients: int(atoi64(kv["connected_clients"])),
		UsedMemory:       atoi64(kv["used_memory"]),
		UsedMemoryHuman:  kv["used_memory_human"],
		MaxMemory:        atoi64(kv["maxmemory"]),
		TotalKeys:        totalKeys,
		KeyspaceHits:     hits,
		KeyspaceMisses:   misses,
		HitRate:          round2(hitRate),
		TotalCommands:    atoi64(kv["total_commands_processed"]),
		ExpiredKeys:      atoi64(kv["expired_keys"]),
		EvictedKeys:      atoi64(kv["evicted_keys"]),
		OpsPerSec:        atoi64(kv["instantaneous_ops_per_sec"]),
	}
	if info.Mode == "" {
		info.Mode = "standalone"
	}
	return info, nil
}

// CacheKeys 扫描缓存键（上限 maxCacheKeys），返回键/类型/TTL/内存占用。
func (r *monitorRepo) CacheKeys(ctx context.Context, pattern string) ([]*biz.CacheKey, error) {
	if strings.TrimSpace(pattern) == "" {
		pattern = "*"
	}

	keys := make([]string, 0, 256)
	iter := r.data.redis.Scan(ctx, 0, pattern, 200).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= maxCacheKeys {
			break
		}
	}
	if err := iter.Err(); err != nil {
		r.log.Errorf("cache key scan error: %v", err)
		return nil, err
	}
	if len(keys) == 0 {
		return []*biz.CacheKey{}, nil
	}

	pipe := r.data.redis.Pipeline()
	typeCmds := make([]*redis.StatusCmd, len(keys))
	ttlCmds := make([]*redis.DurationCmd, len(keys))
	memCmds := make([]*redis.IntCmd, len(keys))
	for i, k := range keys {
		typeCmds[i] = pipe.Type(ctx, k)
		ttlCmds[i] = pipe.TTL(ctx, k)
		memCmds[i] = pipe.MemoryUsage(ctx, k)
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		r.log.Errorf("cache key pipeline error: %v", err)
	}

	// 第二轮 pipeline 按类型读取键值（列表/集合类仅取前 maxCacheValueItems 个元素预览），
	// 未知类型（如 stream）跳过取值避免返回二进制乱码。
	valCmds := make([]redis.Cmder, len(keys))
	valueFetched := make([]bool, len(keys))
	for i, k := range keys {
		switch typeCmds[i].Val() {
		case "string":
			valCmds[i] = pipe.Get(ctx, k)
			valueFetched[i] = true
		case "hash":
			valCmds[i] = pipe.HGetAll(ctx, k)
			valueFetched[i] = true
		case "list":
			valCmds[i] = pipe.LRange(ctx, k, 0, maxCacheValueItems-1)
			valueFetched[i] = true
		case "set":
			valCmds[i] = pipe.SMembers(ctx, k)
			valueFetched[i] = true
		case "zset":
			valCmds[i] = pipe.ZRange(ctx, k, 0, maxCacheValueItems-1)
			valueFetched[i] = true
		}
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		r.log.Errorf("cache value pipeline error: %v", err)
	}

	out := make([]*biz.CacheKey, 0, len(keys))
	for i, k := range keys {
		item := &biz.CacheKey{
			Key:    maskCacheKey(k),
			RawKey: k,
			Type:   typeCmds[i].Val(),
			Size:   -1,
		}
		if ttl, err := ttlCmds[i].Result(); err == nil {
			item.TTLSeconds = int64(ttl.Seconds())
			// go-redis 对不存在的 key / 无过期时间均返回 -1ns/-2ns，统一为 -1 表示永不过期
			if item.TTLSeconds < 0 {
				item.TTLSeconds = -1
			}
		} else {
			item.TTLSeconds = -1
		}
		if mem, err := memCmds[i].Result(); err == nil {
			item.Size = mem
		}
		if valueFetched[i] {
			item.Value = truncateCacheValue(formatCacheValue(valCmds[i]))
		}
		out = append(out, item)
	}
	return out, nil
}

// formatCacheValue 将不同类型命令结果格式化为可读字符串。
func formatCacheValue(cmd redis.Cmder) string {
	switch v := cmd.(type) {
	case *redis.StringCmd:
		return v.Val()
	case *redis.MapStringStringCmd:
		b, err := json.Marshal(v.Val())
		if err != nil {
			return fmt.Sprint(v.Val())
		}
		return string(b)
	case *redis.StringSliceCmd:
		b, err := json.Marshal(v.Val())
		if err != nil {
			return fmt.Sprint(v.Val())
		}
		return string(b)
	default:
		return fmt.Sprint(v)
	}
}

// DatabaseInfo 数据库连接池状态 + 表统计。
func (r *monitorRepo) DatabaseInfo(ctx context.Context) (*biz.DatabaseInfo, error) {
	info := &biz.DatabaseInfo{Driver: dialect.MySQL, Tables: []*biz.DbTableStat{}}

	db := r.data.sqlDB
	if db == nil {
		return info, nil
	}
	stats := db.Stats()
	info.MaxOpenConnections = stats.MaxOpenConnections
	info.OpenConnections = stats.OpenConnections
	info.InUse = stats.InUse
	info.Idle = stats.Idle
	info.WaitCount = stats.WaitCount
	info.WaitDurationMs = stats.WaitDuration.Milliseconds()
	info.MaxIdleClosed = stats.MaxIdleClosed
	info.MaxLifetimeClosed = stats.MaxLifetimeClosed

	if err := db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&info.Version); err != nil {
		r.log.Errorf("query db version error: %v", err)
	}
	if err := db.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&info.Database); err != nil {
		r.log.Errorf("query db name error: %v", err)
	}

	rows, err := db.QueryContext(ctx, `
		SELECT TABLE_NAME, TABLE_ROWS, DATA_LENGTH, INDEX_LENGTH
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = DATABASE()
		ORDER BY (IFNULL(DATA_LENGTH,0) + IFNULL(INDEX_LENGTH,0)) DESC`)
	if err != nil {
		r.log.Errorf("query table stats error: %v", err)
		return info, nil
	}
	defer rows.Close()

	for rows.Next() {
		var (
			name       string
			rowCount   sql.NullInt64
			dataLen    sql.NullInt64
			indexLen   sql.NullInt64
			tableStats = &biz.DbTableStat{}
		)
		if err := rows.Scan(&name, &rowCount, &dataLen, &indexLen); err != nil {
			r.log.Errorf("scan table stat error: %v", err)
			continue
		}
		tableStats.Name = name
		tableStats.Rows = rowCount.Int64
		tableStats.DataSize = dataLen.Int64
		tableStats.IndexSize = indexLen.Int64
		info.Tables = append(info.Tables, tableStats)
	}
	info.TableCount = len(info.Tables)
	return info, nil
}

// JobMonitor 定时任务数量与执行统计；name/group 均为可选模糊条件。
func (r *monitorRepo) JobMonitor(ctx context.Context, name, group string) (*biz.JobMonitorInfo, error) {
	info := &biz.JobMonitorInfo{RecentRuns: []*biz.JobRunLog{}}

	totalJobs, err := r.data.db.SysJob.Query().Where(sysjob.DeletedAtIsNil()).Count(ctx)
	if err != nil {
		return nil, err
	}
	enabledJobs, err := r.data.db.SysJob.Query().
		Where(sysjob.DeletedAtIsNil(), sysjob.StatusEQ(sysjob.StatusEnabled)).Count(ctx)
	if err != nil {
		return nil, err
	}
	info.TotalJobs = int64(totalJobs)
	info.EnabledJobs = int64(enabledJobs)
	info.DisabledJobs = info.TotalJobs - info.EnabledJobs

	logQuery := r.data.db.SysJobLog.Query().Where(sysjoblog.DeletedAtIsNil())
	if name != "" {
		logQuery = logQuery.Where(sysjoblog.JobNameContains(name))
	}
	if group != "" {
		logQuery = logQuery.Where(sysjoblog.JobGroupContains(group))
	}
	totalRuns, err := logQuery.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}
	successRuns, err := logQuery.Clone().Where(sysjoblog.StatusEQ(sysjoblog.StatusSuccess)).Count(ctx)
	if err != nil {
		return nil, err
	}
	info.TotalRuns = int64(totalRuns)
	info.SuccessRuns = int64(successRuns)
	info.FailedRuns = info.TotalRuns - info.SuccessRuns
	if info.TotalRuns > 0 {
		info.SuccessRate = round2(float64(info.SuccessRuns) / float64(info.TotalRuns) * 100)
	}

	recent, err := logQuery.
		Order(ent.Desc(sysjoblog.FieldCreatedAt)).
		Limit(recentJobRunLimit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, v := range recent {
		startTime := ""
		if v.StartTime != nil {
			startTime = v.StartTime.Format(time.RFC3339)
		}
		info.RecentRuns = append(info.RecentRuns, &biz.JobRunLog{
			ID:        v.ID,
			JobName:   v.JobName,
			JobGroup:  v.JobGroup,
			Status:    string(v.Status),
			Message:   v.JobMessage,
			StartTime: startTime,
			Duration:  v.Duration,
		})
	}
	return info, nil
}

// maskCacheKey 对超长键名做脱敏（如 token 键内嵌完整 JWT），
// 仅保留首尾片段以便识别，避免监控页面直接暴露敏感凭证。
func maskCacheKey(k string) string {
	const maxLen = 80
	if len(k) <= maxLen {
		return k
	}
	return k[:60] + "..." + k[len(k)-8:]
}

// truncateCacheValue 对超长键值做截断，避免监控页面响应过大。
func truncateCacheValue(s string) string {
	if len(s) <= maxCacheValueLen {
		return s
	}
	// 按 rune 截断，避免截断多字节字符导致乱码
	rs := []rune(s)
	if len(rs) <= maxCacheValueLen {
		return s
	}
	return string(rs[:maxCacheValueLen]) + "..."
}

// parseRedisInfo 解析 Redis INFO 文本为 key/value（忽略 # 段落行）。
func parseRedisInfo(resp string) map[string]string {
	kv := make(map[string]string, 64)
	for _, line := range strings.Split(resp, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		kv[k] = v
	}
	return kv
}

func atoi64(s string) int64 {
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0
	}
	return v
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
