package pb

import (
	"context"

	"github.com/antsurge/weaver-admin/pkg/tenant"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// mergeContext 将中间件链注入的上下文合并回 http.Context 底层的 request context。
//
// 背景：标准 kratos service handler 的闭包参数名是 ctx（遮蔽外层 http.Context），
// service 收到的是中间件链传入、已注入租户的 context；而本项目的自定义 handler
// 闭包调用 srv.XXX(ctx) 时用的是外层 http.Context（wrapper），其 Value() 委托
// req.Context()，中间件链注入的值不会自动回写，导致租户作用域丢失、写操作报错。
//
// 注意：不能直接把整个 inner 链通过 req.WithContext(inner) 塞回 wrapper——
// inner 是由 ctx(wrapper) 经过中间件逐层 context.WithValue 包装而来，其祖先链
// 末端就是 wrapper 自身；一旦 req.Context() 指向 inner，再调用 wrapper.Value(key)
// 会沿 inner → ... → wrapper → req.Context() → inner 无限循环，触发
// "goroutine stack exceeds 1000000000-byte limit"。
//
// 正确做法：只提取 inner 中新增的键值（目前为租户作用域与本地化上下文），
// 挂载到包装前干净的 req.Context() 链上。
func mergeContext(ctx khttp.Context, inner context.Context) {
	reqCtx := ctx.Request().Context()
	if tid, ok := tenant.From(inner); ok {
		reqCtx = tenant.With(reqCtx, tid)
	}
	if tenant.IsUnscoped(inner) {
		reqCtx = tenant.Unscoped(reqCtx)
	}
	ctx.Reset(ctx.Response(), ctx.Request().WithContext(reqCtx))
}
