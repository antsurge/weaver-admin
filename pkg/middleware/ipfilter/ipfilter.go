package ipfilter

import (
	"context"
	"net"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// Checker 判断来源 IP 是否允许访问（由业务侧实现，通常读取安全策略并带缓存）。
type Checker interface {
	IsIPAllowed(ctx context.Context, ip string) (bool, error)
}

// CheckerFunc 函数适配器，便于将业务方法直接注册为 Checker。
type CheckerFunc func(ctx context.Context, ip string) (bool, error)

// IsIPAllowed 实现 Checker 接口
func (f CheckerFunc) IsIPAllowed(ctx context.Context, ip string) (bool, error) {
	return f(ctx, ip)
}

// ErrIPForbidden 当前 IP 不允许访问。
var ErrIPForbidden = errors.Forbidden("IP_FORBIDDEN", "当前 IP 不允许访问")

// Options 中间件配置
type Options struct {
	checker Checker
}

// Option 中间件配置项
type Option func(*Options)

// WithChecker 设置 IP 校验器（未设置时中间件直接放行）
func WithChecker(c Checker) Option {
	return func(o *Options) {
		o.checker = c
	}
}

// Server 构建 IP 白/黑名单校验中间件。
//
// 行为约定：
//   - 未配置 checker 时直接放行（等价于策略 off）
//   - 仅对 HTTP 请求生效
//   - 校验失败（如存储不可用）时 fail-open（放行），避免策略存储抖动导致全站不可用；
//     策略本身为"关闭"时同样放行
func Server(opts ...Option) middleware.Middleware {
	o := &Options{}
	for _, opt := range opts {
		opt(o)
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			if o.checker == nil {
				return handler(ctx, req)
			}

			ip := ClientIP(ctx)
			allowed, err := o.checker.IsIPAllowed(ctx, ip)
			if err != nil {
				// fail-open：存储不可用时不阻断业务
				return handler(ctx, req)
			}
			if !allowed {
				return nil, ErrIPForbidden
			}
			return handler(ctx, req)
		}
	}
}

// ClientIP 从请求头中提取客户端真实 IP：
// 依次尝试 X-Forwarded-For（取第一个）→ X-Real-IP → RemoteAddr。
func ClientIP(ctx context.Context) string {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return ""
	}
	h := tr.RequestHeader()
	if v := h.Get("X-Forwarded-For"); v != "" {
		if idx := strings.Index(v, ","); idx > 0 {
			return strings.TrimSpace(v[:idx])
		}
		return strings.TrimSpace(v)
	}
	if v := h.Get("X-Real-IP"); v != "" {
		return strings.TrimSpace(v)
	}
	if ht, ok := tr.(*khttp.Transport); ok && ht.Request() != nil {
		if host, _, err := net.SplitHostPort(ht.Request().RemoteAddr); err == nil {
			return host
		}
		return ht.Request().RemoteAddr
	}
	return ""
}
