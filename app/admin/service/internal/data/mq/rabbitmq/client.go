package rabbitmq

import (
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	url  string
	conn *amqp.Connection

	mu sync.Mutex
}

func NewClient(url string) (*Client, error) {
	c := &Client{url: url}
	err := c.connect()
	if err != nil {
		return nil, err
	}
	go c.reconnect()
	return c, nil
}

func (c *Client) connect() error {
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) reconnect() {
	for {
		errChan := make(chan *amqp.Error)
		c.conn.NotifyClose(errChan)

		err := <-errChan
		log.Println("mq connection closed:", err)

		for {
			log.Println("reconnecting...")
			if err := c.connect(); err == nil {
				log.Println("reconnected success")
				break
			}
			time.Sleep(3 * time.Second)
		}
	}
}

func (c *Client) GetConn() *amqp.Connection {
	return c.conn
}
