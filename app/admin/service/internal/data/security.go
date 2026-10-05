package data

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/conf"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/securitypolicy"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// policyCacheTTL 策略缓存时长：保存时主动失效，这里再兜底一个短 TTL。
const policyCacheTTL = 30 * time.Second

var _ biz.SecurityRepo = (*securityRepo)(nil)

type securityRepo struct {
	data *Data
	log  *log.Helper
	conf *conf.Data
}

func NewSecurityRepo(data *Data, logger log.Logger, c *conf.Data) biz.SecurityRepo {
	return &securityRepo{
		data: data,
		log:  log.NewHelper(logger),
		conf: c,
	}
}

// GetPolicy 读取安全策略：优先 Redis 缓存，未命中回源 DB 并回填缓存。
// 无记录返回 (nil, nil)。
func (r *securityRepo) GetPolicy(ctx context.Context) (*biz.SecurityPolicy, error) {
	cacheKey := r.policyCacheKey()
	if v, err := r.data.redis.Get(ctx, cacheKey).Result(); err == nil && v != "" {
		var p biz.SecurityPolicy
		if err := json.Unmarshal([]byte(v), &p); err == nil {
			return &p, nil
		}
	} else if err != nil && err != redis.Nil {
		// 缓存不可用不影响主流程，继续回源 DB
		r.log.Errorf("security policy cache get error: %v", err)
	}

	row, err := r.data.db.SecurityPolicy.Query().
		Where(securitypolicy.IDEQ(biz.SecurityPolicyID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		r.log.Errorf("security policy query error: %v", err)
		return nil, err
	}

	p := toBizSecurityPolicy(row)
	if b, err := json.Marshal(p); err == nil {
		if err := r.data.redis.Set(ctx, cacheKey, b, policyCacheTTL).Err(); err != nil {
			r.log.Errorf("security policy cache set error: %v", err)
		}
	}
	return p, nil
}

// SavePolicy 保存安全策略（单例 upsert），成功后失效缓存。
func (r *securityRepo) SavePolicy(ctx context.Context, p *biz.SecurityPolicy) error {
	ipList := strings.Join(splitIPList(p.IPList), "\n")
	now := time.Now()

	exists, err := r.data.db.SecurityPolicy.Query().
		Where(securitypolicy.IDEQ(biz.SecurityPolicyID)).
		Exist(ctx)
	if err != nil {
		return err
	}

	if exists {
		err = r.data.db.SecurityPolicy.UpdateOneID(biz.SecurityPolicyID).
			SetLoginFailEnabled(p.LoginFailEnabled).
			SetLoginFailMax(p.LoginFailMax).
			SetLoginFailWindowMinutes(p.LoginFailWindowMinutes).
			SetLoginLockMinutes(p.LoginLockMinutes).
			SetIPMode(securitypolicy.IPMode(p.IPMode)).
			SetIPList(ipList).
			SetAccessTokenTTLSeconds(p.AccessTokenTTLSeconds).
			SetRefreshTokenTTLSeconds(p.RefreshTokenTTLSeconds).
			SetRemark(p.Remark).
			SetUpdatedAt(now).
			Exec(ctx)
	} else {
		_, err = r.data.db.SecurityPolicy.Create().
			SetID(biz.SecurityPolicyID).
			SetLoginFailEnabled(p.LoginFailEnabled).
			SetLoginFailMax(p.LoginFailMax).
			SetLoginFailWindowMinutes(p.LoginFailWindowMinutes).
			SetLoginLockMinutes(p.LoginLockMinutes).
			SetIPMode(securitypolicy.IPMode(p.IPMode)).
			SetIPList(ipList).
			SetAccessTokenTTLSeconds(p.AccessTokenTTLSeconds).
			SetRefreshTokenTTLSeconds(p.RefreshTokenTTLSeconds).
			SetRemark(p.Remark).
			SetCreatedAt(now).
			SetUpdatedAt(now).
			Save(ctx)
	}
	if err != nil {
		return err
	}
	return r.invalidatePolicyCache(ctx)
}

// IncrLoginFail 累加登录失败次数（窗口内），返回累加后的值。
func (r *securityRepo) IncrLoginFail(ctx context.Context, username string, ttl time.Duration) (int64, error) {
	key := r.loginFailKey(username)
	pipe := r.data.redis.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		r.log.Errorf("incr login fail error: %v", err)
		return 0, err
	}
	return incr.Val(), nil
}

// ResetLoginFail 清零登录失败次数。
func (r *securityRepo) ResetLoginFail(ctx context.Context, username string) error {
	if err := r.data.redis.Del(ctx, r.loginFailKey(username)).Err(); err != nil {
		r.log.Errorf("reset login fail error: %v", err)
		return err
	}
	return nil
}

// IsLoginLocked 查询锁定期及剩余时间。
func (r *securityRepo) IsLoginLocked(ctx context.Context, username string) (bool, time.Duration, error) {
	ttl, err := r.data.redis.TTL(ctx, r.loginLockKey(username)).Result()
	if err != nil {
		if err == redis.Nil {
			return false, 0, nil
		}
		r.log.Errorf("login lock ttl error: %v", err)
		return false, 0, err
	}
	if ttl > 0 {
		return true, ttl, nil
	}
	return false, 0, nil
}

// LockLogin 锁定账号指定时长。
func (r *securityRepo) LockLogin(ctx context.Context, username string, ttl time.Duration) error {
	if err := r.data.redis.Set(ctx, r.loginLockKey(username), "1", ttl).Err(); err != nil {
		r.log.Errorf("lock login error: %v", err)
		return err
	}
	return nil
}

func (r *securityRepo) invalidatePolicyCache(ctx context.Context) error {
	if err := r.data.redis.Del(ctx, r.policyCacheKey()).Err(); err != nil {
		r.log.Errorf("security policy cache invalidate error: %v", err)
		return err
	}
	return nil
}

func (r *securityRepo) prefix() string {
	return fmt.Sprintf("%ssecurity:", r.conf.Redis.KeyPrefix)
}

func (r *securityRepo) policyCacheKey() string {
	return r.prefix() + "policy"
}

func (r *securityRepo) loginFailKey(username string) string {
	return r.prefix() + "login:fail:" + username
}

func (r *securityRepo) loginLockKey(username string) string {
	return r.prefix() + "login:lock:" + username
}

// toBizSecurityPolicy ent -> biz
func toBizSecurityPolicy(row *ent.SecurityPolicy) *biz.SecurityPolicy {
	return &biz.SecurityPolicy{
		LoginFailEnabled:       row.LoginFailEnabled,
		LoginFailMax:           row.LoginFailMax,
		LoginFailWindowMinutes: row.LoginFailWindowMinutes,
		LoginLockMinutes:       row.LoginLockMinutes,
		IPMode:                 string(row.IPMode),
		IPList:                 splitIPList(strings.FieldsFunc(row.IPList, func(r rune) bool { return r == '\n' || r == ',' })),
		AccessTokenTTLSeconds:  row.AccessTokenTTLSeconds,
		RefreshTokenTTLSeconds: row.RefreshTokenTTLSeconds,
		Remark:                 row.Remark,
		UpdatedAt:              row.UpdatedAt,
	}
}

// splitIPList 去空白/去重
func splitIPList(in []string) []string {
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
