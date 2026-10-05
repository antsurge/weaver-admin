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

const OperationAdminExport = "/admin.service.v1.Admin/Export"

type AdminHandlerHTTPServer interface {
	ExportAdmin(http.Context) error
}

func RegisterAdminHandlerServer(s *http.Server, srv AdminHandlerHTTPServer) {
	r := s.Route("/")
	r.GET("admin/v1/admin/export", _Admin_Export0_HTTP_Handler(srv))
}

func _Admin_Export0_HTTP_Handler(srv AdminHandlerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in struct{}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationAdminExport)

		h := ctx.Middleware(func(ctx1 context.Context, req interface{}) (interface{}, error) {
			mergeContext(ctx, ctx1)
			return nil, srv.ExportAdmin(ctx)
		})
		_, err := h(ctx, &in)
		return err
	}
}

// 确保接口实现
var _ AdminHandlerHTTPServer = (*handler.AdminHandler)(nil)
