package pb

import (
	"context"

	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-kratos/kratos/v2/transport/http/binding"
)

var _ = new(context.Context)
var _ = binding.EncodeURL

const _ = http.SupportPackageIsVersion1

// OrderHandlerHTTPServer 订单导入导出 HTTP 处理器接口
type OrderHandlerHTTPServer interface {
	ExportOrder(http.Context) error
	ImportOrder(http.Context) error
}

const (
	OperationOrderExport = "/order.service.v1.Order/Export"
	OperationOrderImport = "/order.service.v1.Order/Import"
)

// RegisterOrderHandlerServer 注册订单导入导出路由。
// 注意：带字面量路径的路由（/import、/export）必须先于参数化路由（/{id}）注册。
func RegisterOrderHandlerServer(s *http.Server, srv OrderHandlerHTTPServer) {
	r := s.Route("/")
	r.POST("admin/v1/orders/export", _Order_Export0_HTTP_Handler(srv))
	r.POST("admin/v1/orders/import", _Order_Import0_HTTP_Handler(srv))
}

func _Order_Export0_HTTP_Handler(srv OrderHandlerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		h := ctx.Middleware(func(ctx1 context.Context, req interface{}) (interface{}, error) {
			mergeContext(ctx, ctx1)
			return nil, srv.ExportOrder(ctx)
		})
		http.SetOperation(ctx, OperationOrderExport)
		_, err := h(ctx, nil)
		return err
	}
}

func _Order_Import0_HTTP_Handler(srv OrderHandlerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		h := ctx.Middleware(func(ctx1 context.Context, req interface{}) (interface{}, error) {
			mergeContext(ctx, ctx1)
			return nil, srv.ImportOrder(ctx)
		})
		http.SetOperation(ctx, OperationOrderImport)
		_, err := h(ctx, nil)
		return err
	}
}
