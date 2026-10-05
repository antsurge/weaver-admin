package server

import (
	"context"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
)

// NewMQServer 把消息消费封装成 Kratos 服务器，由框架统一管理启停。
//
// 订阅使用「广播语义」（Group 为空）：每个实例都会收到事件，
// 再各自查本地 SSE 注册表投递给本实例在线的连接。
func NewMQServer(uc *biz.NotificationUsecase, logger log.Logger) *mqServer {
	return &mqServer{uc: uc, log: log.NewHelper(logger)}
}

type mqServer struct {
	uc     *biz.NotificationUsecase
	log    *log.Helper
	cancel context.CancelFunc
}

func (s *mqServer) Start(_ context.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	if err := s.uc.StartConsumer(ctx); err != nil {
		cancel()
		return err
	}
	s.cancel = cancel
	s.log.Info("notification consumer started")
	return nil
}

func (s *mqServer) Stop(_ context.Context) error {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.log.Info("notification consumer stopped")
	return nil
}
