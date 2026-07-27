package data

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/menu"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/role"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/rolemenu"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
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
		roleItem := &biz.Role{
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
		}

		// 填充菜单ID列表
		menuIDs, err := r.GetMenuIDsByRole(ctx, v.ID)
		if err != nil {
			// 记录日志但不影响主流程
			r.log.Warnf("获取角色 %s 的菜单失败: %v", v.ID, err)
			menuIDs = []string{}
		}
		roleItem.MenuIDs = menuIDs

		data = append(data, roleItem)
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
			return nil, errors.NotFound("ROLE_NOT_FOUND", "角色不存在")
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

// ====== 菜单关联方法实现 ======

// BindMenusForRole 为角色绑定菜单（全量替换：先删除所有现有绑定，再批量插入新绑定）
func (r *roleRepo) BindMenusForRole(ctx context.Context, roleID string, menuIDs []string) error {
	// 开启事务
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return err
	}

	// 1. 删除该角色的所有现有绑定
	_, err = tx.RoleMenu.Delete().
		Where(rolemenu.RoleIDEQ(roleID)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		r.log.Errorf("删除角色 %s 的旧菜单绑定失败: %v", roleID, err)
		return err
	}

	// 2. 批量插入新的绑定（如果有）
	if len(menuIDs) > 0 {
		builders := make([]*ent.RoleMenuCreate, len(menuIDs))
		for i, menuID := range menuIDs {
			builders[i] = tx.RoleMenu.Create().
				SetID(uuid.GenerateXID()).
				SetRoleID(roleID).
				SetMenuID(menuID)
		}

		err = tx.RoleMenu.CreateBulk(builders...).Exec(ctx)
		if err != nil {
			tx.Rollback()
			r.log.Errorf("为角色 %s 绑定菜单失败: %v", roleID, err)
			return err
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		r.log.Errorf("提交角色菜单绑定事务失败: %v", err)
		return err
	}

	r.log.Infof("角色 %s 绑定了 %d 个菜单", roleID, len(menuIDs))
	return nil
}

// GetMenuIDsByRole 获取角色关联的菜单ID列表
func (r *roleRepo) GetMenuIDsByRole(ctx context.Context, roleID string) ([]string, error) {
	// 查询关联表
	menus, err := r.data.db.RoleMenu.Query().
		Where(rolemenu.RoleIDEQ(roleID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 提取菜单ID
	ids := make([]string, len(menus))
	for i, m := range menus {
		ids[i] = m.MenuID
	}

	return ids, nil
}

// GetMenusByRole 获取角色关联的完整菜单列表（用于返回树形结构）
func (r *roleRepo) GetMenusByRole(ctx context.Context, roleID string) ([]*biz.Menu, error) {
	// 通过 Ent Edge 查询关联的菜单（Eager Loading）
	roleData, err := r.data.db.Role.Query().
		Where(role.ID(roleID)).
		WithMenus(
			func(mq *ent.MenuQuery) {
				mq.Order(ent.Asc(menu.FieldWeight)) // 按权重排序
			},
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errors.NotFound("ROLE_NOT_FOUND", "角色不存在")
		}
		return nil, err
	}

	// 转换为 biz.Menu 结构
	bizMenus := make([]*biz.Menu, len(roleData.Edges.Menus))
	for i, m := range roleData.Edges.Menus {
		bizMenus[i] = &biz.Menu{
			ID:           m.ID,
			ParentID:     m.ParentID,
			Name:         m.Name,
			Code:         m.Code,
			Title:        m.Title,
			Remark:       m.Remark,
			Path:         m.Path,
			Icon:         m.Icon,
			Type:         string(m.Type),
			LinkUrl:      m.URL,
			Component:    m.Component,
			AuthCode:     m.AuthCode,
			BadgeType:    m.BadgeType,
			Badge:        m.Badge,
			BadgeVariants: m.BadgeVariants,
			Weight:       m.Weight,
			Status:       string(m.Status),
			CreatedAt:    m.CreatedAt,
			UpdatedAt:    m.UpdatedAt,
		}
	}

	return bizMenus, nil
}

func (r *roleRepo) GetCodesByIds(ctx context.Context, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return []string{}, nil
	}
	codes, err := r.data.db.Role.
		Query().
		Where(role.IDIn(ids...)).
		Select(role.FieldCode).
		Strings(ctx)

	if err != nil {
		return nil, err
	}

	return codes, nil
}
