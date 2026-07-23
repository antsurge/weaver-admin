package biz

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/log"
)

// Admin 是 biz 层的领域模型 (Domain Object)
// 这里的字段对应你 Proto 中定义的 Admin 消息
type Admin struct {
	ID        string
	RealName  string
	Username  string
	Email     string
	Phone     string
	Avatar    string
	Password  string // 用于创建时传递，查询时通常为空
	CreatedAt time.Time
	UpdatedAt time.Time
	// 关联的角色ID列表
	RoleIDs []string `json:"roleIds"`
}

type ListAdminRequest struct {
	enthelper.PaginationParams
}

type ListAdminResponse struct {
	Items []*Admin
	Total int
}

type ListAdminOption struct {
	enthelper.QueryOption
}

// AdminRepo 是数据持久化层的接口定义
// 它不关心底层是 PostgreSQL 还是 Ent，只关心业务行为
type AdminRepo interface {
	ListAdmin(ctx context.Context, params *ListAdminRequest, opts ...*ListAdminOption) (*ListAdminResponse, error)
	CreateAdmin(ctx context.Context, admin *Admin) error
	UpdateAdmin(ctx context.Context, admin *Admin) error
	FindByUsername(ctx context.Context, username string) (*Admin, error)
	FindByID(ctx context.Context, id string) (*Admin, error)
	DeleteAdmin(ctx context.Context, ids []string) error

	// ====== 角色关联方法 ======

	// BindRolesForAdmin 为用户绑定角色（全量替换：先删除所有现有绑定，再批量插入新绑定）
	BindRolesForAdmin(ctx context.Context, adminID string, roleIDs []string) error

	// GetRoleIDsByAdmin 获取用户关联的角色ID列表
	GetRoleIDsByAdmin(ctx context.Context, adminID string) ([]string, error)
}

// AdminUseCase 是业务逻辑控制器
type AdminUseCase struct {
	repo     AdminRepo
	roleRepo RoleRepo
	log      *log.Helper
}

// NewAdminUseCase 构造函数，由 Wire 注入 Repo
func NewAdminUseCase(repo AdminRepo, roleRepo RoleRepo, logger log.Logger) *AdminUseCase {
	return &AdminUseCase{repo: repo, roleRepo: roleRepo, log: log.NewHelper(logger)}
}

func (uc *AdminUseCase) ListAdmin(ctx context.Context, params *ListAdminRequest) (*ListAdminResponse, error) {
	return uc.repo.ListAdmin(ctx, params)
}

func (uc *AdminUseCase) CreateAdmin(ctx context.Context, admin *Admin) (*Admin, error) {
	now := time.Now()
	admin.ID = uuid.GenerateXID()
	admin.CreatedAt = now
	admin.UpdatedAt = now

	err := uc.repo.CreateAdmin(ctx, admin)
	if err != nil {
		return nil, err
	}

	// 绑定角色（如果有）
	if len(admin.RoleIDs) > 0 {
		if err := uc.repo.BindRolesForAdmin(ctx, admin.ID, admin.RoleIDs); err != nil {
			return nil, err
		}
	}

	return admin, nil
}

func (uc *AdminUseCase) UpdateAdmin(ctx context.Context, admin *Admin) (*Admin, error) {
	now := time.Now()
	admin.UpdatedAt = now
	err := uc.repo.UpdateAdmin(ctx, admin)
	if err != nil {
		return nil, err
	}

	// 重新绑定角色（全量替换）
	if err := uc.repo.BindRolesForAdmin(ctx, admin.ID, admin.RoleIDs); err != nil {
		return nil, err
	}

	return admin, nil
}

// GetAdminWithRoles 获取用户详情（包含角色ID列表）
func (uc *AdminUseCase) GetAdminWithRoles(ctx context.Context, id string) (*Admin, error) {
	admin, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 填充角色ID列表
	roleIDs, err := uc.repo.GetRoleIDsByAdmin(ctx, id)
	if err != nil {
		uc.log.Warnf("获取用户角色失败: %v", err)
		roleIDs = []string{}
	}
	admin.RoleIDs = roleIDs

	return admin, nil
}

// DeleteAdmin 删除用户（软删除）
func (uc *AdminUseCase) DeleteAdmin(ctx context.Context, ids []string) error {
	return uc.repo.DeleteAdmin(ctx, ids)
}
