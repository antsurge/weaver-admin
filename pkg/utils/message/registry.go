package message

import (
	"context"
	"sort"
	"sync"
)

// Factory 根据配置创建 Broker。
type Factory func(*Config) (Broker, error)

var (
	registryMu sync.RWMutex
	factories  = make(map[string]Factory)
)

// Register 注册一个驱动。通常在驱动包的 init() 中调用。
// 重复注册同名驱动会覆盖（后注册者优先）。
func Register(driver string, f Factory) {
	if driver == "" || f == nil {
		return
	}
	registryMu.Lock()
	factories[driver] = f
	registryMu.Unlock()
}

// Unregister 注销驱动，主要用于测试。
func Unregister(driver string) {
	registryMu.Lock()
	delete(factories, driver)
	registryMu.Unlock()
}

// Drivers 返回已注册的驱动名（已排序）。
func Drivers() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	names := make([]string, 0, len(factories))
	for name := range factories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// New 按配置创建 Broker。
//
// 特殊处理：
//   - Enabled == false 或 Driver 为 "" / "noop" → 返回内置空实现，消息被静默丢弃，
//     保证消息中间件不可用时业务仍能启动（通知照常落库，只是不实时推送）。
func New(c *Config) (Broker, error) {
	if c == nil {
		c = &Config{}
	}
	if !c.Enabled || c.Driver == "" || c.Driver == DriverNoop {
		return noopBroker{}, nil
	}
	registryMu.RLock()
	f, ok := factories[c.Driver]
	registryMu.RUnlock()
	if !ok || f == nil {
		return nil, ErrDriverNotFound
	}
	return f(c)
}

// DriverNoop 空实现驱动名。
const DriverNoop = "noop"

// noopBroker 内置的空实现：所有操作都是安全的空操作。
type noopBroker struct{}

func (noopBroker) Name() string                               { return DriverNoop }
func (noopBroker) Publish(context.Context, ...*Message) error { return nil }
func (noopBroker) Subscribe(context.Context, string, SubscribeOption, Handler) error {
	return nil
}
func (noopBroker) Ping(context.Context) error { return nil }
func (noopBroker) Close() error               { return nil }
