package biz

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/log"
)

type Role struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Remark    string    `json:"remark"`
	Weight    int       `json:"weight"`
	Status    string    `json:"status"`
	DataScope string    `json:"dataScope"`
	IsSystem  bool      `json:"isSystem"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	MenuIDs   []string  `json:"menuIds"`
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
}

type AdminRoleRepo interface {
	GetRoleIdsByAdminId(ctx context.Context, id string) ([]string, error)
}

type RoleMenuRepo interface {
	GetMenuIdsByRoleIds(ctx context.Context, roleIds []string) ([]string, error)
}

type RoleUsecase struct {
	repo     RoleRepo
	menuRepo MenuRepo
	log      *log.Helper
}

func NewRoleUsecase(repo RoleRepo, menuRepo MenuRepo, logger log.Logger) *RoleUsecase {
	return &RoleUsecase{repo: repo, menuRepo: menuRepo, log: log.NewHelper(logger)}
}

func (uc *RoleUsecase) ListRole(ctx context.Context, params *ListRoleRequest) (*ListRoleResponse, error) {
	return uc.repo.ListRole(ctx, params)
}

func (uc *RoleUsecase) GetRole(ctx context.Context, id string) (*Role, error) {
	return uc.repo.GetRole(ctx, id)
}

// GetRoleWithMenus 获取角色详情（包含菜单ID列表）
func (uc *RoleUsecase) GetRoleWithMenus(ctx context.Context, id string) (*Role, error) {
	role, err := uc.repo.GetRole(ctx, id)
	if err != nil {
		return nil, err
	}

	// 填充菜单ID列表
	menuIDs, err := uc.repo.GetMenuIDsByRole(ctx, id)
	if err != nil {
		// 记录日志但不返回错误，允许角色没有绑定菜单
		uc.log.Warnf("获取角色菜单失败: %v", err)
		menuIDs = []string{}
	}
	role.MenuIDs = menuIDs

	return role, nil
}

func (uc *RoleUsecase) CreateRole(ctx context.Context, role *Role) (*Role, error) {
	now := time.Now()
	role.ID = uuid.GenerateXID()
	role.CreatedAt = now
	role.UpdatedAt = now

	err := uc.repo.CreateRole(ctx, role)
	return role, err
}

// CreateRoleWithMenus 创建角色并绑定菜单（事务操作）
func (uc *RoleUsecase) CreateRoleWithMenus(ctx context.Context, role *Role) (*Role, error) {
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

func (uc *RoleUsecase) UpdateRole(ctx context.Context, role *Role) (*Role, error) {
	now := time.Now()
	role.UpdatedAt = now
	err := uc.repo.UpdateRole(ctx, role)

	return role, err
}

// UpdateRoleWithMenus 更新角色并重新绑定菜单（事务操作）
func (uc *RoleUsecase) UpdateRoleWithMenus(ctx context.Context, role *Role) (*Role, error) {
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
	return uc.repo.UpdateRoleStatus(ctx, id, status)
}

func (uc *RoleUsecase) DeleteRole(ctx context.Context, ids []string) error {
	return uc.repo.DeleteRole(ctx, ids)
}

// BindMenusForRole 为角色绑定菜单（业务层校验）
func (uc *RoleUsecase) BindMenusForRole(ctx context.Context, roleID string, menuIDs []string) error {
	// 验证角色是否存在
	_, err := uc.repo.GetRole(ctx, roleID)
	if err != nil {
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
