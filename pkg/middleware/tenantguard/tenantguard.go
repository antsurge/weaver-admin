// Package tenantguard 请求期租户状态守卫 + 平台超管跨租户切换。
//
// 作用：
//  1. 每次请求校验当前租户（JWT 中携带）仍处于 enabled 且未过期/未删除，
//     防止登录后租户被禁用/过期仍能继续访问（租户被禁应立即失效）。
//  2. 当平台超管且配置 tenant_superadmin_cross_read=true 时，允许通过
//     X-Tenant-ID 请求头切换到指定租户作用域（平台运维/跨租户只读场景）。
//     未配置/非超管时忽略该请求头（保持强隔离，不降级）。
package tenantguard

import (
	"context"
	"sync"
	"time"

	"github.com/antsurge/weaver-admin/pkg/metadata"
	"github.com/antsurge/weaver-admin/pkg/tenant"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// ErrTenantForbidden 租户被禁用/过期/删除时返回的错误。
var ErrTenantForbidden = errors.New(403, "TENANT_FORBIDDEN", "租户不可用，请联系平台管理员")

// TenantStatusProvider 租户状态查询接口（由 biz.TenantUsecase 适配注入）。
type TenantStatusProvider interface {
	// IsTenantUsable 判断租户当前是否可用（enabled 且未过期且未删除）。
	// 返回 (usable, error)；查询失败时应 fail-open（返回 true）避免雪崩。
	IsTenantUsable(ctx context.Context, tenantID string) (bool, error)
}

// SuperAdminChecker 判断当前登录用户是否为平台超管（用于跨租户切换授权）。
type SuperAdminChecker interface {
	IsSuperAdmin(ctx context.Context, adminID string) (bool, error)
}

// Options 中间件配置。
type Options struct {
	// tenantProvider 租户状态查询器（必填）。
	tenantProvider TenantStatusProvider
	// superAdminChecker 超管判断器（可选；nil 时禁用跨租户切换）。
	superAdminChecker SuperAdminChecker
	// superadminCrossRead 是否允许平台超管跨租户切换（对应配置 tenant_superadmin_cross_read）。
	superadminCrossRead bool
	// tenantMode 租户模式（"multi" 时允许跨租户切换，"single" 时忽略）。
	tenantMode string
	// cacheTTL 租户状态缓存时长（0 表示不缓存）。
	cacheTTL time.Duration
}

type Option func(*Options)

// WithTenantProvider 设置租户状态查询器。
func WithTenantProvider(p TenantStatusProvider) Option {
	return func(o *Options) { o.tenantProvider = p }
}

// WithSuperAdminChecker 设置超管判断器。
func WithSuperAdminChecker(c SuperAdminChecker) Option {
	return func(o *Options) { o.superAdminChecker = c }
}

// WithSuperadminCrossRead 设置是否允许平台超管跨租户切换。
func WithSuperadminCrossRead(enabled bool) Option {
	return func(o *Options) { o.superadminCrossRead = enabled }
}

// WithTenantMode 设置租户模式（"multi"/"single"）。
func WithTenantMode(mode string) Option {
	return func(o *Options) { o.tenantMode = mode }
}

// WithCacheTTL 设置租户状态缓存时长。
func WithCacheTTL(d time.Duration) Option {
	return func(o *Options) { o.cacheTTL = d }
}

func NewOptions(opts ...Option) *Options {
	o := &Options{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// Server 租户守卫中间件。
// 必须挂在 auth 之后（依赖 context 中的 adminID 与 tenantID）。
func Server(opts ...Option) middleware.Middleware {
	o := NewOptions(opts...)
	// 简易内存缓存：tenantID -> (usable, checkedAt)
	cache := newStatusCache()

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			tid, ok := tenant.From(ctx)
			if !ok || tid == "" {
				// 无租户上下文：交由后续中间件处理（登录等白名单接口不在此链）
				return handler(ctx, req)
			}

			// 1. 平台超管跨租户切换（仅当配置开启且为超管）
			if o.superadminCrossRead && o.superAdminChecker != nil {
				if target, ok := trySwitchTenant(ctx, o.superAdminChecker, o.tenantMode); ok {
					// 切换到目标租户后仍需校验目标租户可用
					if usable, err := checkTenant(ctx, o, cache, target); err != nil {
						return nil, err
					} else if usable {
						ctx = tenant.With(ctx, target)
					}
				}
			}

			// 2. 校验当前租户状态（带缓存）
			usable, err := checkTenant(ctx, o, cache, tid)
			if err != nil {
				return nil, err
			}
			if !usable {
				return nil, ErrTenantForbidden
			}

			return handler(ctx, req)
		}
	}
}

// checkTenant 校验租户可用性（带缓存）。
func checkTenant(ctx context.Context, o *Options, cache *statusCache, tid string) (bool, error) {
	// 平台级请求不校验租户状态
	if tenant.IsUnscoped(ctx) {
		return true, nil
	}

	if o.tenantProvider == nil {
		// 未注入查询器时 fail-open（不阻塞请求）
		return true, nil
	}

	if o.cacheTTL > 0 {
		if cached, ok := cache.Get(tid); ok && time.Since(cached.checkedAt) < o.cacheTTL {
			return cached.usable, nil
		}
	}

	usable, err := o.tenantProvider.IsTenantUsable(ctx, tid)
	if err != nil {
		// 查询失败 fail-open，避免中间件异常导致所有请求不可用
		return true, nil
	}

	if o.cacheTTL > 0 {
		cache.Set(tid, usable, time.Now())
	}
	return usable, nil
}

// trySwitchTenant 平台超管跨租户切换。
// 仅当请求头 X-Tenant-ID 存在且用户为平台超管时返回目标租户ID。
func trySwitchTenant(ctx context.Context, checker SuperAdminChecker, tenantMode string) (string, bool) {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return "", false
	}

	// 仅多租户模式下允许切换（single 模式所有数据均在 default 租户，无切换意义）
	if tenantMode == "single" {
		return "", false
	}

	target := tr.RequestHeader().Get(tenant.HeaderName)
	if target == "" || target == tenant.DefaultTenantID {
		return "", false
	}

	// 校验当前用户是否为平台超管
	adminID := metadata.GetAdminID(ctx)
	if adminID == "" {
		return "", false
	}
	isSuper, err := checker.IsSuperAdmin(ctx, adminID)
	if err != nil || !isSuper {
		return "", false
	}

	return target, true
}

// ---------- 简易内存缓存 ----------

type statusEntry struct {
	usable    bool
	checkedAt time.Time
}

type statusCache struct {
	mu sync.RWMutex
	m  map[string]statusEntry
}

func newStatusCache() *statusCache {
	return &statusCache{m: make(map[string]statusEntry)}
}

func (c *statusCache) Get(key string) (statusEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.m[key]
	return v, ok
}

func (c *statusCache) Set(key string, usable bool, at time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = statusEntry{usable: usable, checkedAt: at}
	// 简单防泄漏：超过 10000 条时清空重建
	if len(c.m) > 10000 {
		c.m = make(map[string]statusEntry)
	}
}
