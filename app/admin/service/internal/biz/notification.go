package biz

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/message"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

// ───────────────────────── 常量区 ─────────────────────────

// 通知类型
const (
	NotificationTypeSystem   = "system"   // 系统
	NotificationTypeAnnounce = "announce" // 公告
	NotificationTypeAudit    = "audit"    // 审批
	NotificationTypeTodo     = "todo"     // 待办
	NotificationTypeAlert    = "alert"    // 告警
)

// 通知级别
const (
	NotificationLevelInfo    = "info"
	NotificationLevelSuccess = "success"
	NotificationLevelWarn    = "warn"
	NotificationLevelError   = "error"
)

// 通知状态
const (
	NotificationStatusDraft     = "draft"
	NotificationStatusPublished = "published"
	NotificationStatusRevoked   = "revoked"
)

// 投放范围
const (
	NotifyTargetAll  = "all"  // 全体用户
	NotifyTargetUser = "user" // 指定用户
	NotifyTargetRole = "role" // 指定角色
)

// TopicNotification 通知广播主题。
// 订阅时 Group 留空 → 广播语义：每个实例都会收到，用于把事件扇出到本实例的 SSE 连接。
const TopicNotification = "notification"

// 事件动作
const (
	NotificationActionCreated = "created"
	NotificationActionRevoked = "revoked"
)

// 错误定义
var (
	ErrNotificationNotFound   = errors.NotFound("NOTIFICATION_NOT_FOUND", "notification not found")
	ErrNotificationNoTarget   = errors.BadRequest("NOTIFICATION_NO_TARGET", "no target user matched")
	ErrNotificationBadRequest = errors.BadRequest("NOTIFICATION_BAD_REQUEST", "invalid notification argument")
)

// ───────────────────────── 领域模型 ─────────────────────────

// Notification 通知（正文 + 投放范围）。
type Notification struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Type        string     `json:"type"`
	Level       string     `json:"level"`
	BizType     string     `json:"bizType,omitempty"`
	BizID       string     `json:"bizId,omitempty"`
	SenderID    string     `json:"senderId,omitempty"`
	SenderName  string     `json:"senderName,omitempty"`
	Link        string     `json:"link,omitempty"`
	TargetType  string     `json:"targetType"`
	TargetIDs   []string   `json:"targetIds,omitempty"`
	Status      string     `json:"status"`
	PublishedAt *time.Time `json:"publishedAt,omitempty"`
	ExpireAt    *time.Time `json:"expireAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// NotificationRecord 用户收件箱条目：谁收到了哪条通知、是否已读。
// 这是「用户能看到什么」的真相源，实时推送只是加速通道。
type NotificationRecord struct {
	ID             string     `json:"id"`
	NotificationID string     `json:"notificationId"`
	UserID         string     `json:"userId"`
	IsRead         bool       `json:"isRead"`
	ReadAt         *time.Time `json:"readAt,omitempty"`
	IsDeleted      bool       `json:"isDeleted"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`

	// 联带通知正文（列表接口填充）
	Notification *Notification `json:"notification,omitempty"`
}

// NotificationEvent 跨实例广播的事件（用于 SSE 推送扇出）。
// 只携带展示所需的最小字段，避免消息体过大。
type NotificationEvent struct {
	Action    string   `json:"action"` // created / revoked
	ID        string   `json:"id"`     // 收件记录或通知 ID
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Type      string   `json:"type"`
	Level     string   `json:"level"`
	Link      string   `json:"link"`
	CreatedAt string   `json:"createdAt"`           // RFC3339
	Broadcast bool     `json:"broadcast,omitempty"` // true=推给所有在线用户
	UserIDs   []string `json:"userIds,omitempty"`   // Broadcast=false 时指定接收人
}

// ───────────────────────── 请求/响应 ─────────────────────────

// ListNotificationRequest 我的消息列表
type ListNotificationRequest struct {
	enthelper.PaginationParams
	UserID  string `form:"-" query:"-"`
	IsRead  *bool  `form:"isRead" query:"isRead"`
	Type    string `form:"type" query:"type"`
	Level   string `form:"level" query:"level"`
	Keyword string `form:"keyword" query:"keyword"`
}

type ListNotificationResponse struct {
	Data  []*NotificationRecord
	Total int
}

// ListNotificationManageRequest 管理端通知列表
type ListNotificationManageRequest struct {
	enthelper.PaginationParams
	Title          string `form:"title" query:"title"`
	Type           string `form:"type" query:"type"`
	Level          string `form:"level" query:"level"`
	Status         string `form:"status" query:"status"`
	IncludeDeleted bool   `form:"-" query:"-"`
}

type ListNotificationManageResponse struct {
	Data  []*Notification
	Total int
}

// PublishNotificationRequest 发布通知
type PublishNotificationRequest struct {
	Title      string
	Content    string
	Type       string
	Level      string
	BizType    string
	BizID      string
	Link       string
	TargetType string
	TargetIDs  []string
	SenderID   string
	SenderName string
}

// ───────────────────────── 仓储接口 ─────────────────────────

type NotificationRepo interface {
	// 管理端
	CreateWithFanOut(ctx context.Context, n *Notification, userIDs []string) error
	GetByID(ctx context.Context, id string) (*Notification, error)
	ListManage(ctx context.Context, req *ListNotificationManageRequest) (*ListNotificationManageResponse, error)
	UpdateStatus(ctx context.Context, id string, status string) error

	// 用户侧收件箱
	ListInbox(ctx context.Context, req *ListNotificationRequest) (*ListNotificationResponse, error)
	CountUnread(ctx context.Context, userID string) (int64, error)
	MarkRead(ctx context.Context, userID string, ids []string) error
	MarkAllRead(ctx context.Context, userID string) error
	SoftDeleteRecords(ctx context.Context, userID string, ids []string) error
	DeleteByNotificationID(ctx context.Context, notificationID string) error

	// 投放目标解析
	ResolveUserIDs(ctx context.Context, targetType string, targetIDs []string) ([]string, error)
}

// ───────────────────────── 用例 ─────────────────────────

type NotificationUsecase struct {
	repo   NotificationRepo
	broker message.Broker
	hub    *NotificationHub
	log    *log.Helper
}

func NewNotificationUsecase(
	repo NotificationRepo,
	broker message.Broker,
	hub *NotificationHub,
	logger log.Logger,
) *NotificationUsecase {
	return &NotificationUsecase{
		repo:   repo,
		broker: broker,
		hub:    hub,
		log:    log.NewHelper(logger),
	}
}

// Publish 发布通知：写库 + 扇出到收件箱 + 广播推送事件。
// 推送失败只记日志，不影响发布结果（真相源已落库，前端刷新即可补齐）。
func (uc *NotificationUsecase) Publish(ctx context.Context, req *PublishNotificationRequest) (*Notification, error) {
	if req == nil || req.Title == "" {
		return nil, ErrNotificationBadRequest
	}
	targetType := req.TargetType
	if targetType == "" {
		targetType = NotifyTargetAll
	}

	userIDs, err := uc.repo.ResolveUserIDs(ctx, targetType, req.TargetIDs)
	if err != nil {
		return nil, err
	}
	if targetType != NotifyTargetAll && len(userIDs) == 0 {
		return nil, ErrNotificationNoTarget
	}

	now := time.Now()
	n := &Notification{
		ID:          uuid.GenerateXID(),
		Title:       req.Title,
		Content:     req.Content,
		Type:        defaultString(req.Type, NotificationTypeSystem),
		Level:       defaultString(req.Level, NotificationLevelInfo),
		BizType:     req.BizType,
		BizID:       req.BizID,
		SenderID:    req.SenderID,
		SenderName:  req.SenderName,
		Link:        req.Link,
		TargetType:  targetType,
		TargetIDs:   req.TargetIDs,
		Status:      NotificationStatusPublished,
		PublishedAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.repo.CreateWithFanOut(ctx, n, userIDs); err != nil {
		return nil, err
	}

	uc.emit(ctx, &NotificationEvent{
		Action:    NotificationActionCreated,
		ID:        n.ID,
		Title:     n.Title,
		Content:   n.Content,
		Type:      n.Type,
		Level:     n.Level,
		Link:      n.Link,
		CreatedAt: n.CreatedAt.Format(time.RFC3339),
		Broadcast: targetType == NotifyTargetAll,
		UserIDs:   userIDs,
	})
	return n, nil
}

// Revoke 撤回通知：改状态 + 删除收件记录 + 广播撤回事件。
func (uc *NotificationUsecase) Revoke(ctx context.Context, id string) error {
	if id == "" {
		return ErrNotificationBadRequest
	}
	n, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if n == nil {
		return ErrNotificationNotFound
	}
	if err := uc.repo.UpdateStatus(ctx, id, NotificationStatusRevoked); err != nil {
		return err
	}
	if err := uc.repo.DeleteByNotificationID(ctx, id); err != nil {
		return err
	}

	uc.emit(ctx, &NotificationEvent{
		Action:    NotificationActionRevoked,
		ID:        id,
		Broadcast: true,
	})
	return nil
}

// ListInbox 我的消息列表
func (uc *NotificationUsecase) ListInbox(ctx context.Context, req *ListNotificationRequest) (*ListNotificationResponse, error) {
	return uc.repo.ListInbox(ctx, req)
}

// CountUnread 我的未读数
func (uc *NotificationUsecase) CountUnread(ctx context.Context, userID string) (int64, error) {
	return uc.repo.CountUnread(ctx, userID)
}

// MarkRead 标记已读（支持批量）
func (uc *NotificationUsecase) MarkRead(ctx context.Context, userID string, ids []string) error {
	return uc.repo.MarkRead(ctx, userID, ids)
}

// MarkAllRead 全部已读
func (uc *NotificationUsecase) MarkAllRead(ctx context.Context, userID string) error {
	return uc.repo.MarkAllRead(ctx, userID)
}

// DeleteRecords 删除（软删）我的消息
func (uc *NotificationUsecase) DeleteRecords(ctx context.Context, userID string, ids []string) error {
	return uc.repo.SoftDeleteRecords(ctx, userID, ids)
}

// ListManage 管理端通知列表
func (uc *NotificationUsecase) ListManage(ctx context.Context, req *ListNotificationManageRequest) (*ListNotificationManageResponse, error) {
	return uc.repo.ListManage(ctx, req)
}

// StartConsumer 启动跨实例事件消费（广播语义：每个实例都会收到）。
// ctx 取消时自动退订。
func (uc *NotificationUsecase) StartConsumer(ctx context.Context) error {
	if uc.broker == nil || uc.hub == nil {
		return nil
	}
	return message.SubscribeJSON[NotificationEvent](ctx, uc.broker, TopicNotification, message.SubscribeOption{},
		func(_ context.Context, _ *message.Message, ev NotificationEvent) error {
			uc.hub.Dispatch(&ev)
			return nil
		})
}

// emit 广播事件。失败只记日志：推送是加速通道，不是真相源。
func (uc *NotificationUsecase) emit(ctx context.Context, ev *NotificationEvent) {
	if uc.broker == nil {
		// 无 Broker（如 noop 驱动）时仍尝试本机直投，保证单机可用
		if uc.hub != nil {
			uc.hub.Dispatch(ev)
		}
		return
	}
	if err := message.PublishJSON(ctx, uc.broker, TopicNotification, ev); err != nil {
		uc.log.Errorf("publish notification event failed: %v", err)
		// 降级：至少推给本实例在线的用户
		if uc.hub != nil {
			uc.hub.Dispatch(ev)
		}
	}
}

func defaultString(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
