package biz

import (
	"context"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/antsurge/weaver-admin/pkg/metrics"
	"github.com/go-kratos/kratos/v2/log"
)

// ============ 服务监控 ============

type ServerInfo struct {
	Hostname       string
	OS             string
	Arch           string
	GoVersion      string
	NumCPU         int
	NumGoroutine   int
	GoMaxProcs     int
	HeapAlloc      int64
	HeapSys        int64
	HeapInuse      int64
	StackInuse     int64
	Sys            int64
	TotalAlloc     int64
	NumGC          uint32
	GCPauseTotalMs int64
	StartTime      time.Time
	UptimeSeconds  int64
}

// ============ 缓存监控 ============

type CacheInfo struct {
	Version          string
	Mode             string
	Role             string
	UptimeSeconds    int64
	ConnectedClients int
	UsedMemory       int64
	UsedMemoryHuman  string
	MaxMemory        int64
	TotalKeys        int64
	KeyspaceHits     int64
	KeyspaceMisses   int64
	HitRate          float64
	TotalCommands    int64
	ExpiredKeys      int64
	EvictedKeys      int64
	OpsPerSec        int64
}

type CacheKey struct {
	Key        string
	RawKey     string
	Type       string
	TTLSeconds int64
	Size       int64
	Value      string
}

// ============ 数据库监控 ============

type DbTableStat struct {
	Name      string
	Rows      int64
	DataSize  int64
	IndexSize int64
}

type DatabaseInfo struct {
	Driver             string
	Version            string
	Database           string
	TableCount         int
	MaxOpenConnections int
	OpenConnections    int
	InUse              int
	Idle               int
	WaitCount          int64
	WaitDurationMs     int64
	MaxIdleClosed      int64
	MaxLifetimeClosed  int64
	Tables             []*DbTableStat
}

// ============ 定时任务监控 ============

type JobRunLog struct {
	ID        string
	JobName   string
	JobGroup  string
	Status    string
	Message   string
	StartTime string
	Duration  int64
}

type JobMonitorInfo struct {
	TotalJobs    int64
	EnabledJobs  int64
	DisabledJobs int64
	TotalRuns    int64
	SuccessRuns  int64
	FailedRuns   int64
	SuccessRate  float64
	RunningJobs  int
	RecentRuns   []*JobRunLog
}

// NewApiMetricsCollector 提供进程内接口指标采集器（单例，供中间件与监控用例共用）。
func NewApiMetricsCollector() *metrics.Collector {
	return metrics.NewCollector()
}

// MonitorRepo 监控数据源（缓存/数据库/任务统计）。
type MonitorRepo interface {
	CacheInfo(ctx context.Context) (*CacheInfo, error)
	CacheKeys(ctx context.Context, pattern string) ([]*CacheKey, error)
	DatabaseInfo(ctx context.Context) (*DatabaseInfo, error)
	// JobMonitor 定时任务统计；name/group 均为可选模糊条件
	JobMonitor(ctx context.Context, name, group string) (*JobMonitorInfo, error)
}

// MonitorUsecase 系统监控用例
type MonitorUsecase struct {
	repo      MonitorRepo
	collector *metrics.Collector
	scheduler *JobScheduler
	startedAt time.Time
	log       *log.Helper
}

func NewMonitorUsecase(
	repo MonitorRepo,
	collector *metrics.Collector,
	scheduler *JobScheduler,
	logger log.Logger,
) *MonitorUsecase {
	return &MonitorUsecase{
		repo:      repo,
		collector: collector,
		scheduler: scheduler,
		startedAt: time.Now(),
		log:       log.NewHelper(logger),
	}
}

// ServerInfo 服务运行状态（Go 运行时）
func (uc *MonitorUsecase) ServerInfo(_ context.Context) (*ServerInfo, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	hostname, _ := os.Hostname()
	return &ServerInfo{
		Hostname:       hostname,
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
		GoVersion:      runtime.Version(),
		NumCPU:         runtime.NumCPU(),
		NumGoroutine:   runtime.NumGoroutine(),
		GoMaxProcs:     runtime.GOMAXPROCS(0),
		HeapAlloc:      int64(m.HeapAlloc),
		HeapSys:        int64(m.HeapSys),
		HeapInuse:      int64(m.HeapInuse),
		StackInuse:     int64(m.StackInuse),
		Sys:            int64(m.Sys),
		TotalAlloc:     int64(m.TotalAlloc),
		NumGC:          m.NumGC,
		GCPauseTotalMs: int64(m.PauseTotalNs / 1e6),
		StartTime:      uc.startedAt,
		UptimeSeconds:  int64(time.Since(uc.startedAt).Seconds()),
	}, nil
}

// CacheInfo 缓存状态
func (uc *MonitorUsecase) CacheInfo(ctx context.Context) (*CacheInfo, error) {
	return uc.repo.CacheInfo(ctx)
}

// CacheKeys 缓存键列表（分页）
func (uc *MonitorUsecase) CacheKeys(ctx context.Context, pattern string, page, pageSize int64) ([]*CacheKey, int64, error) {
	all, err := uc.repo.CacheKeys(ctx, pattern)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(all))
	if pageSize <= 0 {
		return all, total, nil
	}
	start := (page - 1) * pageSize
	if start < 0 {
		start = 0
	}
	if start >= total {
		return []*CacheKey{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

// DatabaseInfo 数据库状态
func (uc *MonitorUsecase) DatabaseInfo(ctx context.Context) (*DatabaseInfo, error) {
	return uc.repo.DatabaseInfo(ctx)
}

// ApiMetrics 接口指标（进程内统计，重启清零）。
// method/path 均为可选过滤条件：method 精确匹配（忽略大小写），path 模糊匹配。
func (uc *MonitorUsecase) ApiMetrics(_ context.Context, method, path string) (items []metrics.Metric, totalRequests, totalErrors, uptimeSeconds int64) {
	all, totalRequests, totalErrors := uc.collector.Snapshot()
	uptimeSeconds = int64(time.Since(uc.collector.StartedAt()).Seconds())
	method = strings.TrimSpace(strings.ToUpper(method))
	path = strings.TrimSpace(path)
	if method == "" && path == "" {
		return all, totalRequests, totalErrors, uptimeSeconds
	}
	items = make([]metrics.Metric, 0, len(all))
	for _, m := range all {
		if method != "" && m.Method != method {
			continue
		}
		if path != "" && !strings.Contains(m.Path, path) {
			continue
		}
		items = append(items, m)
	}
	return items, totalRequests, totalErrors, uptimeSeconds
}

// JobMonitor 定时任务监控；name/group 均为可选模糊条件
func (uc *MonitorUsecase) JobMonitor(ctx context.Context, name, group string) (*JobMonitorInfo, error) {
	info, err := uc.repo.JobMonitor(ctx, strings.TrimSpace(name), strings.TrimSpace(group))
	if err != nil {
		return nil, err
	}
	if uc.scheduler != nil {
		info.RunningJobs = uc.scheduler.RunningCount()
	}
	return info, nil
}
