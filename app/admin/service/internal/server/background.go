package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/transport"
)

// NewBackgroundServer 汇总所有非网络服务（消息消费、日志写入、任务调度等），
// 交给 Kratos 统一启停。单独一个 transport.Server 可避免 Wire 上
// 出现多个同类型提供者导致注入歧义。
func NewBackgroundServer(mq *mqServer, logs *logRecorderServer, jobs *jobSchedulerServer) transport.Server {
	return &backgroundServer{servers: []transport.Server{mq, logs, jobs}}
}

type backgroundServer struct {
	servers []transport.Server
}

func (s *backgroundServer) Start(ctx context.Context) error {
	for _, srv := range s.servers {
		if err := srv.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *backgroundServer) Stop(ctx context.Context) error {
	// 逆序停止，保证先停止入口再停依赖
	for i := len(s.servers) - 1; i >= 0; i-- {
		if err := s.servers[i].Stop(ctx); err != nil {
			return err
		}
	}
	return nil
}
