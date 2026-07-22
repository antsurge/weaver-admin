package demo

import (
	"context"
	"net/http"

	authenticationV1 "github.com/antsurge/weaver-admin/api/gen/go/authentication/service/v1"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

var (
	ErrDemoReadonly = authenticationV1.ErrorDemoReadonly("DEMO_READONLY")
)

func DemoReadonly(enabled bool) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {

		return func(ctx context.Context, req interface{}) (interface{}, error) {

			if !enabled {
				return handler(ctx, req)
			}

			// 获取 HTTP Transport
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return handler(ctx, req)
			}

			if tr.Kind() != transport.KindHTTP {
				return handler(ctx, req)
			}

			// 断言成HTTP的Transport可以拿到特殊信息
			if ht, ok := tr.(*khttp.Transport); ok {
				reqMethod := ht.Request().Method
				if reqMethod != http.MethodGet {
					return nil, errors.BadRequest("DemoReadonly", ErrDemoReadonly.Message)
				}
			}

			return handler(ctx, req)
		}
	}
}
