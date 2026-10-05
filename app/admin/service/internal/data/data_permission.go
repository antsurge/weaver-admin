package data

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/datapermission"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.DataPermissionRepo = (*dataPermissionRepo)(nil)

type dataPermissionRepo struct {
	data *Data
	log  *log.Helper
}

func NewDataPermissionRepo(data *Data, logger log.Logger) biz.DataPermissionRepo {
	return &dataPermissionRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// 列表（支持分页）
func (r *dataPermissionRepo) ListDataPermission(ctx context.Context, params *biz.ListDataPermissionRequest, opts ...*biz.ListDataPermissionOption) (*biz.ListDataPermissionResponse, error) {
	query := r.data.db.DataPermission.Query().
		Order(ent.Desc(datapermission.FieldCreatedAt))

	opt := &biz.ListDataPermissionOption{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	if opt.OnlyDeleted {
		query = query.Where(datapermission.DeletedAtNotNil())
	} else if !opt.IncludeDeleted {
		query = query.Where(datapermission.DeletedAtIsNil())
	}

	if v := params.Name; len(v) > 0 {
		query = query.Where(datapermission.NameContains(v))
	}
	if v := params.Code; len(v) > 0 {
		query = query.Where(datapermission.CodeContains(v))
	}
	if v := params.ScopeType; len(v) > 0 {
		query = query.Where(datapermission.ScopeTypeEQ(v))
	}
	if v := params.Status; len(v) > 0 {
		query = query.Where(datapermission.StatusEQ(datapermission.Status(v)))
	}

	res, err := enthelper.Pagination[*ent.DataPermission, *ent.DataPermissionQuery](ctx, query, params.PaginationParams)
	if err != nil {
		return nil, err
	}

	data := make([]*biz.DataPermission, 0, res.Total)
	for _, v := range res.Data {
		data = append(data, toBizDataPermission(v))
	}

	return &biz.ListDataPermissionResponse{
		Data:  data,
		Total: res.Total,
	}, nil
}

// 创建
func (r *dataPermissionRepo) CreateDataPermission(ctx context.Context, d *biz.DataPermission) error {
	_, err := r.data.db.DataPermission.Create().
		SetID(d.ID).
		SetName(d.Name).
		SetCode(d.Code).
		SetScopeType(d.ScopeType).
		SetNillableDeptIds(&d.DeptIds).
		SetNillableRoleIds(&d.RoleIds).
		SetStatus(datapermission.Status(d.Status)).
		SetNillableRemark(&d.Remark).
		SetCreatedAt(d.CreatedAt).
		SetUpdatedAt(d.UpdatedAt).
		Save(ctx)
	return err
}

// 编辑
func (r *dataPermissionRepo) UpdateDataPermission(ctx context.Context, d *biz.DataPermission) error {
	_, err := r.data.db.DataPermission.UpdateOneID(d.ID).
		SetName(d.Name).
		SetCode(d.Code).
		SetScopeType(d.ScopeType).
		SetNillableDeptIds(&d.DeptIds).
		SetNillableRoleIds(&d.RoleIds).
		SetStatus(datapermission.Status(d.Status)).
		SetNillableRemark(&d.Remark).
		SetUpdatedAt(d.UpdatedAt).
		Save(ctx)
	return err
}

// 删除（软删除）
func (r *dataPermissionRepo) DeleteDataPermission(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now()
	return r.data.db.DataPermission.
		Update().
		Where(datapermission.IDIn(ids...)).
		SetDeletedAt(now).
		Exec(ctx)
}

func (r *dataPermissionRepo) UpdateDataPermissionStatus(ctx context.Context, id, status string) error {
	_, err := r.data.db.DataPermission.UpdateOneID(id).
		SetStatus(datapermission.Status(status)).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}

// ListEnabledDataPermissionsByIDs 按ID批量查询启用的数据权限规则
func (r *dataPermissionRepo) ListEnabledDataPermissionsByIDs(ctx context.Context, ids []string) ([]*biz.DataPermission, error) {
	if len(ids) == 0 {
		return []*biz.DataPermission{}, nil
	}
	rows, err := r.data.db.DataPermission.Query().
		Where(
			datapermission.IDIn(ids...),
			datapermission.StatusEQ(datapermission.StatusEnabled),
			datapermission.DeletedAtIsNil(),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*biz.DataPermission, 0, len(rows))
	for _, v := range rows {
		list = append(list, toBizDataPermission(v))
	}
	return list, nil
}

// IsDataPermissionCodeExists 判断规则编码是否存在；excludeID 非空时排除该记录（编辑场景）
func (r *dataPermissionRepo) IsDataPermissionCodeExists(ctx context.Context, code, excludeID string) (bool, error) {
	query := r.data.db.DataPermission.Query().
		Where(
			datapermission.CodeEQ(code),
			datapermission.DeletedAtIsNil(),
		)

	if excludeID != "" {
		query = query.Where(datapermission.IDNEQ(excludeID))
	}

	exists, err := query.Exist(ctx)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func toBizDataPermission(v *ent.DataPermission) *biz.DataPermission {
	return &biz.DataPermission{
		ID:        v.ID,
		Name:      v.Name,
		Code:      v.Code,
		ScopeType: v.ScopeType,
		DeptIds:   v.DeptIds,
		RoleIds:   v.RoleIds,
		Status:    string(v.Status),
		Remark:    v.Remark,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
		DeletedAt: v.DeletedAt,
	}
}
