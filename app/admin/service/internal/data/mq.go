package data

import (
	"github.com/antsurge/weaver-admin/app/admin/service/internal/conf"
	"github.com/antsurge/weaver-admin/pkg/utils/message"
	// 注册可用驱动：按需切换 conf.mq.driver 即可，业务代码无需改动
	_ "github.com/antsurge/weaver-admin/pkg/utils/message/memory"
	_ "github.com/antsurge/weaver-admin/pkg/utils/message/redis"
	"github.com/go-kratos/kratos/v2/log"
)

// NewMessageBroker 按配置创建消息中间件实例。
//
// 降级策略：enabled=false 或 driver=noop 时返回内置空实现，消息被丢弃但服务照常启动
// （通知仍会落库，只是不做实时推送）。
func NewMessageBroker(c *conf.MQ, logger log.Logger) (message.Broker, func(), error) {
	l := log.NewHelper(logger)

	cfg := &message.Config{}
	if c != nil {
		cfg.Enabled = c.GetEnabled()
		cfg.Driver = c.GetDriver()
		cfg.Addrs = c.GetAddrs()
		cfg.Username = c.GetUsername()
		cfg.Password = c.GetPassword()
		cfg.VHost = c.GetVhost()
		cfg.Prefix = c.GetPrefix()
		cfg.Options = c.GetOptions()
	}

	broker, err := message.New(cfg)
	if err != nil {
		return nil, nil, err
	}

	l.Infof("message broker initialized: %s", cfg.String())
	cleanup := func() {
		if err := broker.Close(); err != nil {
			l.Errorf("closing message broker: %v", err)
		}
	}
	return broker, cleanup, nil
}
