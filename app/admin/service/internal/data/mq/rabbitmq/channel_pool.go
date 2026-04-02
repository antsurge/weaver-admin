package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type ChannelPool struct {
	conn     *amqp.Connection
	channels chan *amqp.Channel
}

func NewChannelPool(conn *amqp.Connection, size int) (*ChannelPool, error) {
	pool := &ChannelPool{
		conn:     conn,
		channels: make(chan *amqp.Channel, size),
	}

	for i := 0; i < size; i++ {
		ch, err := conn.Channel()
		if err != nil {
			return nil, err
		}
		pool.channels <- ch
	}
	return pool, nil
}

func (p *ChannelPool) Get() *amqp.Channel {
	return <-p.channels
}

func (p *ChannelPool) Put(ch *amqp.Channel) {
	p.channels <- ch
}
