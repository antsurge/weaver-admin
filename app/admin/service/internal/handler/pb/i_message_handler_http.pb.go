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

const OperationMessageStream = "/admin.service.v1.Message/Stream"

type MessageHandlerHTTPServer interface {
	StreamNotification(http.Context) error
}

// RegisterMessageHandlerServer 注册消息通知的自定义 HTTP 路由。
// 注意：字面量路径（.../stream）必须先于参数化路由注册，否则会被 {id} 抢先匹配。
func RegisterMessageHandlerServer(s *http.Server, srv MessageHandlerHTTPServer) {
	r := s.Route("/")
	r.GET("admin/v1/notifications/stream", _Message_Stream0_HTTP_Handler(srv))
}

func _Message_Stream0_HTTP_Handler(srv MessageHandlerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in struct{}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationMessageStream)

		h := ctx.Middleware(func(ctx1 context.Context, req interface{}) (interface{}, error) {
			mergeContext(ctx, ctx1)
			return nil, srv.StreamNotification(ctx)
		})
		_, err := h(ctx, &in)
		return err
	}
}

// 确保接口实现
var _ MessageHandlerHTTPServer = (*handler.NotificationHandler)(nil)
