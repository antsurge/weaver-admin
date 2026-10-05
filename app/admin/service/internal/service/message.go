package service

import (
	"context"
	"time"

	adminV1 "github.com/antsurge/weaver-admin/api/gen/go/admin/service/v1"
	messageV1 "github.com/antsurge/weaver-admin/api/gen/go/message/service/v1"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/pkg/metadata"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ErrNotificationUnauthorized = errors.Unauthorized("NOTIFICATION_UNAUTHORIZED", "login required")

// MessageService 消息通知服务（REST 部分；实时推送见 handler.NotificationHandler）。
type MessageService struct {
	adminV1.UnimplementedMessageServer

	notificationUc *biz.NotificationUsecase
	log            *log.Helper
}

func NewMessageService(notificationUc *biz.NotificationUsecase, logger log.Logger) *MessageService {
	return &MessageService{
		notificationUc: notificationUc,
		log:            log.NewHelper(logger),
	}
}

// ListNotification 我的消息列表
func (s *MessageService) ListNotification(ctx context.Context, req *messageV1.ListNotificationRequest) (*messageV1.ListNotificationResponse, error) {
	adminID := metadata.GetAdminID(ctx)
	if adminID == "" {
		return nil, ErrNotificationUnauthorized
	}

	input := &biz.ListNotificationRequest{
		UserID: adminID,
		Type:   req.GetType(),
		Level:  req.GetLevel(),
	}
	input.CurrentPage = int(req.GetCurrentPage())
	input.PageSize = int(req.GetPageSize())
	if req.IsRead != nil {
		v := req.GetIsRead()
		input.IsRead = &v
	}

	res, err := s.notificationUc.ListInbox(ctx, input)
	if err != nil {
		return nil, err
	}

	out := &messageV1.ListNotificationResponse{Total: int64(res.Total)}
	out.Items = make([]*messageV1.NotificationRecord, 0, len(res.Data))
	for _, item := range res.Data {
		out.Items = append(out.Items, toPbRecord(item))
	}
	return out, nil
}

// GetUnreadCount 未读数量
func (s *MessageService) GetUnreadCount(ctx context.Context, _ *emptypb.Empty) (*messageV1.GetUnreadCountResponse, error) {
	adminID := metadata.GetAdminID(ctx)
	if adminID == "" {
		return nil, ErrNotificationUnauthorized
	}
	count, err := s.notificationUc.CountUnread(ctx, adminID)
	if err != nil {
		return nil, err
	}
	return &messageV1.GetUnreadCountResponse{Count: count}, nil
}

// ReadNotification 标记已读
func (s *MessageService) ReadNotification(ctx context.Context, req *messageV1.ReadNotificationRequest) (*emptypb.Empty, error) {
	adminID := metadata.GetAdminID(ctx)
	if adminID == "" {
		return nil, ErrNotificationUnauthorized
	}
	if req.GetId() == "" {
		return nil, biz.ErrNotificationBadRequest
	}
	if err := s.notificationUc.MarkRead(ctx, adminID, []string{req.GetId()}); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ReadAllNotification 全部已读
func (s *MessageService) ReadAllNotification(ctx context.Context, _ *messageV1.ReadAllNotificationRequest) (*emptypb.Empty, error) {
	adminID := metadata.GetAdminID(ctx)
	if adminID == "" {
		return nil, ErrNotificationUnauthorized
	}
	if err := s.notificationUc.MarkAllRead(ctx, adminID); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// DeleteNotification 删除消息（软删）
func (s *MessageService) DeleteNotification(ctx context.Context, req *messageV1.DeleteNotificationRequest) (*emptypb.Empty, error) {
	adminID := metadata.GetAdminID(ctx)
	if adminID == "" {
		return nil, ErrNotificationUnauthorized
	}
	if len(req.GetIds()) == 0 {
		return nil, biz.ErrNotificationBadRequest
	}
	if err := s.notificationUc.DeleteRecords(ctx, adminID, req.GetIds()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ListNotificationManage 管理端通知列表
func (s *MessageService) ListNotificationManage(ctx context.Context, req *messageV1.ListNotificationManageRequest) (*messageV1.ListNotificationManageResponse, error) {
	input := &biz.ListNotificationManageRequest{
		Title:  req.GetTitle(),
		Type:   req.GetType(),
		Level:  req.GetLevel(),
		Status: req.GetStatus(),
	}
	input.CurrentPage = int(req.GetCurrentPage())
	input.PageSize = int(req.GetPageSize())

	res, err := s.notificationUc.ListManage(ctx, input)
	if err != nil {
		return nil, err
	}

	out := &messageV1.ListNotificationManageResponse{Total: int64(res.Total)}
	out.Items = make([]*messageV1.Notification, 0, len(res.Data))
	for _, item := range res.Data {
		out.Items = append(out.Items, toPbNotification(item))
	}
	return out, nil
}

// CreateNotification 发布通知
func (s *MessageService) CreateNotification(ctx context.Context, req *messageV1.CreateNotificationRequest) (*messageV1.Notification, error) {
	adminID := metadata.GetAdminID(ctx)
	if adminID == "" {
		return nil, ErrNotificationUnauthorized
	}

	n, err := s.notificationUc.Publish(ctx, &biz.PublishNotificationRequest{
		Title:      req.GetTitle(),
		Content:    req.GetContent(),
		Type:       req.GetType(),
		Level:      req.GetLevel(),
		BizType:    req.GetBizType(),
		BizID:      req.GetBizId(),
		Link:       req.GetLink(),
		TargetType: req.GetTargetType(),
		TargetIDs:  req.GetTargetIds(),
		SenderID:   adminID,
	})
	if err != nil {
		return nil, err
	}
	return toPbNotification(n), nil
}

// RevokeNotification 撤回通知
func (s *MessageService) RevokeNotification(ctx context.Context, req *messageV1.RevokeNotificationRequest) (*emptypb.Empty, error) {
	if req.GetId() == "" {
		return nil, biz.ErrNotificationBadRequest
	}
	if err := s.notificationUc.Revoke(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ───────────────────────── 模型转换 ─────────────────────────

func toPbNotification(n *biz.Notification) *messageV1.Notification {
	if n == nil {
		return nil
	}
	return &messageV1.Notification{
		Id:          n.ID,
		Title:       n.Title,
		Content:     n.Content,
		Type:        n.Type,
		Level:       n.Level,
		BizType:     n.BizType,
		BizId:       n.BizID,
		SenderId:    n.SenderID,
		SenderName:  n.SenderName,
		Link:        n.Link,
		TargetType:  n.TargetType,
		TargetIds:   n.TargetIDs,
		Status:      n.Status,
		PublishedAt: formatTimePtr(n.PublishedAt),
		ExpireAt:    formatTimePtr(n.ExpireAt),
		CreatedAt:   formatTime(n.CreatedAt),
		UpdatedAt:   formatTime(n.UpdatedAt),
	}
}

func toPbRecord(r *biz.NotificationRecord) *messageV1.NotificationRecord {
	if r == nil {
		return nil
	}
	return &messageV1.NotificationRecord{
		Id:             r.ID,
		NotificationId: r.NotificationID,
		UserId:         r.UserID,
		IsRead:         r.IsRead,
		ReadAt:         formatTimePtr(r.ReadAt),
		IsDeleted:      r.IsDeleted,
		CreatedAt:      formatTime(r.CreatedAt),
		UpdatedAt:      formatTime(r.UpdatedAt),
		Notification:   toPbNotification(r.Notification),
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return formatTime(*t)
}
