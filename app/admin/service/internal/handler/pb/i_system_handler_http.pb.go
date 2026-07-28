package pb

import (
	"context"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/handler"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-kratos/kratos/v2/transport/http/binding"
)

var _ = new(context.Context)
var _ = binding.EncodeURL

const _ = http.SupportPackageIsVersion1

const OperationSystemImportApiInterface = "/admin.service.v1.System/ImportApiInterface"

type SystemHandlerHTTPServer interface {
	ImportApiInterface(http.Context) error
}

func RegisterSystemHandlerServer(s *http.Server, srv SystemHandlerHTTPServer) {
	r := s.Route("/")
	r.POST("admin/v1/api-interface/import", _System_ImportApiInterface0_HTTP_Handler(srv))
}

func _System_ImportApiInterface0_HTTP_Handler(srv SystemHandlerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in struct{}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationSystemImportApiInterface)

		h := ctx.Middleware(func(ctx1 context.Context, req interface{}) (interface{}, error) {
			return nil, srv.ImportApiInterface(ctx)
		})
		_, err := h(ctx, &in)
		return err
	}
}

// 确保接口实现
var _ SystemHandlerHTTPServer = (*handler.SystemHandler)(nil)