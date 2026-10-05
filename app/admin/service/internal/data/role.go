package data

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/adminrole"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/datapermission"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/menu"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/role"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/roledatapermission"
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

	// 转换为biz结构，并批量填充菜单ID列表（一次查询，避免 N+1）
	roleIDs := make([]string, 0, len(res.Data))
	for _, v := range res.Data {
		roleIDs = append(roleIDs, v.ID)
	}
	menuIDsByRole, err := r.getMenuIDsByRoles(ctx, roleIDs)
	if err != nil {
		r.log.Warnf("批量获取角色菜单绑定失败: %v", err)
		menuIDsByRole = map[string][]string{}
	}

	data := make([]*biz.Role, 0, res.Total)
	for _, v := range res.Data {
		roleItem := &biz.Role{
			ID:        v.ID,
			Name:      v.Name,
			Code:      v.Code,
			Remark:    v.Remark,
			Weight:    v.Weight,
			Status:    string(v.Status),
			IsSystem:  v.IsSystem,
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
			MenuIDs:   menuIDsByRole[v.ID],
		}
		if roleItem.MenuIDs == nil {
			roleItem.MenuIDs = []string{}
		}

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
		SetIsSystem(data.IsSystem).
		SetUpdatedAt(data.UpdatedAt).
		Save(ctx)

	return err
}

// DeleteRole 批量软删除角色（事务内：有用户绑定时拒绝删除，并清理菜单绑定）
func (r *roleRepo) DeleteRole(ctx context.Context, ids []string) error {
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return err
	}

	// 1. 有用户绑定的角色禁止删除（需先解绑用户）
	bindings, err := tx.AdminRole.Query().
		Where(adminrole.RoleIDIn(ids...)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}
	if len(bindings) > 0 {
		tx.Rollback()
		return errors.BadRequest("ROLE_HAS_ADMINS", "角色已分配给用户，请先解除用户绑定")
	}

	// 2. 清理角色-菜单绑定
	_, err = tx.RoleMenu.Delete().
		Where(rolemenu.RoleIDIn(ids...)).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		r.log.Errorf("清理角色菜单绑定失败: %v", err)
		return err
	}

	// 3. 软删除角色
	err = tx.Role.Update().
		Where(role.IDIn(ids...)).
		SetDeletedAt(time.Now()).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
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

// getMenuIDsByRoles 批量查询多个角色的菜单ID（一次查询，按 roleID 分组）
func (r *roleRepo) getMenuIDsByRoles(ctx context.Context, roleIDs []string) (map[string][]string, error) {
	if len(roleIDs) == 0 {
		return map[string][]string{}, nil
	}
	bindings, err := r.data.db.RoleMenu.Query().
		Where(rolemenu.RoleIDIn(roleIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[string][]string, len(roleIDs))
	for _, b := range bindings {
		m[b.RoleID] = append(m[b.RoleID], b.MenuID)
	}
	return m, nil
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
			ID:            m.ID,
			ParentID:      m.ParentID,
			Name:          m.Name,
			Code:          m.Code,
			Title:         m.Title,
			Remark:        m.Remark,
			Path:          m.Path,
			Icon:          m.Icon,
			Type:          string(m.Type),
			LinkUrl:       m.URL,
			Component:     m.Component,
			AuthCode:      m.AuthCode,
			BadgeType:     m.BadgeType,
			Badge:         m.Badge,
			BadgeVariants: m.BadgeVariants,
			Weight:        m.Weight,
			Status:        string(m.Status),
			CreatedAt:     m.CreatedAt,
			UpdatedAt:     m.UpdatedAt,
		}
	}

	return bizMenus, nil
}

func (r *roleRepo) IsRoleCodeExists(ctx context.Context, code string, excludeID string) (bool, error) {
	query := r.data.db.Role.Query().
		Where(role.CodeEQ(code))

	// 如果是编辑，排除当前记录
	if excludeID != "" {
		query = query.Where(role.IDNEQ(excludeID))
	}

	// 判断是否存在
	return query.Exist(ctx)
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

// GetNamesByIds 批量查询角色名称（个人中心展示用，空 ids 返回空列表）
func (r *roleRepo) GetNamesByIds(ctx context.Context, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return []string{}, nil
	}
	names, err := r.data.db.Role.
		Query().
		Where(role.IDIn(ids...)).
		Select(role.FieldName).
		Strings(ctx)

	if err != nil {
		return nil, err
	}

	return names, nil
}

// ====== 数据权限规则关联实现 ======

// BindDataPermissionsForRole 为角色绑定数据权限规则（全量替换：事务内删除旧绑定 + 批量插入新绑定）
func (r *roleRepo) BindDataPermissionsForRole(ctx context.Context, roleID string, dataPermissionIDs []string) error {
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return err
	}

	// 1. 删除该角色现有绑定
	if _, err := tx.RoleDataPermission.Delete().
		Where(roledatapermission.RoleID(roleID)).
		Exec(ctx); err != nil {
		tx.Rollback()
		return err
	}

	// 2. 批量插入新绑定
	for _, dpID := range dataPermissionIDs {
		if dpID == "" {
			continue
		}
		if _, err := tx.RoleDataPermission.Create().
			SetID(uuid.GenerateXID()).
			SetRoleID(roleID).
			SetDataPermissionID(dpID).
			Save(ctx); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// GetDataPermissionIDsByRole 获取角色绑定的数据权限规则ID列表
func (r *roleRepo) GetDataPermissionIDsByRole(ctx context.Context, roleID string) ([]string, error) {
	ids, err := r.data.db.RoleDataPermission.Query().
		Where(roledatapermission.RoleID(roleID)).
		Select(roledatapermission.FieldDataPermissionID).
		Strings(ctx)
	if err != nil {
		return nil, err
	}
	if ids == nil {
		ids = []string{}
	}
	return ids, nil
}

// GetDataPermissionsByRole 获取角色绑定的数据权限规则列表
func (r *roleRepo) GetDataPermissionsByRole(ctx context.Context, roleID string) ([]*biz.DataPermission, error) {
	dpIDs, err := r.GetDataPermissionIDsByRole(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if len(dpIDs) == 0 {
		return []*biz.DataPermission{}, nil
	}
	rows, err := r.data.db.DataPermission.Query().
		Where(datapermission.IDIn(dpIDs...), datapermission.DeletedAtIsNil()).
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

// GetDataPermissionsByRoleIDs 批量查询多个角色绑定的数据权限规则（按角色ID分组）
func (r *roleRepo) GetDataPermissionsByRoleIDs(ctx context.Context, roleIDs []string) (map[string][]*biz.DataPermission, error) {
	result := make(map[string][]*biz.DataPermission)
	if len(roleIDs) == 0 {
		return result, nil
	}

	// 查询关联表的角色ID → 规则ID 映射
	rows, err := r.data.db.RoleDataPermission.Query().
		Where(roledatapermission.RoleIDIn(roleIDs...)).
		Select(roledatapermission.FieldRoleID, roledatapermission.FieldDataPermissionID).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 收集规则ID并去重
	dpIDMap := make(map[string]struct{})
	roleDPIDs := make(map[string][]string, len(roleIDs))
	for _, row := range rows {
		dpIDMap[row.DataPermissionID] = struct{}{}
		roleDPIDs[row.RoleID] = append(roleDPIDs[row.RoleID], row.DataPermissionID)
	}

	// 批量加载规则详情
	dpDetails := []*biz.DataPermission{}
	if len(dpIDMap) > 0 {
		dpIDs := make([]string, 0, len(dpIDMap))
		for id := range dpIDMap {
			dpIDs = append(dpIDs, id)
		}
		dpRows, err := r.data.db.DataPermission.Query().
			Where(datapermission.IDIn(dpIDs...), datapermission.DeletedAtIsNil()).
			All(ctx)
		if err != nil {
			return nil, err
		}
		for _, v := range dpRows {
			dpDetails = append(dpDetails, toBizDataPermission(v))
		}
	}

	// 按角色ID分组返回（保留完整规则信息）
	detailByID := make(map[string]*biz.DataPermission, len(dpDetails))
	for _, dp := range dpDetails {
		detailByID[dp.ID] = dp
	}
	for roleID, ids := range roleDPIDs {
		var perms []*biz.DataPermission
		for _, id := range ids {
			if dp, ok := detailByID[id]; ok {
				perms = append(perms, dp)
			}
		}
		result[roleID] = perms
	}

	return result, nil
}
