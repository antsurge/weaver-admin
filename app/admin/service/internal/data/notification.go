package data

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/admin"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/adminrole"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/notification"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/notificationrecord"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/predicate"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/log"
)

// fanOutBatchSize 扇出插入的批量大小，避免单次 SQL 过大。
const fanOutBatchSize = 500

var _ biz.NotificationRepo = (*notificationRepo)(nil)

type notificationRepo struct {
	data *Data
	log  *log.Helper
}

func NewNotificationRepo(data *Data, logger log.Logger) biz.NotificationRepo {
	return &notificationRepo{data: data, log: log.NewHelper(logger)}
}

// ───────────────────────── 管理端 ─────────────────────────

// CreateWithFanOut 在同一事务内写通知正文并扇出到所有接收人的收件箱。
func (r *notificationRepo) CreateWithFanOut(ctx context.Context, n *biz.Notification, userIDs []string) error {
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()

	create := tx.Notification.Create().
		SetID(n.ID).
		SetTitle(n.Title).
		SetContent(n.Content).
		SetType(notification.Type(n.Type)).
		SetLevel(notification.Level(n.Level)).
		SetTargetType(notification.TargetType(n.TargetType)).
		SetStatus(notification.Status(n.Status)).
		SetCreatedAt(n.CreatedAt).
		SetUpdatedAt(n.UpdatedAt)

	if n.BizType != "" {
		create = create.SetBizType(n.BizType)
	}
	if n.BizID != "" {
		create = create.SetBizID(n.BizID)
	}
	if n.SenderID != "" {
		create = create.SetSenderID(n.SenderID)
	}
	if n.SenderName != "" {
		create = create.SetSenderName(n.SenderName)
	}
	if n.Link != "" {
		create = create.SetLink(n.Link)
	}
	if len(n.TargetIDs) > 0 {
		create = create.SetTargetIds(n.TargetIDs)
	}
	if n.PublishedAt != nil {
		create = create.SetPublishedAt(*n.PublishedAt)
	}
	if n.ExpireAt != nil {
		create = create.SetExpireAt(*n.ExpireAt)
	}

	if err := create.Exec(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := fanOutRecords(ctx, tx, n.ID, userIDs); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// fanOutRecords 批量写入收件记录（自动去重）。
func fanOutRecords(ctx context.Context, tx *ent.Tx, notificationID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}
	now := time.Now()
	seen := make(map[string]struct{}, len(userIDs))
	builders := make([]*ent.NotificationRecordCreate, 0, len(userIDs))
	for _, uid := range userIDs {
		if uid == "" {
			continue
		}
		if _, dup := seen[uid]; dup {
			continue
		}
		seen[uid] = struct{}{}
		builders = append(builders, tx.NotificationRecord.Create().
			SetID(uuid.GenerateXID()).
			SetNotificationID(notificationID).
			SetUserID(uid).
			SetIsRead(false).
			SetIsDeleted(false).
			SetCreatedAt(now).
			SetUpdatedAt(now))
	}

	for i := 0; i < len(builders); i += fanOutBatchSize {
		end := i + fanOutBatchSize
		if end > len(builders) {
			end = len(builders)
		}
		if err := tx.NotificationRecord.CreateBulk(builders[i:end]...).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *notificationRepo) GetByID(ctx context.Context, id string) (*biz.Notification, error) {
	v, err := r.data.db.Notification.Query().
		Where(notification.IDEQ(id), notification.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return toBizNotification(v), nil
}

func (r *notificationRepo) ListManage(ctx context.Context, req *biz.ListNotificationManageRequest) (*biz.ListNotificationManageResponse, error) {
	query := r.data.db.Notification.Query().
		Order(ent.Desc(notification.FieldCreatedAt))
	if !req.IncludeDeleted {
		query = query.Where(notification.DeletedAtIsNil())
	}
	if v := req.Title; v != "" {
		query = query.Where(notification.TitleContains(v))
	}
	if v := req.Type; v != "" {
		query = query.Where(notification.TypeEQ(notification.Type(v)))
	}
	if v := req.Level; v != "" {
		query = query.Where(notification.LevelEQ(notification.Level(v)))
	}
	if v := req.Status; v != "" {
		query = query.Where(notification.StatusEQ(notification.Status(v)))
	}

	res, err := enthelper.Pagination[*ent.Notification, *ent.NotificationQuery](ctx, query, req.PaginationParams)
	if err != nil {
		return nil, err
	}

	data := make([]*biz.Notification, 0, res.Total)
	for _, v := range res.Data {
		data = append(data, toBizNotification(v))
	}
	return &biz.ListNotificationManageResponse{Data: data, Total: res.Total}, nil
}

func (r *notificationRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := r.data.db.Notification.Update().
		Where(notification.IDEQ(id)).
		SetStatus(notification.Status(status)).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

// ───────────────────────── 用户侧收件箱 ─────────────────────────

func (r *notificationRepo) ListInbox(ctx context.Context, req *biz.ListNotificationRequest) (*biz.ListNotificationResponse, error) {
	query := r.data.db.NotificationRecord.Query().
		Where(
			notificationrecord.UserIDEQ(req.UserID),
			notificationrecord.IsDeleted(false),
		).
		Order(ent.Desc(notificationrecord.FieldCreatedAt)).
		WithNotification()

	if req.IsRead != nil {
		query = query.Where(notificationrecord.IsReadEQ(*req.IsRead))
	}

	// 正文侧过滤条件（类型/级别/关键字）
	var ps []predicate.Notification
	if req.Type != "" {
		ps = append(ps, notification.TypeEQ(notification.Type(req.Type)))
	}
	if req.Level != "" {
		ps = append(ps, notification.LevelEQ(notification.Level(req.Level)))
	}
	if req.Keyword != "" {
		ps = append(ps, notification.Or(
			notification.TitleContains(req.Keyword),
			notification.ContentContains(req.Keyword),
		))
	}
	if len(ps) > 0 {
		query = query.Where(notificationrecord.HasNotificationWith(ps...))
	}

	res, err := enthelper.Pagination[*ent.NotificationRecord, *ent.NotificationRecordQuery](ctx, query, req.PaginationParams)
	if err != nil {
		return nil, err
	}

	data := make([]*biz.NotificationRecord, 0, res.Total)
	for _, v := range res.Data {
		data = append(data, toBizRecord(v))
	}
	return &biz.ListNotificationResponse{Data: data, Total: res.Total}, nil
}

func (r *notificationRepo) CountUnread(ctx context.Context, userID string) (int64, error) {
	count, err := r.data.db.NotificationRecord.Query().
		Where(
			notificationrecord.UserIDEQ(userID),
			notificationrecord.IsRead(false),
			notificationrecord.IsDeleted(false),
		).
		Count(ctx)
	return int64(count), err
}

func (r *notificationRepo) MarkRead(ctx context.Context, userID string, ids []string) error {
	now := time.Now()
	_, err := r.data.db.NotificationRecord.Update().
		Where(
			notificationrecord.UserIDEQ(userID),
			notificationrecord.IDIn(ids...),
			notificationrecord.IsRead(false),
		).
		SetIsRead(true).
		SetReadAt(now).
		SetUpdatedAt(now).
		Save(ctx)
	return err
}

func (r *notificationRepo) MarkAllRead(ctx context.Context, userID string) error {
	now := time.Now()
	_, err := r.data.db.NotificationRecord.Update().
		Where(
			notificationrecord.UserIDEQ(userID),
			notificationrecord.IsRead(false),
			notificationrecord.IsDeleted(false),
		).
		SetIsRead(true).
		SetReadAt(now).
		SetUpdatedAt(now).
		Save(ctx)
	return err
}

func (r *notificationRepo) SoftDeleteRecords(ctx context.Context, userID string, ids []string) error {
	_, err := r.data.db.NotificationRecord.Update().
		Where(
			notificationrecord.UserIDEQ(userID),
			notificationrecord.IDIn(ids...),
		).
		SetIsDeleted(true).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

func (r *notificationRepo) DeleteByNotificationID(ctx context.Context, notificationID string) error {
	_, err := r.data.db.NotificationRecord.Update().
		Where(notificationrecord.NotificationIDEQ(notificationID)).
		SetIsDeleted(true).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

// ───────────────────────── 投放目标解析 ─────────────────────────

// ResolveUserIDs 把投放范围解析为具体的用户 ID 列表（已去重）。
func (r *notificationRepo) ResolveUserIDs(ctx context.Context, targetType string, targetIDs []string) ([]string, error) {
	switch targetType {
	case biz.NotifyTargetUser:
		if len(targetIDs) == 0 {
			return nil, nil
		}
		return r.data.db.Admin.Query().
			Where(
				admin.IDIn(targetIDs...),
				admin.DeletedAtIsNil(),
				admin.StatusEQ(admin.Status(biz.StatusEnabled)),
			).
			IDs(ctx)

	case biz.NotifyTargetRole:
		if len(targetIDs) == 0 {
			return nil, nil
		}
		rows, err := r.data.db.AdminRole.Query().
			Where(adminrole.RoleIDIn(targetIDs...)).
			All(ctx)
		if err != nil {
			return nil, err
		}
		ids := make([]string, 0, len(rows))
		seen := make(map[string]struct{}, len(rows))
		for _, row := range rows {
			if _, dup := seen[row.AdminID]; dup {
				continue
			}
			seen[row.AdminID] = struct{}{}
			ids = append(ids, row.AdminID)
		}
		return ids, nil

	case biz.NotifyTargetAll:
		fallthrough
	default:
		return r.data.db.Admin.Query().
			Where(
				admin.DeletedAtIsNil(),
				admin.StatusEQ(admin.Status(biz.StatusEnabled)),
			).
			IDs(ctx)
	}
}

// ───────────────────────── 模型转换 ─────────────────────────

func toBizNotification(v *ent.Notification) *biz.Notification {
	if v == nil {
		return nil
	}
	return &biz.Notification{
		ID:          v.ID,
		Title:       v.Title,
		Content:     v.Content,
		Type:        string(v.Type),
		Level:       string(v.Level),
		BizType:     v.BizType,
		BizID:       v.BizID,
		SenderID:    v.SenderID,
		SenderName:  v.SenderName,
		Link:        v.Link,
		TargetType:  string(v.TargetType),
		TargetIDs:   v.TargetIds,
		Status:      string(v.Status),
		PublishedAt: v.PublishedAt,
		ExpireAt:    v.ExpireAt,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}

func toBizRecord(v *ent.NotificationRecord) *biz.NotificationRecord {
	if v == nil {
		return nil
	}
	rec := &biz.NotificationRecord{
		ID:             v.ID,
		NotificationID: v.NotificationID,
		UserID:         v.UserID,
		IsRead:         v.IsRead,
		ReadAt:         v.ReadAt,
		IsDeleted:      v.IsDeleted,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}
	if n := v.Edges.Notification; n != nil {
		rec.Notification = toBizNotification(n)
	}
	return rec
}
