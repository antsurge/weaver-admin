package mq

import (
	"context"
)

// 消息结构
type Message struct {
	ID        string
	Topic     string
	Body      []byte
	Headers   map[string]interface{}
	Timestamp int64
}

// 消费处理函数
type Handler func(ctx context.Context, msg *Message) error

// 生产者
type Producer interface {
	Publish(ctx context.Context, msg *Message, opts ...Option) error
}

// 消费者
type Consumer interface {
	Subscribe(topic string, handler Handler, opts ...Option) error
}

type MQ interface {
	Producer
	Consumer
	Close() error
}
