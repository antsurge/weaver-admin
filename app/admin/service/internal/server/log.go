package server

import (
	"context"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/openapi_scanner"
	"github.com/antsurge/weaver-admin/pkg/middleware/oplog"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
)

// 不采集的操作：高频且无审计价值，或本身就是长连接
var skipOperations = []string{
	"/admin.service.v1.AuthenticationService/GetCaptcha",
	"/admin.service.v1.Message/Stream",
}

// 不采集的 HTTP 路径前缀
var skipPathPrefixes = []string{
	"/admin/v1/notifications/stream",
}

// oplogRecorder 把中间件采集的 Entry 转成 biz 模型交给异步写入器。
type oplogRecorder struct {
	rec *biz.OperationLogRecorder
}

func (a oplogRecorder) Record(e *oplog.Entry) {
	if e == nil {
		return
	}
	a.rec.Record(&biz.OperationLog{
		TenantID:  e.TenantID,
		AdminID:   e.AdminID,
		AdminName: e.AdminName,
		Module:    e.Module,
		Operation: e.Operation,
		Method:    e.Method,
		Path:      e.Path,
		Summary:   e.Summary,
		IP:        e.IP,
		UserAgent: e.UserAgent,
		Params:    e.Params,
		Status:    e.Status,
		Code:      e.Code,
		Message:   e.Message,
		CostMS:    e.CostMS,
	})
}

// NewOperationLogMiddleware 构建操作日志采集中间件。
// recorder 为 nil 时返回空中间件（不影响服务启动）。
func NewOperationLogMiddleware(
	recorder *biz.OperationLogRecorder,
	scanner *openapi_scanner.Service,
) middleware.Middleware {
	if recorder == nil {
		return func(next middleware.Handler) middleware.Handler { return next }
	}

	opts := []oplog.Option{
		oplog.WithSkipOperations(skipOperations...),
		oplog.WithSkipPathPrefix(skipPathPrefixes...),
	}
	if scanner != nil {
		opts = append(opts, oplog.WithSummaryResolver(scanner.SummaryByEndpoint))
	}
	return oplog.Server(oplogRecorder{rec: recorder}, opts...)
}

// NewLogRecorderServer 把操作日志异步写入器托管给框架，随服务启停。
func NewLogRecorderServer(recorder *biz.OperationLogRecorder, logger log.Logger) *logRecorderServer {
	return &logRecorderServer{recorder: recorder, log: log.NewHelper(logger)}
}

type logRecorderServer struct {
	recorder *biz.OperationLogRecorder
	log      *log.Helper
}

func (s *logRecorderServer) Start(_ context.Context) error {
	if s.recorder == nil {
		return nil
	}
	s.recorder.Start()
	s.log.Info("operation log recorder started")
	return nil
}

func (s *logRecorderServer) Stop(_ context.Context) error {
	if s.recorder == nil {
		return nil
	}
	// Stop 内部会把缓冲中剩余的日志刷入库
	s.recorder.Stop()
	s.log.Info("operation log recorder stopped")
	return nil
}
