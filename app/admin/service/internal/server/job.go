package server

import (
	"context"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
)

// NewJobSchedulerServer 把定时任务调度器封装成 Kratos 服务器，
// 由框架统一管理启停：启动时加载所有启用任务并开始调度。
func NewJobSchedulerServer(scheduler *biz.JobScheduler, logger log.Logger) *jobSchedulerServer {
	// 注入日志回调
	scheduler.SetLogDebug(func(format string, args ...interface{}) {
		log.NewHelper(logger).Debugf("[job-scheduler] "+format, args...)
	})
	return &jobSchedulerServer{scheduler: scheduler, log: log.NewHelper(logger)}
}

type jobSchedulerServer struct {
	scheduler *biz.JobScheduler
	log       *log.Helper
}

func (s *jobSchedulerServer) Start(ctx context.Context) error {
	if err := s.scheduler.Start(ctx); err != nil {
		return err
	}
	s.log.Info("job scheduler started")
	return nil
}

func (s *jobSchedulerServer) Stop(ctx context.Context) error {
	if err := s.scheduler.Stop(ctx); err != nil {
		return err
	}
	s.log.Info("job scheduler stopped")
	return nil
}
