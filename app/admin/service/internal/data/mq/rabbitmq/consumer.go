package rabbitmq

import (
	"context"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Handler func(ctx context.Context, body []byte) error

type Consumer struct {
	conn *amqp.Connection
}

const maxRetry = 3

func NewConsumer(conn *amqp.Connection) *Consumer {
	return &Consumer{conn: conn}
}

func getRetryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}
	if v, ok := headers["x-retry-count"]; ok {
		return int(v.(int32))
	}
	return 0
}

func publishToRetry(ch *amqp.Channel, msg amqp.Delivery, retry int) {
	headers := msg.Headers
	if headers == nil {
		headers = amqp.Table{}
	}
	headers["x-retry-count"] = retry

	ch.Publish(
		"",
		"test.queue.retry",
		false,
		false,
		amqp.Publishing{
			Body:    msg.Body,
			Headers: headers,
		},
	)
}

func publishToDead(ch *amqp.Channel, msg amqp.Delivery) {
	ch.Publish(
		"",
		"test.queue.dead",
		false,
		false,
		amqp.Publishing{
			Body: msg.Body,
		},
	)
}

func (c *Consumer) Consume(queue string, prefetch int, handler Handler) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}

	ch.Qos(prefetch, 0, false)

	msgs, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for msg := range msgs {
		go func(m amqp.Delivery) {
			ctx := context.Background()

			err := handler(ctx, m.Body)
			if err != nil {

				retryCount := getRetryCount(m.Headers)

				if retryCount >= maxRetry {
					log.Println("进入死信队列")
					publishToDead(ch, m)
					m.Ack(false)
					return
				}

				log.Println("进入重试队列，第", retryCount+1, "次")

				publishToRetry(ch, m, retryCount+1)
				m.Ack(false)
				return
			}

			m.Ack(false)
		}(msg)
	}

	return nil
}
