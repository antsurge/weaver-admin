package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	permissionV1 "github.com/hypercoze/kratos-admin/api/gen/go/permission/service/v1"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/biz"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/data/ent"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/data/ent/permission"
)

type permissionRepo struct {
	data *Data
	log  *log.Helper
}

func NewPermissionRepo(data *Data, logger log.Logger) biz.PermissionRepo {
	return &permissionRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *permissionRepo) CreatePermission(ctx context.Context, p *biz.Permission) error {
	_, err := r.data.db.Permission.Create().
		SetID(p.ID).
		SetParentID(p.ParentId).
		SetName(p.Name).
		SetCode(p.Code).
		SetNillableDescription(&p.Description).
		SetPath(p.Path).
		SetIcon(p.Icon).
		SetType(permission.Type(p.Type)).
		//SetURL(p.URL).
		SetComponent(p.Component).
		SetWeight(p.Weight).
		SetStatus(permission.Status(p.Status)).
		SetCreatedAt(p.CreatedAt).
		SetUpdatedAt(p.UpdatedAt).
		Save(ctx)
	return err
}

func (r *permissionRepo) UpdatePermission(ctx context.Context, p *biz.Permission) error {
	_, err := r.data.db.Permission.UpdateOneID(p.ID).
		SetParentID(p.ParentId).
		SetName(p.Name).
		SetCode(p.Code).
		SetNillableDescription(&p.Description).
		SetPath(p.Path).
		SetIcon(p.Icon).
		SetType(permission.Type(p.Type)).
		//SetURL(p.URL).
		SetComponent(p.Component).
		SetWeight(p.Weight).
		SetStatus(permission.Status(p.Status)).
		SetUpdatedAt(p.UpdatedAt).
		Save(ctx)
	return err
}

// DeletePermissions 删除权限及其子孙
func (r *permissionRepo) DeletePermission(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	var allIDs []string

	// 递归收集子孙节点
	var collect func(ids []string) error
	collect = func(ids []string) error {
		if len(ids) == 0 {
			return nil
		}

		children, err := r.data.db.Permission.
			Query().
			Where(permission.ParentIDIn(ids...)).
			All(ctx)
		if err != nil {
			return err
		}

		// 收集子节点 ID
		childIDs := make([]string, len(children))
		for i, c := range children {
			childIDs[i] = c.ID
		}
		allIDs = append(allIDs, childIDs...)

		// 递归查子节点
		return collect(childIDs)
	}

	// 初始化 allIDs
	allIDs = append(allIDs, ids...)
	if err := collect(ids); err != nil {
		return err
	}

	// 删除所有节点
	_, err := r.data.db.Permission.
		Delete().
		Where(permission.IDIn(allIDs...)).
		Exec(ctx)
	return err
}

func (r *permissionRepo) ListPermission(
	ctx context.Context,
	req *permissionV1.ListPermissionRequest,
) ([]*biz.Permission, error) {
	list, err := r.data.db.Permission.
		Query().
		Order(ent.Asc(permission.FieldWeight)).
		All(ctx)

	if err != nil {
		return nil, err
	}

	res := make([]*biz.Permission, 0, len(list))

	for _, v := range list {
		res = append(res, &biz.Permission{
			ID:          v.ID,
			ParentId:    v.ParentID,
			Name:        v.Name,
			Code:        v.Code,
			Description: v.Description,
			Path:        v.Path,
			Icon:        v.Icon,
			Type:        string(v.Type),
			Url:         v.URL,
			Component:   v.Component,
			Weight:      v.Weight,
			Status:      string(v.Status),
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		})
	}

	return res, nil
}

func (r *permissionRepo) UpdatePermissionStatus(ctx context.Context, id, status string) error {
	_, err := r.data.db.Permission.UpdateOneID(id).
		SetStatus(permission.Status(status)).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	return err
}
