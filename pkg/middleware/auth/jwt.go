package auth

import (
	"context"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	authenticationV1 "github.com/hypercoze/kratos-admin/api/gen/go/authentication/service/v1"
	"github.com/hypercoze/kratos-admin/pkg/metadata"
	"github.com/hypercoze/kratos-admin/pkg/utils/auth"
	"strings"
)

const (

	// bearerWord the bearer key word for authorization
	bearerWord string = "Bearer"

	// bearerFormat authorization token format
	bearerFormat string = "Bearer %s"

	// authorizationKey holds the key used to store the JWT Token in the request tokenHeader.
	authorizationKey string = "Authorization"

	// reason holds the error reason.
	reason string = "UNAUTHORIZED"
)

var (
	ErrTokenInvalid = authenticationV1.ErrorInvalidToken("INVALID_TOKEN")
)

func Server(opts ...auth.Option) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, ErrTokenInvalid
			}

			authHeader := tr.RequestHeader().Get(authorizationKey)

			jwtToken, err := extractBearerToken(authHeader)
			if err != nil {
				return nil, ErrTokenInvalid
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

			// 注入到 ctx
			ctx = metadata.SetAdminID(ctx, claims.UserID)

			return handler(ctx, req)
		}
	}
}

func extractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", ErrTokenInvalid
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return "", ErrTokenInvalid
	}

	if !strings.EqualFold(parts[0], bearerWord) {
		return "", ErrTokenInvalid
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", ErrTokenInvalid
	}

	// JWT 必须是三段
	if strings.Count(token, ".") != 2 {
		return "", ErrTokenInvalid
	}

	return token, nil
}
