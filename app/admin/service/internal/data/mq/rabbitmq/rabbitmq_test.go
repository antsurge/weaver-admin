package rabbitmq

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"testing"
	"time"
)

func TestRabbitMQ_Flow(t *testing.T) {
	cfg := Config{
		URL:        "amqp://guest:guest@localhost:5672/",
		Exchange:   "test.exchange",
		Queue:      "test.queue",
		RoutingKey: "test.key",
		Prefetch:   1,
		PoolSize:   2,
	}

	// 1. 初始化 client
	client, err := NewClient(cfg.URL)
	if err != nil {
		t.Fatal(err)
	}

	conn := client.GetConn()

	// 2. 创建 channel
	ch, err := conn.Channel()
	if err != nil {
		t.Fatal(err)
	}

	// 3. 声明基础设施（非常关键）
	err = ch.ExchangeDeclare(cfg.Exchange, "direct", true, false, false, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = DeclareQueue(ch, cfg)
	if err != nil {
		t.Fatal(err)
	}

	err = DeclareRetryQueue(ch, cfg, 2000) // 2秒重试
	if err != nil {
		t.Fatal(err)
	}

	err = DeclareDeadQueue(ch, cfg)
	if err != nil {
		t.Fatal(err)
	}

	// 绑定
	err = ch.QueueBind(cfg.Queue, cfg.RoutingKey, cfg.Exchange, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	// retry 队列绑定
	//err = ch.QueueBind(cfg.Queue+".retry", "", "", false, nil)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//// dead 队列绑定
	//err = ch.QueueBind(cfg.Queue+".dead", "", "", false, nil)
	//if err != nil {
	//	t.Fatal(err)
	//}

	// 4. 初始化 publisher
	pool, err := NewChannelPool(conn, cfg.PoolSize)
	if err != nil {
		t.Fatal(err)
	}
	pub := NewPublisher(pool)

	// 5. 初始化 consumer
	consumer := NewConsumer(conn)

	var successCount int32
	var retryCount int32

	// 启动消费
	go func() {
		err := consumer.Consume(cfg.Queue, cfg.Prefetch, func(ctx context.Context, body []byte) error {
			msg := string(body)
			log.Println("消费:", msg)

			if msg == "fail" {
				atomic.AddInt32(&retryCount, 1)
				return errors.New("模拟失败")
			}

			atomic.AddInt32(&successCount, 1)
			return nil
		})
		if err != nil {
			t.Error(err)
		}
	}()

	time.Sleep(1 * time.Second)

	// 6. 发送测试消息
	err = pub.Publish(context.Background(), cfg.Exchange, cfg.RoutingKey, []byte("ok"))
	if err != nil {
		t.Fatal(err)
	}

	err = pub.Publish(context.Background(), cfg.Exchange, cfg.RoutingKey, []byte("fail"))
	if err != nil {
		t.Fatal(err)
	}

	// 等待重试完成
	time.Sleep(10 * time.Second)

	// 7. 断言
	if successCount == 0 {
		t.Fatal("正常消息未被消费")
	}

	if retryCount == 0 {
		t.Fatal("失败消息未触发重试")
	}

	log.Println("成功消费:", successCount)
	log.Println("重试次数:", retryCount)
}
