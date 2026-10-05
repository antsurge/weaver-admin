package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/pkg/metadata"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

const (
	// sseHeartbeat 心跳间隔，需小于常见反向代理的 60s 读超时。
	sseHeartbeat = 15 * time.Second
	// sseEventBuffer 单个连接的事件缓冲，满了丢弃（真相源在数据库，前端可重新拉取）。
	sseEventBuffer = 32
)

// NotificationHandler 消息通知相关的自定义 HTTP 处理（SSE 实时推送）。
type NotificationHandler struct {
	notificationUc *biz.NotificationUsecase
	hub            *biz.NotificationHub
}

func NewNotificationHandler(
	notificationUc *biz.NotificationUsecase,
	hub *biz.NotificationHub,
) *NotificationHandler {
	return &NotificationHandler{notificationUc: notificationUc, hub: hub}
}

// StreamNotification SSE 长连接：向当前登录用户推送通知事件。
//
// 事件类型：
//   - connected  连接建立确认
//   - notification 新通知 / 撤回（data 为 NotificationEvent）
//   - 注释帧 ": ping" 心跳
//
// 鉴权：EventSource 不支持自定义 Header，token 通过 query 参数 ?token= 传递
// （见 pkg/middleware/auth）。
func (h *NotificationHandler) StreamNotification(ctx khttp.Context) error {
	adminID := metadata.GetAdminID(ctx)
	if adminID == "" {
		ctx.Response().WriteHeader(http.StatusUnauthorized)
		return nil
	}

	w := ctx.Response()
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	// 关闭 Nginx 缓冲，否则事件会被攒着不下发
	w.Header().Set("X-Accel-Buffering", "no")

	flusher := resolveFlusher(w)

	sub := h.hub.Subscribe(adminID, sseEventBuffer)
	defer h.hub.Unsubscribe(sub)

	writeSSE(w, flusher, "connected", map[string]any{"userId": adminID})

	ticker := time.NewTicker(sseHeartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case ev, ok := <-sub.C():
			if !ok {
				return nil
			}
			if writeSSE(w, flusher, "notification", ev) != nil {
				return nil
			}
		case <-ticker.C:
			if _, err := io.WriteString(w, ": ping\n\n"); err != nil {
				return nil
			}
			if flusher != nil {
				flusher.Flush()
			}
		}
	}
}

// resolveFlusher 取出底层 http.Flusher。
// kratos 用 responseWriter 包了一层（未实现 Flusher），需通过 Unwrap 拿到原始 writer。
func resolveFlusher(w http.ResponseWriter) http.Flusher {
	if f, ok := w.(http.Flusher); ok {
		return f
	}
	if u, ok := w.(interface{ Unwrap() http.ResponseWriter }); ok {
		if f, ok := u.Unwrap().(http.Flusher); ok {
			return f
		}
	}
	return nil
}

// writeSSE 写入一个 SSE 事件帧并立即 flush。
func writeSSE(w http.ResponseWriter, f http.Flusher, event string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	var sb strings.Builder
	sb.WriteString("event: ")
	sb.WriteString(event)
	sb.WriteString("\ndata: ")
	sb.Write(payload)
	sb.WriteString("\n\n")
	if _, err := io.WriteString(w, sb.String()); err != nil {
		return err
	}
	if f != nil {
		f.Flush()
	}
	return nil
}
