package rabbitmq

import (
	"context"
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	pool *ChannelPool
}

func NewPublisher(pool *ChannelPool) *Publisher {
	return &Publisher{pool: pool}
}

func (p *Publisher) Publish(ctx context.Context, exchange, key string, body []byte) error {
	ch := p.pool.Get()
	defer p.pool.Put(ch)

	if err := ch.Confirm(false); err != nil {
		return err
	}

	ackChan := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	err := ch.PublishWithContext(
		ctx,
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		return err
	}

	confirm := <-ackChan
	if !confirm.Ack {
		return errors.New("message nack")
	}

	return nil
}
