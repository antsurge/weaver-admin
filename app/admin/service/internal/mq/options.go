package mq

type Option func(*Options)

type Options struct {
	Key     string // kafka key
	Delay   int    // 延迟（秒）
	Retry   int    // 重试次数
	Durable bool   // 持久化（RabbitMQ）
}

func WithKey(key string) Option {
	return func(o *Options) {
		o.Key = key
	}
}
