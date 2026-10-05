package data

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/loginlog"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/operationlog"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/tenant"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.LogRepo = (*logRepo)(nil)

type logRepo struct {
	data *Data
	log  *log.Helper
}

func NewLogRepo(data *Data, logger log.Logger) biz.LogRepo {
	return &logRepo{data: data, log: log.NewHelper(logger)}
}

// ───────────────────────── 操作日志 ─────────────────────────

func (r *logRepo) CreateOperationLog(ctx context.Context, item *biz.OperationLog) error {
	return r.CreateOperationLogBulk(ctx, []*biz.OperationLog{item})
}

func (r *logRepo) CreateOperationLogBulk(ctx context.Context, items []*biz.OperationLog) error {
	if len(items) == 0 {
		return nil
	}
	builders := make([]*ent.OperationLogCreate, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		if item.ID == "" {
			item.ID = uuid.GenerateXID()
		}
		if item.TenantID == "" {
			item.TenantID = tenant.DefaultTenantID
		}
		if item.CreatedAt.IsZero() {
			item.CreatedAt = time.Now()
		}
		create := r.data.db.OperationLog.Create().
			SetID(item.ID).
			SetTenantID(item.TenantID).
			SetAdminID(item.AdminID).
			SetAdminName(item.AdminName).
			SetModule(item.Module).
			SetOperation(item.Operation).
			SetMethod(item.Method).
			SetPath(item.Path).
			SetSummary(item.Summary).
			SetIP(item.IP).
			SetUserAgent(item.UserAgent).
			SetParams(item.Params).
			SetStatus(operationlog.Status(item.Status)).
			SetCode(item.Code).
			SetMessage(item.Message).
			SetCostMs(int(item.CostMS)).
			SetCreatedAt(item.CreatedAt)
		builders = append(builders, create)
	}
	// 操作日志由异步写入器写入，调用方 ctx 无租户上下文，
	// 租户已在 item.TenantID 中显式指定，故此处使用 Unscoped 跳过拦截器填充。
	return r.data.db.OperationLog.CreateBulk(builders...).Exec(tenant.Unscoped(ctx))
}

func (r *logRepo) ListOperationLog(ctx context.Context, req *biz.ListOperationLogRequest) (*biz.ListOperationLogResponse, error) {
	query := r.data.db.OperationLog.Query().
		Order(ent.Desc(operationlog.FieldCreatedAt))

	if v := req.AdminID; v != "" {
		query = query.Where(operationlog.AdminIDEQ(v))
	}
	if v := req.AdminName; v != "" {
		query = query.Where(operationlog.AdminNameContains(v))
	}
	if v := req.Module; v != "" {
		query = query.Where(operationlog.ModuleEQ(v))
	}
	if v := req.Operation; v != "" {
		query = query.Where(operationlog.OperationEQ(v))
	}
	if v := req.Status; v != "" {
		query = query.Where(operationlog.StatusEQ(operationlog.Status(v)))
	}
	if v := req.Keyword; v != "" {
		query = query.Where(operationlog.Or(
			operationlog.PathContains(v),
			operationlog.SummaryContains(v),
			operationlog.OperationContains(v),
		))
	}
	if req.StartTime != nil {
		query = query.Where(operationlog.CreatedAtGTE(*req.StartTime))
	}
	if req.EndTime != nil {
		query = query.Where(operationlog.CreatedAtLTE(*req.EndTime))
	}

	res, err := enthelper.Pagination[*ent.OperationLog, *ent.OperationLogQuery](ctx, query, req.PaginationParams)
	if err != nil {
		return nil, err
	}

	data := make([]*biz.OperationLog, 0, res.Total)
	for _, v := range res.Data {
		data = append(data, toBizOperationLog(v))
	}
	return &biz.ListOperationLogResponse{Data: data, Total: res.Total}, nil
}

func (r *logRepo) GetOperationLog(ctx context.Context, id string) (*biz.OperationLog, error) {
	v, err := r.data.db.OperationLog.Query().Where(operationlog.IDEQ(id)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return toBizOperationLog(v), nil
}

func (r *logRepo) DeleteOperationLog(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.data.db.OperationLog.Delete().Where(operationlog.IDIn(ids...)).Exec(ctx)
	return err
}

func (r *logRepo) ClearOperationLog(ctx context.Context) error {
	_, err := r.data.db.OperationLog.Delete().Exec(ctx)
	return err
}

// ───────────────────────── 登录日志 ─────────────────────────

func (r *logRepo) CreateLoginLog(ctx context.Context, item *biz.LoginLog) error {
	if item == nil {
		return nil
	}
	if item.ID == "" {
		item.ID = uuid.GenerateXID()
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}
	return r.data.db.LoginLog.Create().
		SetID(item.ID).
		SetAdminID(item.AdminID).
		SetUsername(item.Username).
		SetIP(item.IP).
		SetUserAgent(item.UserAgent).
		SetStatus(loginlog.Status(item.Status)).
		SetReason(item.Reason).
		SetCreatedAt(item.CreatedAt).
		Exec(ctx)
}

func (r *logRepo) ListLoginLog(ctx context.Context, req *biz.ListLoginLogRequest) (*biz.ListLoginLogResponse, error) {
	query := r.data.db.LoginLog.Query().
		Order(ent.Desc(loginlog.FieldCreatedAt))

	if v := req.Username; v != "" {
		query = query.Where(loginlog.UsernameContains(v))
	}
	if v := req.Status; v != "" {
		query = query.Where(loginlog.StatusEQ(loginlog.Status(v)))
	}
	if v := req.IP; v != "" {
		query = query.Where(loginlog.IPContains(v))
	}
	if req.StartTime != nil {
		query = query.Where(loginlog.CreatedAtGTE(*req.StartTime))
	}
	if req.EndTime != nil {
		query = query.Where(loginlog.CreatedAtLTE(*req.EndTime))
	}

	res, err := enthelper.Pagination[*ent.LoginLog, *ent.LoginLogQuery](ctx, query, req.PaginationParams)
	if err != nil {
		return nil, err
	}

	data := make([]*biz.LoginLog, 0, res.Total)
	for _, v := range res.Data {
		data = append(data, toBizLoginLog(v))
	}
	return &biz.ListLoginLogResponse{Data: data, Total: res.Total}, nil
}

func (r *logRepo) DeleteLoginLog(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.data.db.LoginLog.Delete().Where(loginlog.IDIn(ids...)).Exec(ctx)
	return err
}

func (r *logRepo) ClearLoginLog(ctx context.Context) error {
	_, err := r.data.db.LoginLog.Delete().Exec(ctx)
	return err
}

// ───────────────────────── 模型转换 ─────────────────────────

func toBizOperationLog(v *ent.OperationLog) *biz.OperationLog {
	if v == nil {
		return nil
	}
	return &biz.OperationLog{
		ID:        v.ID,
		TenantID:  v.TenantID,
		AdminID:   v.AdminID,
		AdminName: v.AdminName,
		Module:    v.Module,
		Operation: v.Operation,
		Method:    v.Method,
		Path:      v.Path,
		Summary:   v.Summary,
		IP:        v.IP,
		UserAgent: v.UserAgent,
		Params:    v.Params,
		Status:    string(v.Status),
		Code:      v.Code,
		Message:   v.Message,
		CostMS:    int64(v.CostMs),
		CreatedAt: v.CreatedAt,
	}
}

func toBizLoginLog(v *ent.LoginLog) *biz.LoginLog {
	if v == nil {
		return nil
	}
	return &biz.LoginLog{
		ID:        v.ID,
		AdminID:   v.AdminID,
		Username:  v.Username,
		IP:        v.IP,
		UserAgent: v.UserAgent,
		Status:    string(v.Status),
		Reason:    v.Reason,
		CreatedAt: v.CreatedAt,
	}
}
