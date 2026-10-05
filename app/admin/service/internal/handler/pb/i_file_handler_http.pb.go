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

const OperationFileUploadFile = "/admin.service.v1.File/UploadFile"
const OperationFileGetFile = "/admin.service.v1.File/GetFile"

type FileHandlerHTTPServer interface {
	UploadFile(http.Context) error
	GetFile(http.Context) error
}

func RegisterFileHandlerServer(s *http.Server, srv FileHandlerHTTPServer) {
	r := s.Route("/")
	r.POST("admin/v1/file/upload", _File_UploadFile0_HTTP_Handler(srv))
	// 文件访问统一走 query 形式：GET /admin/v1/file?objectKey=xxx
	// 说明：kratos 路由对多段 catch-all（{objectKey...}）匹配不可靠，
	// 故不使用 /admin/v1/file/{objectKey...} 路径形式。
	r.GET("admin/v1/file", _File_GetFile0_HTTP_Handler(srv))
}

func _File_UploadFile0_HTTP_Handler(srv FileHandlerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in struct{}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationFileUploadFile)

		h := ctx.Middleware(func(ctx1 context.Context, req interface{}) (interface{}, error) {
			mergeContext(ctx, ctx1)
			return nil, srv.UploadFile(ctx)
		})
		_, err := h(ctx, &in)
		return err
	}
}

func _File_GetFile0_HTTP_Handler(srv FileHandlerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in struct{}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationFileGetFile)

		h := ctx.Middleware(func(ctx1 context.Context, req interface{}) (interface{}, error) {
			mergeContext(ctx, ctx1)
			return nil, srv.GetFile(ctx)
		})
		_, err := h(ctx, &in)
		return err
	}
}

// 确保接口实现
var _ FileHandlerHTTPServer = (*handler.FileHandler)(nil)
