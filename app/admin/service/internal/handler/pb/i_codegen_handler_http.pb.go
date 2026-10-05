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

const OperationCodegenDownload = "/admin.service.v1.CodeGen/Download"
const OperationCodegenApplyMenu = "/admin.service.v1.CodeGen/ApplyMenu"
const OperationCodegenDeleteGenerated = "/admin.service.v1.CodeGen/DeleteGenerated"

type CodegenHandlerHTTPServer interface {
	DownloadCode(http.Context) error
	ApplyMenu(http.Context) error
	DeleteGenerated(http.Context) error
}

func RegisterCodegenHandlerServer(s *http.Server, srv CodegenHandlerHTTPServer) {
	r := s.Route("/")
	r.POST("admin/v1/codegen:download", _Codegen_Download0_HTTP_Handler(srv))
	r.POST("admin/v1/codegen/apply-menu", _Codegen_ApplyMenu0_HTTP_Handler(srv))
	r.POST("admin/v1/codegen/delete-generated", _Codegen_DeleteGenerated0_HTTP_Handler(srv))
}

func _Codegen_Download0_HTTP_Handler(srv CodegenHandlerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in struct{}
		_ = in
		http.SetOperation(ctx, OperationCodegenDownload)

		h := ctx.Middleware(func(ctx1 context.Context, req interface{}) (interface{}, error) {
			mergeContext(ctx, ctx1)
			return nil, srv.DownloadCode(ctx)
		})
		_, err := h(ctx, &in)
		return err
	}
}

func _Codegen_ApplyMenu0_HTTP_Handler(srv CodegenHandlerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in struct{}
		_ = in
		http.SetOperation(ctx, OperationCodegenApplyMenu)

		h := ctx.Middleware(func(ctx1 context.Context, req interface{}) (interface{}, error) {
			mergeContext(ctx, ctx1)
			return nil, srv.ApplyMenu(ctx)
		})
		_, err := h(ctx, &in)
		return err
	}
}

func _Codegen_DeleteGenerated0_HTTP_Handler(srv CodegenHandlerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in struct{}
		_ = in
		http.SetOperation(ctx, OperationCodegenDeleteGenerated)

		h := ctx.Middleware(func(ctx1 context.Context, req interface{}) (interface{}, error) {
			mergeContext(ctx, ctx1)
			return nil, srv.DeleteGenerated(ctx)
		})
		_, err := h(ctx, &in)
		return err
	}
}

var _ CodegenHandlerHTTPServer = (*handler.CodegenHandler)(nil)
