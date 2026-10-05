package message

import (
	"context"
	"errors"
)

// 通用错误。使用标准库 errors，避免抽象层依赖具体框架。
var (
	// ErrDriverNotFound 驱动未注册（忘记 import 对应驱动包）。
	ErrDriverNotFound = errors.New("message: driver not found")
	// ErrUnsupported 当前驱动不支持该能力（如 redis pub/sub 不支持消费组）。
	ErrUnsupported = errors.New("message: unsupported operation")
	// ErrClosed Broker 已关闭。
	ErrClosed = errors.New("message: broker closed")
	// ErrEmptyTopic topic 为空。
	ErrEmptyTopic = errors.New("message: empty topic")
	// ErrNilMessage 消息为空。
	ErrNilMessage = errors.New("message: nil message")
	// ErrNilHandler 处理函数为空。
	ErrNilHandler = errors.New("message: nil handler")
)

// Producer 消息生产者。
type Producer interface {
	// Publish 批量投递消息。返回 error 表示投递失败（调用方决定重试或降级）。
	Publish(ctx context.Context, msgs ...*Message) error
}

// Handler 消息处理函数。返回 error 视为消费失败。
type Handler func(ctx context.Context, msg *Message) error

// SubscribeOption 订阅参数。
type SubscribeOption struct {
	// Group 消费组。
	//   空字符串 → 广播语义：每个订阅者都收到一份副本（适合本地缓存刷新、SSE 推送）。
	//   非空   → 负载均衡：同组内只有一个订阅者收到（适合异步任务）。
	Group string
	// Concurrency 消费并发数，<=0 时为 1。
	Concurrency int
	// MaxRetry 处理失败后的重试次数，<=0 表示不重试。
	MaxRetry int
}

// Consumer 消息消费者。
// Subscribe 是非阻塞的：注册成功后立即返回，消息在后台 goroutine 中处理。
// 当 ctx 被取消时，订阅自动注销。
type Consumer interface {
	Subscribe(ctx context.Context, topic string, opt SubscribeOption, h Handler) error
}

// Broker 消息中间件抽象：同时具备生产与消费能力。
type Broker interface {
	Producer
	Consumer

	// Name 驱动名称，如 memory / redis / rabbitmq。
	Name() string
	// Ping 健康检查。
	Ping(ctx context.Context) error
	// Close 释放连接与后台 goroutine。
	Close() error
}

// PublishJSON 以默认 Codec 序列化 v 后投递。
func PublishJSON(ctx context.Context, p Producer, topic string, v any, opts ...Option) error {
	m, err := NewJSONMessage(topic, v, opts...)
	if err != nil {
		return err
	}
	return p.Publish(ctx, m)
}

// SubscribeJSON 订阅并自动把负载反序列化为 T 后交给 h。
func SubscribeJSON[T any](ctx context.Context, c Consumer, topic string, opt SubscribeOption, h func(context.Context, *Message, T) error) error {
	return c.Subscribe(ctx, topic, opt, func(ctx context.Context, msg *Message) error {
		var v T
		if err := msg.Decode(&v); err != nil {
			return err
		}
		return h(ctx, msg, v)
	})
}
