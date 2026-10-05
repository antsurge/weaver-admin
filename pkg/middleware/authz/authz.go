package authz

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"

	"github.com/antsurge/weaver-admin/pkg/metadata"
)

// CodeSeparator 权限码中 method 与 path 的分隔符，格式："METHOD|pathTemplate"，
// 如 "GET|/admin/v1/admin/{adminId}"
const CodeSeparator = "|"

// Permission 用户的接口权限集合（由业务侧提供，通常来源于「角色→菜单→接口权限」关联查询）
type Permission struct {
	// SuperAdmin 超级管理员，拥有全部接口权限
	SuperAdmin bool
	// Codes 接口权限码集合，格式 "METHOD|pathTemplate"
	Codes []string
}

// Checker 权限查询器，由业务侧实现（建议实现缓存以降低每请求查询开销）
type Checker interface {
	GetPermission(ctx context.Context, adminID string) (*Permission, error)
}

// CheckerFunc 函数适配器，便于将业务方法直接注册为 Checker
type CheckerFunc func(ctx context.Context, adminID string) (*Permission, error)

// GetPermission 实现 Checker 接口
func (f CheckerFunc) GetPermission(ctx context.Context, adminID string) (*Permission, error) {
	return f(ctx, adminID)
}

// ErrPermissionDenied 无接口访问权限。
// 登录态校验由 auth 中间件负责，这里只负责"登录之后能不能访问该接口"。
var ErrPermissionDenied = errors.Forbidden("PERMISSION_DENIED", "无接口访问权限")

// Options 中间件配置
type Options struct {
	checker Checker
	// bypass 豁免接口权限校验的自服务接口，key 为 "METHOD|pathTemplate"
	// （Login/GetCaptcha/RefreshToken 等在 auth 白名单中无 adminID，天然跳过，无需在此登记）
	bypass map[string]struct{}
}

// Option 中间件配置项
type Option func(*Options)

// WithChecker 设置权限查询器（必填；未设置时中间件直接放行）
func WithChecker(c Checker) Option {
	return func(o *Options) {
		o.checker = c
	}
}

// WithBypass 登记豁免接口权限校验的自服务接口（如当前用户信息、登出）
func WithBypass(method string, paths ...string) Option {
	return func(o *Options) {
		for _, p := range paths {
			o.bypass[methodKey(method, p)] = struct{}{}
		}
	}
}

// Server 构建接口权限校验中间件。
//
// 匹配模型：ApiPermission 的 Method+Path（来自 OpenAPI 导入）与请求的
// HTTP Method + 路由路径模板（kratos PathTemplate，如 /admin/v1/admin/{adminId}）
// 做比对，路径模板与 OpenAPI path 同源于 google.api.http 注解。
//
// 行为约定：
//   - 未登录（无 adminID，白名单接口）或未配置 checker：直接放行
//   - 仅对 HTTP 请求生效；gRPC 场景的接口权限模型尚未定义，暂不校验
//   - 权限查询失败时 fail-closed（拒绝访问）
func Server(opts ...Option) middleware.Middleware {
	o := &Options{bypass: make(map[string]struct{})}
	for _, opt := range opts {
		opt(o)
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			adminID := metadata.GetAdminID(ctx)
			if adminID == "" || o.checker == nil {
				return handler(ctx, req)
			}

			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return handler(ctx, req)
			}
			ht, ok := tr.(*khttp.Transport)
			if !ok {
				// 非 HTTP 传输（gRPC 等）不做接口权限校验
				return handler(ctx, req)
			}

			method := strings.ToUpper(ht.Request().Method)
			if method == http.MethodOptions {
				return handler(ctx, req)
			}

			pathTemplate := ht.PathTemplate()
			if pathTemplate == "" {
				pathTemplate = ht.Request().URL.Path
			}
			if _, ok := o.bypass[methodKey(method, pathTemplate)]; ok {
				return handler(ctx, req)
			}

			perm, err := o.checker.GetPermission(ctx, adminID)
			if err != nil {
				// fail-closed：无法确认权限时拒绝
				return nil, ErrPermissionDenied
			}
			if perm != nil && perm.SuperAdmin {
				return handler(ctx, req)
			}
			if perm != nil && matchCode(perm.Codes, method, pathTemplate) {
				return handler(ctx, req)
			}
			return nil, ErrPermissionDenied
		}
	}
}

// methodKey 生成 "METHOD|path" 形式的匹配键
func methodKey(method, path string) string {
	return strings.ToUpper(method) + CodeSeparator + path
}

// matchCode 判断请求是否命中权限码集合：
// 1. method 精确匹配（权限码可用 * 表示任意 method）
// 2. path 先做模板串精确匹配，再用 {param} 转正则兜底（兼容参数命名差异）
func matchCode(codes []string, method, pathTemplate string) bool {
	for _, code := range codes {
		codeMethod, codePath, ok := strings.Cut(code, CodeSeparator)
		if !ok {
			continue
		}
		if codeMethod != "*" && codeMethod != method {
			continue
		}
		if codePath == pathTemplate || matchTemplate(codePath, pathTemplate) {
			return true
		}
	}
	return false
}

// matchTemplate 将权限码中的路径模板（如 /admin/v1/admin/{adminId}）转为正则
// 与请求路径模板匹配。两端均为模板（kratos PathTemplate 与 OpenAPI path 同源），
// 通常字符串相等即可命中，正则仅兜底处理 {param} 命名差异。
func matchTemplate(pattern, target string) bool {
	if !strings.Contains(pattern, "{") {
		return false
	}
	parts := strings.Split(pattern, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			parts[i] = "[^/]+"
		} else {
			parts[i] = regexp.QuoteMeta(part)
		}
	}
	re, err := regexp.Compile("^" + strings.Join(parts, "/") + "$")
	if err != nil {
		return false
	}
	return re.MatchString(target)
}
