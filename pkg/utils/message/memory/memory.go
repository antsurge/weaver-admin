// Package memory 提供基于进程内 channel 的消息实现。
//
// 适用场景：本地开发、单副本部署、单元测试。零外部依赖。
// 局限：消息不跨进程、不持久化，进程退出即丢失。
package memory

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/antsurge/weaver-admin/pkg/utils/message"
)

// DriverName 驱动名。
const DriverName = "memory"

// defaultBufferSize 每个订阅者的消息缓冲大小，满了则丢弃（推送类消息允许丢失）。
const defaultBufferSize = 256

func init() {
	message.Register(DriverName, func(c *message.Config) (message.Broker, error) {
		return New(c)
	})
}

// Broker 进程内消息实现。
type Broker struct {
	mu     sync.RWMutex
	topics map[string]*topic
	closed bool

	delayWG sync.WaitGroup
	dropped int64
}

// New 创建 memory Broker。
func New(_ *message.Config) (*Broker, error) {
	return &Broker{topics: make(map[string]*topic)}, nil
}

func (b *Broker) Name() string { return DriverName }

func (b *Broker) Ping(_ context.Context) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.closed {
		return message.ErrClosed
	}
	return nil
}

// Publish 投递消息。支持 Delay（通过定时器延迟投递）。
func (b *Broker) Publish(ctx context.Context, msgs ...*message.Message) error {
	for _, m := range msgs {
		if err := message.Normalize(m); err != nil {
			return err
		}
	}

	b.mu.RLock()
	closed := b.closed
	b.mu.RUnlock()
	if closed {
		return message.ErrClosed
	}

	for _, m := range msgs {
		if m.Delay > 0 {
			delayed := m
			b.delayWG.Add(1)
			time.AfterFunc(m.Delay, func() {
				defer b.delayWG.Done()
				_ = b.deliver(delayed)
			})
			continue
		}
		if err := b.deliver(m); err != nil {
			return err
		}
	}
	return nil
}

func (b *Broker) deliver(m *message.Message) error {
	b.mu.RLock()
	t := b.topics[m.Topic]
	closed := b.closed
	b.mu.RUnlock()
	if closed {
		return message.ErrClosed
	}
	if t == nil {
		return nil
	}
	t.deliver(m)
	return nil
}

// Subscribe 注册订阅。ctx 取消时自动注销并停止消费 goroutine。
func (b *Broker) Subscribe(ctx context.Context, topicName string, opt message.SubscribeOption, h message.Handler) error {
	if topicName == "" {
		return message.ErrEmptyTopic
	}
	if h == nil {
		return message.ErrNilHandler
	}
	if ctx == nil {
		ctx = context.Background()
	}

	handler := message.Chain(h,
		message.RecoverMW(nil),
		message.RetryMW(opt.MaxRetry, nil),
	)

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return message.ErrClosed
	}
	t := b.topics[topicName]
	if t == nil {
		t = &topic{}
		b.topics[topicName] = t
	}
	sub := newSubscriber(ctx, opt.Group, b)
	t.add(sub)
	b.mu.Unlock()

	conc := opt.Concurrency
	if conc <= 0 {
		conc = 1
	}
	for i := 0; i < conc; i++ {
		sub.wg.Add(1)
		go sub.run(handler)
	}

	// ctx 结束时注销订阅
	go func() {
		<-ctx.Done()
		t.remove(sub)
	}()
	return nil
}

// Dropped 返回因缓冲满而被丢弃的消息数（用于观测）。
func (b *Broker) Dropped() int64 { return atomic.LoadInt64(&b.dropped) }

func (b *Broker) Close() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	subs := make([]*subscriber, 0)
	for _, t := range b.topics {
		subs = append(subs, t.drain()...)
	}
	b.topics = make(map[string]*topic)
	b.mu.Unlock()

	for _, s := range subs {
		s.close()
	}
	return nil
}

// topic 单个主题下的订阅者集合。
type topic struct {
	mu   sync.RWMutex
	subs []*subscriber
	// rr 用于同 Group 内轮询分发
	rr map[string]int
}

func (t *topic) add(s *subscriber) {
	t.mu.Lock()
	t.subs = append(t.subs, s)
	if s.group != "" && t.rr == nil {
		t.rr = make(map[string]int)
	}
	t.mu.Unlock()
}

func (t *topic) remove(s *subscriber) {
	t.mu.Lock()
	for i, item := range t.subs {
		if item == s {
			t.subs = append(t.subs[:i], t.subs[i+1:]...)
			break
		}
	}
	t.mu.Unlock()
	s.close()
}

// drain 取出并清空所有订阅者。
func (t *topic) drain() []*subscriber {
	t.mu.Lock()
	defer t.mu.Unlock()
	subs := t.subs
	t.subs = nil
	return subs
}

// deliver 分发消息：Group 为空的订阅者各得一份；同 Group 内只有一个订阅者得到。
func (t *topic) deliver(m *message.Message) {
	t.mu.Lock()
	defer t.mu.Unlock()

	sent := make(map[string]bool, len(t.subs))
	for _, s := range t.subs {
		if s.group == "" {
			s.send(m)
			continue
		}
		if sent[s.group] {
			continue
		}
		if s.send(m) {
			sent[s.group] = true
		}
	}
}

// subscriber 单个订阅者。
type subscriber struct {
	ctx   context.Context
	group string
	ch    chan *message.Message

	mu     sync.RWMutex
	closed bool

	wg      sync.WaitGroup
	closeMu sync.Once
	broker  *Broker
}

func newSubscriber(ctx context.Context, group string, b *Broker) *subscriber {
	return &subscriber{
		ctx:    ctx,
		group:  group,
		ch:     make(chan *message.Message, defaultBufferSize),
		broker: b,
	}
}

// send 非阻塞投递，缓冲满则丢弃。
func (s *subscriber) send(m *message.Message) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return false
	}
	select {
	case s.ch <- m:
		return true
	default:
		if s.broker != nil {
			atomic.AddInt64(&s.broker.dropped, 1)
		}
		return false
	}
}

func (s *subscriber) close() {
	s.closeMu.Do(func() {
		s.mu.Lock()
		s.closed = true
		s.mu.Unlock()
		close(s.ch)
	})
}

func (s *subscriber) run(h message.Handler) {
	defer s.wg.Done()
	for m := range s.ch {
		if err := h(s.ctx, m); err != nil {
			// 处理失败已由 RetryMW 重试；这里交给驱动自行记录，
			// memory 驱动选择忽略，避免日志风暴。
			continue
		}
	}
}
