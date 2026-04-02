package biz

import "context"

type MQ interface {
	Publish(ctx context.Context, topic string, data []byte) error
	Subscribe(topic string, handler func(context.Context, []byte) error) error
}
