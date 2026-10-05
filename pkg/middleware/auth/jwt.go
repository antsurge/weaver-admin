package auth

import (
	"context"
	"strings"

	authenticationV1 "github.com/antsurge/weaver-admin/api/gen/go/authentication/service/v1"
	"github.com/antsurge/weaver-admin/pkg/metadata"
	"github.com/antsurge/weaver-admin/pkg/tenant"
	"github.com/antsurge/weaver-admin/pkg/utils/auth"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

const (

	// bearerWord the bearer key word for authorization
	bearerWord string = "Bearer"

	// bearerFormat authorization token format
	bearerFormat string = "Bearer %s"

	// authorizationKey holds the key used to store the JWT Token in the request tokenHeader.
	authorizationKey string = "Authorization"

	// queryTokenKey SSE 等无法携带 Header 的场景下，从 query 参数读取 token。
	queryTokenKey string = "token"

	// reason holds the error reason.
	reason string = "UNAUTHORIZED"
)

var (
	ErrTokenInvalid = authenticationV1.ErrorInvalidToken("INVALID_TOKEN")
)

func Server(opts ...auth.Option) middleware.Middleware {
	// 构建一次配置，用于读取 TokenStore 等扩展配置
	cfg := auth.NewOptions(opts...)

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, ErrTokenInvalid
			}

			authHeader := tr.RequestHeader().Get(authorizationKey)

			jwtToken, err := extractBearerToken(authHeader)
			if err != nil {
				// 回退：从 query 参数读取 token。
				// EventSource（SSE）无法设置自定义 Header，只能把 token 放在 URL 上。
				jwtToken, err = extractQueryToken(ctx)
				if err != nil {
					return nil, ErrTokenInvalid
				}
			}

			claims := &auth.BaseClaims{}
			_, err = auth.ParseToken(jwtToken, claims, opts...)
			if err != nil {
				// 不要把 jwt 的原始错误直接透出
				return nil, ErrTokenInvalid
			}
			if claims.Type != "access" {
				return nil, ErrTokenInvalid
			}

			// 校验 token 是否仍存在于存储中（登出/强制下线后立即失效）
			if cfg.TokenStore != nil {
				exists, err := cfg.TokenStore.Exists(ctx, jwtToken)
				if err != nil || !exists {
					return nil, ErrTokenInvalid
				}
			}

			// 注入到 ctx
			ctx = metadata.SetAdminID(ctx, claims.UserID)

			// 注入租户上下文：JWT 中携带租户ID（multi 模式登录时确定）；
			// 旧 token/无租户时回退默认租户（单租户模式下所有数据均在 default 租户）。
			tid := claims.TenantID
			if tid == "" {
				tid = tenant.DefaultTenantID
			}
			ctx = tenant.With(ctx, tid)

			return handler(ctx, req)
		}
	}
}

// extractQueryToken 从 HTTP query 参数 ?token= 中提取 JWT（SSE 场景专用）。
// 兼容 "Bearer xxx" 与裸 token 两种写法。
func extractQueryToken(ctx context.Context) (string, error) {
	req, ok := khttp.RequestFromServerContext(ctx)
	if !ok || req == nil || req.URL == nil {
		return "", ErrTokenInvalid
	}
	return extractToken(req.URL.Query().Get(queryTokenKey), false)
}

// extractBearerToken 从 Authorization 头提取，严格要求 "Bearer <token>" 格式。
func extractBearerToken(authHeader string) (string, error) {
	return extractToken(authHeader, true)
}

// extractToken 提取并校验 JWT。
// requireBearer=true 时必须是 "Bearer <token>" 格式；false 时裸 token 也接受。
func extractToken(v string, requireBearer bool) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", ErrTokenInvalid
	}

	if parts := strings.SplitN(v, " ", 2); len(parts) == 2 {
		if !strings.EqualFold(parts[0], bearerWord) {
			return "", ErrTokenInvalid
		}
		v = strings.TrimSpace(parts[1])
	} else if requireBearer {
		return "", ErrTokenInvalid
	}

	if v == "" || strings.Count(v, ".") != 2 {
		// JWT 必须是三段
		return "", ErrTokenInvalid
	}
	return v, nil
}
