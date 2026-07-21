package rabbitmq

import amqp "github.com/rabbitmq/amqp091-go"

// 声明主队列 + 死信
func DeclareQueue(ch *amqp.Channel, cfg Config) error {
	// 死信交换机
	err := ch.ExchangeDeclare(
		cfg.Exchange+".dlx",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// 主队列（绑定死信）
	args := amqp.Table{
		"x-dead-letter-exchange": cfg.Exchange + ".dlx",
	}

	_, err = ch.QueueDeclare(
		cfg.Queue,
		true,
		false,
		false,
		false,
		args,
	)
	if err != nil {
		return err
	}

	return nil
}

func DeclareRetryQueue(ch *amqp.Channel, cfg Config, ttl int) error {
	args := amqp.Table{
		"x-message-ttl":             ttl,
		"x-dead-letter-exchange":    cfg.Exchange,
		"x-dead-letter-routing-key": cfg.RoutingKey,
	}

	_, err := ch.QueueDeclare(
		cfg.Queue+".retry",
		true,
		false,
		false,
		false,
		args,
	)
	return err
}

func DeclareDeadQueue(ch *amqp.Channel, cfg Config) error {
	_, err := ch.QueueDeclare(
		cfg.Queue+".dead",
		true,
		false,
		false,
		false,
		nil,
	)
	return err
}

// 初始化订单延迟队列
func DeclareOrderDelay(ch *amqp.Channel) error {
	// 1. 声明主交换机（取消订单用）
	err := ch.ExchangeDeclare(
		"order.exchange",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// 2. 声明取消队列
	_, err = ch.QueueDeclare(
		"order.cancel.queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// 3. 绑定取消队列
	err = ch.QueueBind(
		"order.cancel.queue",
		"order.cancel",
		"order.exchange",
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// 4. 声明延迟队列（核心🔥）
	args := amqp.Table{
		"x-message-ttl":             int32(30 * 60 * 1000), // 30分钟
		"x-dead-letter-exchange":    "order.exchange",
		"x-dead-letter-routing-key": "order.cancel",
	}

	_, err = ch.QueueDeclare(
		"order.delay.queue",
		true,
		false,
		false,
		false,
		args,
	)

	return err
}
