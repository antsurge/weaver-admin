package data

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/role"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

type roleRepo struct {
	data *Data
	log  *log.Helper
}

func NewRoleRepo(data *Data, logger log.Logger) biz.RoleRepo {
	return &roleRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *roleRepo) ListRole(ctx context.Context, params *biz.ListRoleRequest, opts ...*biz.ListRoleOption) (*biz.ListRoleResponse, error) {
	query := r.data.db.Role.Query().Order(ent.Desc(role.FieldCreatedAt))

	opt := &biz.ListRoleOption{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	if opt.OnlyDeleted {
		query = query.Where(role.DeletedAtNotNil())
	} else if !opt.IncludeDeleted {
		query = query.Where(role.DeletedAtIsNil())
	}

	// 名称
	if v := params.Name; len(v) > 0 {
		query = query.Where(role.NameContains(v))
	}

	// code
	if v := params.Code; len(v) > 0 {
		query = query.Where(role.CodeContains(v))
	}

	// 状态
	if v := params.Status; len(v) > 0 {
		query = query.Where(role.StatusEQ(role.Status(v)))
	}

	res, err := enthelper.Pagination[
		*ent.Role,
		*ent.RoleQuery,
	](ctx, query, params.PaginationParams)
	if err != nil {
		return nil, err
	}

	// 转换为biz结构
	data := make([]*biz.Role, 0, res.Total)
	for _, v := range res.Data {
		data = append(data, &biz.Role{
			ID:        v.ID,
			Name:      v.Name,
			Code:      v.Code,
			Remark:    v.Remark,
			Weight:    v.Weight,
			Status:    string(v.Status),
			DataScope: v.DataScope,
			IsSystem:  v.IsSystem,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
		})
	}

	return &biz.ListRoleResponse{
		Items: data,
		Total: res.Total,
	}, nil
}

func (r *roleRepo) GetRole(ctx context.Context, id string) (*biz.Role, error) {
	data, err := r.data.db.Role.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.NotFound("ROLE_NOT_FOUND", "岗位不存在")
		}
		return nil, err
	}
	role := &biz.Role{
		ID:        data.ID,
		Name:      data.Name,
		Code:      data.Code,
		Remark:    data.Remark,
		Weight:    data.Weight,
		Status:    string(data.Status),
		DataScope: data.DataScope,
		IsSystem:  data.IsSystem,
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.UpdatedAt,
	}

	return role, nil
}

func (r *roleRepo) CreateRole(ctx context.Context, data *biz.Role) error {
	_, err := r.data.db.Role.Create().
		SetID(data.ID).
		SetName(data.Name).
		SetCode(data.Code).
		SetWeight(data.Weight).
		SetStatus(role.Status(data.Status)).
		SetRemark(data.Remark).
		SetDataScope(data.DataScope).
		SetIsSystem(data.IsSystem).
		SetCreatedAt(data.CreatedAt).
		SetUpdatedAt(data.UpdatedAt).
		Save(ctx)
	return err
}

func (r *roleRepo) UpdateRole(ctx context.Context, data *biz.Role) error {
	_, err := r.data.db.Role.
		UpdateOneID(data.ID).
		SetName(data.Name).
		SetCode(data.Code).
		SetWeight(data.Weight).
		SetStatus(role.Status(data.Status)).
		SetRemark(data.Remark).
		SetDataScope(data.DataScope).
		SetIsSystem(data.IsSystem).
		SetCreatedAt(data.CreatedAt).
		SetUpdatedAt(data.UpdatedAt).
		Save(ctx)

	return err
}

func (r *roleRepo) DeleteRole(ctx context.Context, ids []string) error {
	err := r.data.db.Role.
		Update().
		Where(role.IDIn(ids...)).
		SetDeletedAt(time.Now()).
		Exec(ctx)

	return err
}

func (r *roleRepo) UpdateRoleStatus(ctx context.Context, id string, status string) error {
	_, err := r.data.db.Role.UpdateOneID(id).
		SetStatus(role.Status(status)).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}
