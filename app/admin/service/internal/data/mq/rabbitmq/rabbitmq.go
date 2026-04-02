package rabbitmq

import "context"

func (r *RabbitMQ) Publish(ctx context.Context, topic string, data []byte) error {
	return r.publisher.Publish(ctx, "main.exchange", topic, data)
}

func (r *RabbitMQ) Subscribe(topic string, handler func(context.Context, []byte) error) error {
	go func() {
		err := r.consumer.Consume(topic, 5, handler)
		if err != nil {
			panic(err)
		}
	}()
	return nil
}
