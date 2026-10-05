package pb

import (
	"context"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/handler"
	"github.com/go-kratos/kratos/v2/transport/http"
)

var _ = new(context.Context)

const _ = http.SupportPackageIsVersion1

const OperationProfileUpdateCurrentUser = "/admin.service.v1.Profile/UpdateCurrentUser"
const OperationProfileUpdateCurrentUserPassword = "/admin.service.v1.Profile/UpdateCurrentUserPassword"

type ProfileHandlerHTTPServer interface {
	UpdateCurrentUser(http.Context) error
	UpdateCurrentUserPassword(http.Context) error
}

func RegisterProfileHandlerServer(s *http.Server, srv ProfileHandlerHTTPServer) {
	r := s.Route("/")
	r.PUT("admin/v1/current-user", _Profile_UpdateCurrentUser0_HTTP_Handler(srv))
	r.PUT("admin/v1/current-user/password", _Profile_UpdateCurrentUserPassword0_HTTP_Handler(srv))
}

func _Profile_UpdateCurrentUser0_HTTP_Handler(srv ProfileHandlerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in struct{}
		http.SetOperation(ctx, OperationProfileUpdateCurrentUser)
		h := ctx.Middleware(func(ctx1 context.Context, req interface{}) (interface{}, error) {
			mergeContext(ctx, ctx1)
			return nil, srv.UpdateCurrentUser(ctx)
		})
		_, err := h(ctx, &in)
		return err
	}
}

func _Profile_UpdateCurrentUserPassword0_HTTP_Handler(srv ProfileHandlerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in struct{}
		http.SetOperation(ctx, OperationProfileUpdateCurrentUserPassword)
		h := ctx.Middleware(func(ctx1 context.Context, req interface{}) (interface{}, error) {
			mergeContext(ctx, ctx1)
			return nil, srv.UpdateCurrentUserPassword(ctx)
		})
		_, err := h(ctx, &in)
		return err
	}
}

// 确保接口实现
var _ ProfileHandlerHTTPServer = (*handler.ProfileHandler)(nil)
