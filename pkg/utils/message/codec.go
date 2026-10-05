package message

import (
	"encoding/json"
	"sync"
)

// Codec 负责消息负载的序列化。
type Codec interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
	// ContentType 负载的内容类型，会写入 Content-Type 头（部分驱动使用）。
	ContentType() string
}

// JSONCodec 默认编解码器。
type JSONCodec struct{}

func (JSONCodec) Marshal(v any) ([]byte, error) { return json.Marshal(v) }

func (JSONCodec) Unmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }

func (JSONCodec) ContentType() string { return "application/json" }

var (
	codecMu      sync.RWMutex
	defaultCodec Codec = JSONCodec{}
)

// SetDefaultCodec 替换默认的负载编解码器（如改为 protobuf / msgpack）。
func SetDefaultCodec(c Codec) {
	if c == nil {
		return
	}
	codecMu.Lock()
	defaultCodec = c
	codecMu.Unlock()
}

// DefaultCodec 返回当前默认的负载编解码器。
func DefaultCodec() Codec {
	codecMu.RLock()
	defer codecMu.RUnlock()
	return defaultCodec
}
