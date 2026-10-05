package message

import (
	"errors"
	"fmt"
	"strconv"
)

// Config 消息中间件配置。与具体驱动无关，驱动按需读取自己关心的字段。
//
// 约定：所有驱动私有的参数放 Options，避免每加一个驱动就改一次配置结构。
type Config struct {
	// Driver 驱动名：memory / redis / rabbitmq / kafka / noop。
	Driver string
	// Enabled 总开关，false 时 New 返回内置空实现（消息被丢弃但不报错）。
	Enabled bool
	// Addrs 地址列表，多数驱动只用第一个。
	Addrs []string
	// Username 用户名。
	Username string
	// Password 密码。
	Password string
	// VHost 虚拟主机（RabbitMQ）。
	VHost string
	// Prefix topic 前缀，如 "weaver."。
	Prefix string
	// Options 驱动私有参数，如 db、pool_size、network。
	Options map[string]string
}

// Validate 基本校验。
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("message: nil config")
	}
	if c.Driver == "" {
		return errors.New("message: empty driver")
	}
	return nil
}

// Addr 返回第一个地址，无则空字符串。
func (c *Config) Addr() string {
	if c == nil || len(c.Addrs) == 0 {
		return ""
	}
	return c.Addrs[0]
}

// Option 读取驱动私有参数，缺失时返回默认值。
func (c *Config) Option(key, fallback string) string {
	if c == nil || c.Options == nil {
		return fallback
	}
	if v, ok := c.Options[key]; ok && v != "" {
		return v
	}
	return fallback
}

// OptionInt 读取整型驱动私有参数。
func (c *Config) OptionInt(key string, fallback int) int {
	v := c.Option(key, "")
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

// OptionBool 读取布尔型驱动私有参数。
func (c *Config) OptionBool(key string, fallback bool) bool {
	v := c.Option(key, "")
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

// String 便于日志打印（不输出密码）。
func (c *Config) String() string {
	if c == nil {
		return "<nil>"
	}
	return fmt.Sprintf("driver=%s enabled=%v addrs=%v prefix=%q", c.Driver, c.Enabled, c.Addrs, c.Prefix)
}
