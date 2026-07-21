package rabbitmq

import (
	"github.com/antsurge/weaver-admin/app/admin/service/internal/conf"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewRabbitMQ,
)

type RabbitMQ struct {
	client    *Client
	publisher *Publisher
	consumer  *Consumer
}

func NewRabbitMQ(c *conf.MQ) (*RabbitMQ, error) {
	conf := c.GetRabbitmq()

	client, err := NewClient(conf.Url)
	if err != nil {
		return nil, err
	}

	//ch, _ := client.GetConn().Channel()
	//err = DeclareOrderDelay(ch)
	//if err != nil {
	//	panic(err)
	//}

	pool, err := NewChannelPool(client.GetConn(), int(conf.PoolSize))
	if err != nil {
		return nil, err
	}

	return &RabbitMQ{
		client:    client,
		publisher: NewPublisher(pool),
		consumer:  NewConsumer(client.GetConn()),
	}, nil
}
