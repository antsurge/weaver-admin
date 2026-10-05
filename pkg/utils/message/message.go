// Package message 定义与具体中间件无关的消息抽象。
//
// 设计目标：业务层只依赖本包的接口，底层可在 memory / redis / rabbitmq / kafka
// 之间自由切换，切换成本 = 改一行配置。
//
// 核心语义约定（所有驱动必须遵守）：
//   - SubscribeOption.Group == ""  → 广播：每个订阅者都会收到一份副本
//   - SubscribeOption.Group != ""  → 负载均衡：同一 Group 内只有一个订阅者收到
package message

import (
	"encoding/json"
	"time"

	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
)

// Message 抽象消息信封。与任何中间件实现无关。
type Message struct {
	// ID 消息唯一标识，用于幂等与链路追踪；为空时 Publish 会自动生成。
	ID string `json:"id,omitempty"`
	// Topic 主题名。Publish/Subscribe 时会自动拼接 Config.Prefix。
	Topic string `json:"topic,omitempty"`
	// Key 路由/分区键，用于保证同一 Key 的消息顺序（部分驱动支持）。
	Key string `json:"key,omitempty"`
	// Headers 元数据，用于透传 trace_id、租户、消息类型等。
	Headers map[string]string `json:"headers,omitempty"`
	// Body 序列化后的负载。
	Body []byte `json:"body"`
	// Timestamp 消息产生时间。
	Timestamp time.Time `json:"timestamp,omitempty"`
	// Delay 延迟投递时间，0 表示立即投递（部分驱动支持）。
	Delay time.Duration `json:"delay,omitempty"`
}

// Option 构造 Message 的可选参数。
type Option func(*Message)

// WithID 指定消息 ID（幂等键）。
func WithID(id string) Option {
	return func(m *Message) { m.ID = id }
}

// WithKey 指定路由/分区键。
func WithKey(key string) Option {
	return func(m *Message) { m.Key = key }
}

// WithHeader 追加单个元数据。
func WithHeader(k, v string) Option {
	return func(m *Message) {
		if m.Headers == nil {
			m.Headers = make(map[string]string)
		}
		m.Headers[k] = v
	}
}

// WithHeaders 批量追加元数据。
func WithHeaders(h map[string]string) Option {
	return func(m *Message) {
		if len(h) == 0 {
			return
		}
		if m.Headers == nil {
			m.Headers = make(map[string]string, len(h))
		}
		for k, v := range h {
			m.Headers[k] = v
		}
	}
}

// WithDelay 延迟投递。
func WithDelay(d time.Duration) Option {
	return func(m *Message) { m.Delay = d }
}

// WithTimestamp 指定消息时间。
func WithTimestamp(t time.Time) Option {
	return func(m *Message) { m.Timestamp = t }
}

// NewMessage 构造一条原始消息（Body 已是序列化后的字节）。
func NewMessage(topic string, body []byte, opts ...Option) *Message {
	m := &Message{Topic: topic, Body: body}
	for _, opt := range opts {
		if opt != nil {
			opt(m)
		}
	}
	Normalize(m)
	return m
}

// NewJSONMessage 构造一条消息，负载由默认 Codec 序列化。
func NewJSONMessage(topic string, v any, opts ...Option) (*Message, error) {
	body, err := DefaultCodec().Marshal(v)
	if err != nil {
		return nil, err
	}
	return NewMessage(topic, body, opts...), nil
}

// Normalize 补齐默认值并做基本校验。驱动在 Publish 前应调用。
func Normalize(m *Message) error {
	if m == nil {
		return ErrNilMessage
	}
	if m.Topic == "" {
		return ErrEmptyTopic
	}
	if m.ID == "" {
		m.ID = uuid.GenerateXID()
	}
	if m.Timestamp.IsZero() {
		m.Timestamp = time.Now()
	}
	return nil
}

// GetHeader 读取元数据。
func (m *Message) GetHeader(k string) string {
	if m == nil || m.Headers == nil {
		return ""
	}
	return m.Headers[k]
}

// Marshal 把消息序列化为可跨进程传输的字节（信封 + 负载）。
func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// Unmarshal 反序列化消息。
func Unmarshal(data []byte) (*Message, error) {
	m := &Message{}
	if err := json.Unmarshal(data, m); err != nil {
		return nil, err
	}
	return m, nil
}

// Decode 把负载反序列化到 v。
func (m *Message) Decode(v any) error {
	if m == nil {
		return ErrNilMessage
	}
	return DefaultCodec().Unmarshal(m.Body, v)
}

// WithPrefix 为 topic 拼接前缀，已含前缀时原样返回。
func WithPrefix(prefix, topic string) string {
	if prefix == "" {
		return topic
	}
	if len(topic) >= len(prefix) && topic[:len(prefix)] == prefix {
		return topic
	}
	return prefix + topic
}
