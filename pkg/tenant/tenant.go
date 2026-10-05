// Package tenant 提供多租户上下文与作用域控制。
//
// 设计约定（见 docs/bugfix-log.md 多租户方案）：
//   - 所有请求默认在某个租户作用域内（HTTP 链路由 tenant 中间件注入）；
//   - 平台级接口/后台任务可通过 Unscoped 显式声明"平台作用域"，不做租户过滤；
//   - 写操作若拿不到租户作用域必须报错，避免数据静默写入错误租户。
package tenant

import "context"

const (
	// DefaultTenantID 默认（系统）租户。单租户模式下有且仅有该租户。
	DefaultTenantID = "default"

	// PlatformTenantID 平台级数据保留值（如平台级菜单）。
	PlatformTenantID = "platform"

	// HeaderName 平台超管切换租户时携带的请求头。
	HeaderName = "X-Tenant-ID"
)

type ctxKey struct{}

type unscopedKey struct{}

// With 将租户 ID 写入上下文（HTTP 链路由中间件调用）。
func With(ctx context.Context, tenantID string) context.Context {
	if tenantID == "" {
		return ctx
	}
	return context.WithValue(ctx, ctxKey{}, tenantID)
}

// From 读取上下文中的租户 ID；第二个返回值表示是否已显式设置。
func From(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxKey{}).(string)
	return v, ok && v != ""
}

// Unscoped 标记为平台作用域：隔离层不做租户过滤。
//
// 使用限制：仅允许平台级接口与后台任务使用；HTTP 业务链路禁止使用，
// 否则会绕过租户隔离。
func Unscoped(ctx context.Context) context.Context {
	return context.WithValue(ctx, unscopedKey{}, true)
}

// IsUnscoped 判断上下文是否为平台作用域。
func IsUnscoped(ctx context.Context) bool {
	v, _ := ctx.Value(unscopedKey{}).(bool)
	return v
}
