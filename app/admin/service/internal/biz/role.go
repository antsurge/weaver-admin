package biz

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/conf"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/tenant"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

// tenantIDFromCtx 从上下文中读取当前租户 ID（未设置返回空串，配额校验跳过）。
func tenantIDFromCtx(ctx context.Context) string {
	tid, _ := tenant.From(ctx)
	return tid
}

type Role struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Code              string    `json:"code"`
	Remark            string    `json:"remark"`
	Weight            int       `json:"weight"`
	Status            string    `json:"status"`
	IsSystem          bool      `json:"isSystem"`
	IsSuperAdmin      bool      `json:"isSuperAdmin"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	MenuIDs           []string  `json:"menuIds"`
	DataPermissionIDs []string  `json:"dataPermissionIds"`
}

type AdminRole struct{}

type ListRoleRequest struct {
	enthelper.PaginationParams
	Name   string `form:"name" query:"name"`
	Code   string `form:"code" query:"code"`
	Status string `form:"status" query:"status"`
}

type ListRoleResponse struct {
	Items []*Role `json:"items"`
	Total int     `json:"total"`
}

type ListRoleOption struct {
	enthelper.QueryOption
	Name   string `form:"name" query:"name"`
	Code   string `form:"code" query:"code"`
	Status string `form:"status" query:"status"`
}

type RoleRepo interface {
	ListRole(ctx context.Context, params *ListRoleRequest, opts ...*ListRoleOption) (*ListRoleResponse, error)
	GetRole(ctx context.Context, id string) (*Role, error)
	CreateRole(ctx context.Context, role *Role) error
	UpdateRole(ctx context.Context, role *Role) error
	UpdateRoleStatus(ctx context.Context, id string, status string) error
	DeleteRole(ctx context.Context, ids []string) error

	// ====== 菜单关联方法 ======

	// BindMenusForRole 为角色绑定菜单（全量替换：先删除所有现有绑定，再批量插入新绑定）
	BindMenusForRole(ctx context.Context, roleID string, menuIDs []string) error

	// GetMenuIDsByRole 获取角色关联的菜单ID列表
	GetMenuIDsByRole(ctx context.Context, roleID string) ([]string, error)

	// GetMenusByRole 获取角色关联的完整菜单列表（用于返回树形结构）
	GetMenusByRole(ctx context.Context, roleID string) ([]*Menu, error)

	GetCodesByIds(ctx context.Context, ids []string) ([]string, error)

	// GetNamesByIds 批量查询角色名称（按 ids 顺序返回，个人中心展示用）
	GetNamesByIds(ctx context.Context, ids []string) ([]string, error)

	// ====== 数据权限规则关联方法 ======

	// BindDataPermissionsForRole 为角色绑定数据权限规则（全量替换：先删除现有绑定，再批量插入新绑定）
	BindDataPermissionsForRole(ctx context.Context, roleID string, dataPermissionIDs []string) error

	// GetDataPermissionIDsByRole 获取角色绑定的数据权限规则ID列表
	GetDataPermissionIDsByRole(ctx context.Context, roleID string) ([]string, error)

	// GetDataPermissionsByRole 获取角色绑定的数据权限规则列表
	GetDataPermissionsByRole(ctx context.Context, roleID string) ([]*DataPermission, error)

	// GetDataPermissionsByRoleIDs 批量查询多个角色绑定的数据权限规则（用于 current-user 数据权限解析，按角色ID分组）
	GetDataPermissionsByRoleIDs(ctx context.Context, roleIDs []string) (map[string][]*DataPermission, error)

	// IsRoleCodeExists 判断角色编码是否存在（excludeID 非空时排除该角色，用于编辑场景）
	IsRoleCodeExists(ctx context.Context, code string, excludeID string) (bool, error)
}

type AdminRoleRepo interface {
	GetRoleIdsByAdminId(ctx context.Context, id string) ([]string, error)
}

type RoleMenuRepo interface {
	GetMenuIdsByRoleIds(ctx context.Context, roleIds []string) ([]string, error)
}

type RoleUsecase struct {
	repo          RoleRepo
	menuRepo      MenuRepo
	tenantUsecase *TenantUsecase
	appConf       *conf.App
	log           *log.Helper
}

func NewRoleUsecase(repo RoleRepo, menuRepo MenuRepo, tenantUsecase *TenantUsecase, appConf *conf.App, logger log.Logger) *RoleUsecase {
	return &RoleUsecase{repo: repo, menuRepo: menuRepo, tenantUsecase: tenantUsecase, appConf: appConf, log: log.NewHelper(logger)}
}

func (uc *RoleUsecase) ListRole(ctx context.Context, params *ListRoleRequest) (*ListRoleResponse, error) {
	res, err := uc.repo.ListRole(ctx, params)
	if err != nil {
		return nil, err
	}
	if res != nil {
		for _, role := range res.Items {
			role.IsSuperAdmin = uc.isSuperAdminRole(role)
		}
	}
	return res, nil
}

func (uc *RoleUsecase) GetRole(ctx context.Context, id string) (*Role, error) {
	role, err := uc.repo.GetRole(ctx, id)
	if err != nil {
		return nil, err
	}
	if role != nil {
		role.IsSuperAdmin = uc.isSuperAdminRole(role)
	}
	return role, nil
}

// GetRoleWithMenus 获取角色详情（包含菜单ID列表与数据权限规则ID列表）
func (uc *RoleUsecase) GetRoleWithMenus(ctx context.Context, id string) (*Role, error) {
	role, err := uc.repo.GetRole(ctx, id)
	if err != nil {
		return nil, err
	}
	if role != nil {
		role.IsSuperAdmin = uc.isSuperAdminRole(role)
	}

	// 填充菜单ID列表
	menuIDs, err := uc.repo.GetMenuIDsByRole(ctx, id)
	if err != nil {
		// 记录日志但不返回错误，允许角色没有绑定菜单
		uc.log.Warnf("获取角色菜单失败: %v", err)
		menuIDs = []string{}
	}
	role.MenuIDs = menuIDs

	// 填充数据权限规则ID列表
	dataPermissionIDs, err := uc.repo.GetDataPermissionIDsByRole(ctx, id)
	if err != nil {
		// 记录日志但不返回错误，允许角色没有绑定数据权限规则
		uc.log.Warnf("获取角色数据权限规则失败: %v", err)
		dataPermissionIDs = []string{}
	}
	role.DataPermissionIDs = dataPermissionIDs

	return role, nil
}

func (uc *RoleUsecase) CreateRole(ctx context.Context, role *Role) (*Role, error) {
	// 租户角色数配额校验（max_roles>0 时强制）
	if uc.tenantUsecase != nil {
		if err := uc.tenantUsecase.CheckQuota(ctx, tenantIDFromCtx(ctx), "role"); err != nil {
			return nil, err
		}
	}

	// 唯一性校验（角色编码）
	if err := uc.validateUnique(ctx, role, ""); err != nil {
		return nil, err
	}

	now := time.Now()
	role.ID = uuid.GenerateXID()
	role.CreatedAt = now
	role.UpdatedAt = now

	err := uc.repo.CreateRole(ctx, role)
	return role, err
}

// CreateRoleWithMenus 创建角色并绑定菜单（事务操作）
func (uc *RoleUsecase) CreateRoleWithMenus(ctx context.Context, role *Role) (*Role, error) {
	// 租户角色数配额校验（max_roles>0 时强制）
	if uc.tenantUsecase != nil {
		if err := uc.tenantUsecase.CheckQuota(ctx, tenantIDFromCtx(ctx), "role"); err != nil {
			return nil, err
		}
	}

	// 唯一性校验（角色编码）
	if err := uc.validateUnique(ctx, role, ""); err != nil {
		return nil, err
	}

	now := time.Now()
	role.ID = uuid.GenerateXID()
	role.CreatedAt = now
	role.UpdatedAt = now

	// 创建角色
	if err := uc.repo.CreateRole(ctx, role); err != nil {
		return nil, err
	}

	// 绑定菜单（如果有）
	if len(role.MenuIDs) > 0 {
		if err := uc.repo.BindMenusForRole(ctx, role.ID, role.MenuIDs); err != nil {
			return nil, err
		}
	}

	return role, nil
}

// IsRoleCodeExists 判断角色编码是否存在（excludeID 非空时排除该角色，用于编辑场景）
func (uc *RoleUsecase) IsRoleCodeExists(ctx context.Context, code string, excludeID string) (bool, error) {
	return uc.repo.IsRoleCodeExists(ctx, code, excludeID)
}

// validateUnique 校验角色编码唯一性；excludeID 非空时排除该角色（编辑场景）
func (uc *RoleUsecase) validateUnique(ctx context.Context, role *Role, excludeID string) error {
	codeExists, err := uc.IsRoleCodeExists(ctx, role.Code, excludeID)
	if err != nil {
		return err
	}
	if codeExists {
		return errors.BadRequest("ROLE_CODE_EXISTS", "角色编码已存在")
	}
	return nil
}

// isSuperAdminRole 判断角色是否为超级管理员角色（按配置的 super_admin_code 匹配）
func (uc *RoleUsecase) isSuperAdminRole(role *Role) bool {
	return role != nil &&
		uc.appConf != nil &&
		uc.appConf.SuperAdminCode != "" &&
		role.Code == uc.appConf.SuperAdminCode
}

// ensureNotSystemRole 校验目标角色既不是系统内置角色，也不是超级管理员角色
// 系统内置角色（is_system）与超级管理员角色均不允许修改、修改状态以及重新绑定菜单
func (uc *RoleUsecase) ensureNotSystemRole(ctx context.Context, id string) error {
	role, err := uc.repo.GetRole(ctx, id)
	if err != nil {
		return err
	}
	if uc.isSuperAdminRole(role) {
		return errors.BadRequest("CANNOT_MODIFY_SUPER_ADMIN_ROLE", "超级管理员角色不允许修改")
	}
	if role != nil && role.IsSystem {
		return errors.BadRequest("CANNOT_MODIFY_SYSTEM_ROLE", "系统内置角色不允许修改")
	}
	return nil
}

func (uc *RoleUsecase) UpdateRole(ctx context.Context, role *Role) (*Role, error) {
	if err := uc.ensureNotSystemRole(ctx, role.ID); err != nil {
		return nil, err
	}

	// 唯一性校验（角色编码，编辑时排除自身）
	if err := uc.validateUnique(ctx, role, role.ID); err != nil {
		return nil, err
	}

	now := time.Now()
	role.UpdatedAt = now
	err := uc.repo.UpdateRole(ctx, role)

	return role, err
}

// UpdateRoleWithMenus 更新角色并重新绑定菜单（事务操作）
func (uc *RoleUsecase) UpdateRoleWithMenus(ctx context.Context, role *Role) (*Role, error) {
	if err := uc.ensureNotSystemRole(ctx, role.ID); err != nil {
		return nil, err
	}

	// 唯一性校验（角色编码，编辑时排除自身）
	if err := uc.validateUnique(ctx, role, role.ID); err != nil {
		return nil, err
	}

	now := time.Now()
	role.UpdatedAt = now

	// 更新角色基本信息
	if err := uc.repo.UpdateRole(ctx, role); err != nil {
		return nil, err
	}

	// 重新绑定菜单（全量替换）
	if err := uc.repo.BindMenusForRole(ctx, role.ID, role.MenuIDs); err != nil {
		return nil, err
	}

	return role, nil
}

func (uc *RoleUsecase) UpdateRoleStatus(ctx context.Context, id string, status string) error {
	if err := uc.ensureNotSystemRole(ctx, id); err != nil {
		return err
	}
	return uc.repo.UpdateRoleStatus(ctx, id, status)
}

// DeleteRole 删除角色（软删除）
// 超级管理员角色与系统内置角色（is_system）均禁止删除
func (uc *RoleUsecase) DeleteRole(ctx context.Context, ids []string) error {
	for _, id := range ids {
		role, err := uc.repo.GetRole(ctx, id)
		if err != nil {
			return err
		}
		if uc.isSuperAdminRole(role) {
			return errors.BadRequest("CANNOT_DELETE_SUPER_ADMIN_ROLE", "超级管理员角色不允许删除")
		}
		if role != nil && role.IsSystem {
			return errors.BadRequest("CANNOT_DELETE_SYSTEM_ROLE", "系统内置角色不允许删除")
		}
	}
	return uc.repo.DeleteRole(ctx, ids)
}

// BindMenusForRole 为角色绑定菜单（业务层校验）
func (uc *RoleUsecase) BindMenusForRole(ctx context.Context, roleID string, menuIDs []string) error {
	// 验证角色是否存在，且不是系统内置角色
	if err := uc.ensureNotSystemRole(ctx, roleID); err != nil {
		return err
	}

	// 执行绑定
	return uc.repo.BindMenusForRole(ctx, roleID, menuIDs)
}

// GetMenusByRole 获取角色的菜单树
func (uc *RoleUsecase) GetMenusByRole(ctx context.Context, roleID string) ([]*Menu, error) {
	// 先验证角色是否存在
	_, err := uc.repo.GetRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	// 查询关联的菜单
	menus, err := uc.repo.GetMenusByRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	// 构建树形结构
	tree := buildMenuTree(menus)
	return tree, nil
}

// BindDataPermissionsForRole 为角色绑定数据权限规则（业务层校验）
func (uc *RoleUsecase) BindDataPermissionsForRole(ctx context.Context, roleID string, dataPermissionIDs []string) error {
	// 验证角色是否存在，且不是系统内置角色
	if err := uc.ensureNotSystemRole(ctx, roleID); err != nil {
		return err
	}

	// 执行绑定（全量替换）
	return uc.repo.BindDataPermissionsForRole(ctx, roleID, dataPermissionIDs)
}

// GetDataPermissionsByRole 获取角色绑定的数据权限规则列表
func (uc *RoleUsecase) GetDataPermissionsByRole(ctx context.Context, roleID string) ([]*DataPermission, error) {
	// 先验证角色是否存在
	_, err := uc.repo.GetRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	return uc.repo.GetDataPermissionsByRole(ctx, roleID)
}
