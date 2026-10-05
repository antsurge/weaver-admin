package data

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	entadmin "github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/admin"
	entdepartment "github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/department"
	entdicttype "github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/dicttype"
	entnotification "github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/notification"
	entposition "github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/position"
	entrole "github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/role"
	entsysjob "github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/sysjob"
	enttenant "github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/tenant"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/tenant"
	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.TenantRepo = (*tenantRepo)(nil)

type tenantRepo struct {
	data *Data
	log  *log.Helper
}

func NewTenantRepo(data *Data, logger log.Logger) biz.TenantRepo {
	return &tenantRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *tenantRepo) GetByCode(ctx context.Context, code string) (*biz.Tenant, error) {
	t, err := r.data.db.Tenant.Query().
		Where(
			enttenant.CodeEQ(code),
			enttenant.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return toBizTenant(t), nil
}

func (r *tenantRepo) GetTenantByID(ctx context.Context, id string) (*biz.Tenant, error) {
	t, err := r.data.db.Tenant.Query().
		Where(
			enttenant.IDEQ(id),
			enttenant.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return toBizTenant(t), nil
}

// ListTenant 租户分页列表
func (r *tenantRepo) ListTenant(ctx context.Context, params *biz.ListTenantRequest, opts ...*biz.ListTenantOption) (*biz.ListTenantResponse, error) {
	query := r.data.db.Tenant.Query().
		Order(ent.Desc(enttenant.FieldCreatedAt))

	opt := &biz.ListTenantOption{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	if opt.OnlyDeleted {
		query = query.Where(enttenant.DeletedAtNotNil())
	} else if !opt.IncludeDeleted {
		query = query.Where(enttenant.DeletedAtIsNil())
	}

	if v := params.Name; len(v) > 0 {
		query = query.Where(enttenant.NameContains(v))
	}
	if v := params.Code; len(v) > 0 {
		query = query.Where(enttenant.CodeContains(v))
	}
	if v := params.Status; len(v) > 0 {
		query = query.Where(enttenant.StatusEQ(enttenant.Status(v)))
	}

	res, err := enthelper.Pagination[*ent.Tenant, *ent.TenantQuery](ctx, query, params.PaginationParams)
	if err != nil {
		return nil, err
	}

	data := make([]*biz.Tenant, 0, res.Total)
	for _, v := range res.Data {
		data = append(data, toBizTenant(v))
	}

	return &biz.ListTenantResponse{
		Data:  data,
		Total: res.Total,
	}, nil
}

// CreateTenant 创建租户
func (r *tenantRepo) CreateTenant(ctx context.Context, t *biz.Tenant) error {
	_, err := r.data.db.Tenant.Create().
		SetID(t.ID).
		SetCode(t.Code).
		SetName(t.Name).
		SetStatus(enttenant.Status(t.Status)).
		SetNillableExpireAt(t.ExpireAt).
		SetMaxUsers(t.MaxUsers).
		SetMaxRoles(t.MaxRoles).
		SetNillableContactName(&t.ContactName).
		SetNillableContactPhone(&t.ContactPhone).
		SetNillableRemark(&t.Remark).
		SetCreatedAt(t.CreatedAt).
		SetUpdatedAt(t.UpdatedAt).
		Save(ctx)
	return err
}

// UpdateTenant 更新租户（普通字段；code 为 Immutable 字段，ent 生成代码不含 SetCode，天然不可修改）
func (r *tenantRepo) UpdateTenant(ctx context.Context, t *biz.Tenant) error {
	_, err := r.data.db.Tenant.UpdateOneID(t.ID).
		SetName(t.Name).
		SetStatus(enttenant.Status(t.Status)).
		SetNillableExpireAt(t.ExpireAt).
		SetMaxUsers(t.MaxUsers).
		SetMaxRoles(t.MaxRoles).
		SetNillableContactName(&t.ContactName).
		SetNillableContactPhone(&t.ContactPhone).
		SetNillableRemark(&t.Remark).
		SetUpdatedAt(t.UpdatedAt).
		Save(ctx)
	return err
}

// DeleteTenant 批量软删除租户
func (r *tenantRepo) DeleteTenant(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now()
	return r.data.db.Tenant.
		Update().
		Where(enttenant.IDIn(ids...)).
		SetDeletedAt(now).
		Exec(ctx)
}

// UpdateTenantStatus 更新租户状态
func (r *tenantRepo) UpdateTenantStatus(ctx context.Context, id, status string) error {
	_, err := r.data.db.Tenant.UpdateOneID(id).
		SetStatus(enttenant.Status(status)).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

// ExistsTenantCode 判断租户编码是否存在（excludeID 非空时排除该租户）
func (r *tenantRepo) ExistsTenantCode(ctx context.Context, code, excludeID string) (bool, error) {
	query := r.data.db.Tenant.Query().
		Where(
			enttenant.CodeEQ(code),
			enttenant.DeletedAtIsNil(),
		)
	if excludeID != "" {
		query = query.Where(enttenant.IDNEQ(excludeID))
	}
	return query.Exist(ctx)
}

// GetUsage 统计租户下各资源当前用量。
// 各资源表均为租户级表，借助租户上下文（tenant.With）将查询限定在指定租户作用域。
func (r *tenantRepo) GetUsage(ctx context.Context, tenantID string) (*biz.TenantUsage, error) {
	tctx := tenant.With(ctx, tenantID)

	usage := &biz.TenantUsage{TenantID: tenantID}

	var err error

	usage.Admins, err = r.data.db.Admin.Query().
		Where(entadmin.DeletedAtIsNil()).
		Count(tctx)
	if err != nil {
		return nil, err
	}

	usage.Roles, err = r.data.db.Role.Query().
		Where(entrole.DeletedAtIsNil()).
		Count(tctx)
	if err != nil {
		return nil, err
	}

	usage.Menus, err = r.data.db.Menu.Query().
		Count(tctx)
	if err != nil {
		return nil, err
	}

	usage.Departments, err = r.data.db.Department.Query().
		Where(entdepartment.DeletedAtIsNil()).
		Count(tctx)
	if err != nil {
		return nil, err
	}

	usage.Positions, err = r.data.db.Position.Query().
		Where(entposition.DeletedAtIsNil()).
		Count(tctx)
	if err != nil {
		return nil, err
	}

	usage.DictTypes, err = r.data.db.DictType.Query().
		Where(entdicttype.DeletedAtIsNil()).
		Count(tctx)
	if err != nil {
		return nil, err
	}

	usage.Jobs, err = r.data.db.SysJob.Query().
		Where(entsysjob.DeletedAtIsNil()).
		Count(tctx)
	if err != nil {
		return nil, err
	}

	usage.Notifications, err = r.data.db.Notification.Query().
		Where(entnotification.DeletedAtIsNil()).
		Count(tctx)
	if err != nil {
		return nil, err
	}

	return usage, nil
}

func toBizTenant(t *ent.Tenant) *biz.Tenant {
	return &biz.Tenant{
		ID:           t.ID,
		Code:         t.Code,
		Name:         t.Name,
		Status:       string(t.Status),
		ExpireAt:     t.ExpireAt,
		MaxUsers:     t.MaxUsers,
		MaxRoles:     t.MaxRoles,
		ContactName:  t.ContactName,
		ContactPhone: t.ContactPhone,
		Remark:       t.Remark,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
		DeletedAt:    t.DeletedAt,
	}
}
