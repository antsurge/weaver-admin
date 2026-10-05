// Package redis 基于 Redis Pub/Sub 的消息实现。
//
// 语义说明：
//   - Group == ""  → 广播：每个实例都会收到（Redis Pub/Sub 天然语义）
//   - Group != ""  → 返回 message.ErrUnsupported：Pub/Sub 没有消费组概念，
//     需要负载均衡语义时请改用 redis-stream / rabbitmq / kafka 驱动。
//
// 适用：站内信实时推送这类「允许偶发丢失」的加速通道——真相源在数据库，
// 丢一条只是晚几秒，用户刷新仍能拿到。
package redis

import (
	"context"
	"sync"
	"time"

	"github.com/antsurge/weaver-admin/pkg/utils/message"
	goredis "github.com/redis/go-redis/v9"
)

// DriverName 驱动名。
const DriverName = "redis"

// ChannelSize go-redis PubSub 内部 channel 缓冲大小。
const ChannelSize = 256

func init() {
	message.Register(DriverName, func(c *message.Config) (message.Broker, error) {
		return New(c)
	})
}

// Broker 基于 Redis Pub/Sub 的实现。
type Broker struct {
	client *goredis.Client
	prefix string
	owned  bool // client 是否由本 Broker 创建（决定 Close 时是否关闭）

	mu     sync.Mutex
	subs   []*goredis.PubSub
	closed bool
}

// New 根据配置创建 Broker（使用独立的 Redis 连接，避免 Pub/Sub 占用缓存连接）。
func New(c *message.Config) (*Broker, error) {
	if c == nil {
		c = &message.Config{}
	}
	addr := c.Addr()
	if addr == "" {
		addr = "127.0.0.1:6379"
	}

	opt := &goredis.Options{
		Network:      c.Option("network", "tcp"),
		Addr:         addr,
		Username:     c.Username,
		Password:     c.Password,
		DB:           c.OptionInt("db", 0),
		PoolSize:     c.OptionInt("pool_size", 10),
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	client := goredis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return NewWithClient(client, c.Prefix, true), nil
}

// NewWithClient 使用已有的 Redis 客户端创建 Broker（owned=false 时 Close 不关闭客户端）。
func NewWithClient(client *goredis.Client, prefix string, owned bool) *Broker {
	return &Broker{client: client, prefix: prefix, owned: owned}
}

func (b *Broker) Name() string { return DriverName }

func (b *Broker) Ping(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return b.client.Ping(ctx).Err()
}

// Publish 批量投递。Redis Pub/Sub 不支持延迟投递，Delay > 0 会被忽略。
func (b *Broker) Publish(ctx context.Context, msgs ...*message.Message) error {
	if ctx == nil {
		ctx = context.Background()
	}
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return message.ErrClosed
	}
	b.mu.Unlock()

	for _, m := range msgs {
		if err := message.Normalize(m); err != nil {
			return err
		}
		payload, err := m.Marshal()
		if err != nil {
			return err
		}
		if err := b.client.Publish(ctx, message.WithPrefix(b.prefix, m.Topic), payload).Err(); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe 订阅主题。ctx 取消时自动退订。
func (b *Broker) Subscribe(ctx context.Context, topic string, opt message.SubscribeOption, h message.Handler) error {
	if topic == "" {
		return message.ErrEmptyTopic
	}
	if h == nil {
		return message.ErrNilHandler
	}
	if ctx == nil {
		ctx = context.Background()
	}
	// Pub/Sub 是纯广播，无法表达消费组语义。
	if opt.Group != "" {
		return message.ErrUnsupported
	}

	handler := message.Chain(h,
		message.RecoverMW(nil),
		message.RetryMW(opt.MaxRetry, nil),
	)

	channel := message.WithPrefix(b.prefix, topic)
	pubsub := b.client.Subscribe(ctx, channel)

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		_ = pubsub.Close()
		return message.ErrClosed
	}
	b.subs = append(b.subs, pubsub)
	b.mu.Unlock()

	ch := pubsub.Channel(goredis.WithChannelSize(ChannelSize))

	conc := opt.Concurrency
	if conc <= 0 {
		conc = 1
	}
	var wg sync.WaitGroup
	for i := 0; i < conc; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for raw := range ch {
				m, err := message.Unmarshal([]byte(raw.Payload))
				if err != nil {
					continue
				}
				_ = handler(ctx, m)
			}
		}()
	}

	// ctx 结束时退订
	go func() {
		<-ctx.Done()
		_ = pubsub.Close()
		wg.Wait()
		b.remove(pubsub)
	}()

	return nil
}

func (b *Broker) remove(pubsub *goredis.PubSub) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, s := range b.subs {
		if s == pubsub {
			b.subs = append(b.subs[:i], b.subs[i+1:]...)
			return
		}
	}
}

// Client 返回底层 Redis 客户端（便于复用做健康检查等）。
func (b *Broker) Client() *goredis.Client { return b.client }

func (b *Broker) Close() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	subs := b.subs
	b.subs = nil
	b.mu.Unlock()

	for _, s := range subs {
		_ = s.Close()
	}
	if b.owned && b.client != nil {
		return b.client.Close()
	}
	return nil
}
