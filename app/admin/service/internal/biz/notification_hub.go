package biz

import "sync"

// NotificationSubscriber 一个 SSE 连接的订阅端。
type NotificationSubscriber struct {
	UserID string
	ch     chan *NotificationEvent
	once   sync.Once
}

// C 返回只读事件通道，供 SSE handler 消费。
func (s *NotificationSubscriber) C() <-chan *NotificationEvent { return s.ch }

func (s *NotificationSubscriber) close() {
	s.once.Do(func() { close(s.ch) })
}

// trySend 非阻塞投递：客户端消费过慢时丢弃该条。
// 丢弃是安全的——真相源在数据库，前端刷新或重连即可补齐。
func (s *NotificationSubscriber) trySend(ev *NotificationEvent) {
	select {
	case s.ch <- ev:
	default:
	}
}

// NotificationHub 本实例内「用户 → SSE 连接」的注册表。
// 只维护本进程的连接；跨实例靠消息中间件广播事件后再各自查 Hub 投递。
type NotificationHub struct {
	mu    sync.RWMutex
	users map[string]map[*NotificationSubscriber]struct{}
}

func NewNotificationHub() *NotificationHub {
	return &NotificationHub{users: make(map[string]map[*NotificationSubscriber]struct{})}
}

// Subscribe 注册一个用户的订阅端，返回后可通过 C() 读取事件。
// 调用方负责在连接断开时调用 Unsubscribe。
func (h *NotificationHub) Subscribe(userID string, buffer int) *NotificationSubscriber {
	if buffer <= 0 {
		buffer = 32
	}
	s := &NotificationSubscriber{UserID: userID, ch: make(chan *NotificationEvent, buffer)}

	h.mu.Lock()
	set, ok := h.users[userID]
	if !ok {
		set = make(map[*NotificationSubscriber]struct{})
		h.users[userID] = set
	}
	set[s] = struct{}{}
	h.mu.Unlock()
	return s
}

// Unsubscribe 注销订阅端并关闭其通道。
func (h *NotificationHub) Unsubscribe(s *NotificationSubscriber) {
	if s == nil {
		return
	}
	h.mu.Lock()
	if set, ok := h.users[s.UserID]; ok {
		delete(set, s)
		if len(set) == 0 {
			delete(h.users, s.UserID)
		}
	}
	h.mu.Unlock()
	s.close()
}

// Dispatch 按事件投放范围分发到本实例在线的连接。
func (h *NotificationHub) Dispatch(ev *NotificationEvent) {
	if ev == nil {
		return
	}
	if ev.Broadcast {
		h.Broadcast(ev)
		return
	}
	h.SendTo(ev.UserIDs, ev)
}

// SendTo 推给指定用户（多端登录时每个连接各推一份）。
func (h *NotificationHub) SendTo(userIDs []string, ev *NotificationEvent) {
	if len(userIDs) == 0 || ev == nil {
		return
	}
	h.mu.RLock()
	targets := make([]*NotificationSubscriber, 0, len(userIDs))
	for _, uid := range userIDs {
		for s := range h.users[uid] {
			targets = append(targets, s)
		}
	}
	h.mu.RUnlock()

	for _, s := range targets {
		s.trySend(ev)
	}
}

// Broadcast 推给本实例所有在线连接。
func (h *NotificationHub) Broadcast(ev *NotificationEvent) {
	if ev == nil {
		return
	}
	h.mu.RLock()
	targets := make([]*NotificationSubscriber, 0, len(h.users))
	for _, set := range h.users {
		for s := range set {
			targets = append(targets, s)
		}
	}
	h.mu.RUnlock()

	for _, s := range targets {
		s.trySend(ev)
	}
}

// OnlineUsers 当前在线用户数（观测用）。
func (h *NotificationHub) OnlineUsers() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.users)
}
