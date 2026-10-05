package biz

import (
	"context"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
)

// AuthzPermission 用户的接口权限集合
type AuthzPermission struct {
	// SuperAdmin 是否超级管理员（拥有全部接口权限）
	SuperAdmin bool
	// Codes 接口权限码集合，格式 "METHOD|pathTemplate"，如 "GET|/admin/v1/admin/{adminId}"
	// 来源：启用角色 → 绑定的启用菜单（含按钮/接口权限）→ ApiPermission
	Codes []string
}

// AuthzRepo 鉴权数据查询接口
type AuthzRepo interface {
	// GetPermissionByAdmin 查询用户权限：是否超级管理员 + 绑定的接口权限码
	GetPermissionByAdmin(ctx context.Context, adminID string, superAdminCode string) (*AuthzPermission, error)
}

// AuthzCache 权限缓存接口（实现应设置 TTL，保证权限变更可在有限时间内生效）
type AuthzCache interface {
	Get(ctx context.Context, adminID string) (*AuthzPermission, error)
	Set(ctx context.Context, adminID string, p *AuthzPermission) error
}

// AuthzUseCase 接口权限用例
type AuthzUseCase struct {
	repo    AuthzRepo
	cache   AuthzCache
	appConf *conf.App
	log     *log.Helper
}

// NewAuthzUseCase 创建接口权限用例
func NewAuthzUseCase(
	repo AuthzRepo,
	cache AuthzCache,
	appConf *conf.App,
	logger log.Logger,
) *AuthzUseCase {
	return &AuthzUseCase{
		repo:    repo,
		cache:   cache,
		appConf: appConf,
		log:     log.NewHelper(logger),
	}
}

// GetPermission 获取用户接口权限（带缓存）。
// 缓存读取失败时回源 DB，不影响正确性；权限变更在缓存 TTL 内延迟生效。
func (uc *AuthzUseCase) GetPermission(ctx context.Context, adminID string) (*AuthzPermission, error) {
	if adminID == "" {
		return &AuthzPermission{}, nil
	}

	// 1. 读缓存（任何缓存错误都回源，不做 fail-closed）
	if p, err := uc.cache.Get(ctx, adminID); err == nil {
		return p, nil
	} else {
		uc.log.Debugf("authz cache miss, adminID=%s, err=%v", adminID, err)
	}

	// 2. 回源查询
	p, err := uc.repo.GetPermissionByAdmin(ctx, adminID, uc.appConf.SuperAdminCode)
	if err != nil {
		return nil, err
	}

	// 3. 写缓存（best-effort，失败仅记日志）
	if err := uc.cache.Set(ctx, adminID, p); err != nil {
		uc.log.Errorf("authz cache set error, adminID=%s: %v", adminID, err)
	}

	return p, nil
}

// IsSuperAdmin 判断用户是否为平台超管（供租户守卫跨租户切换授权）。
func (uc *AuthzUseCase) IsSuperAdmin(ctx context.Context, adminID string) (bool, error) {
	p, err := uc.GetPermission(ctx, adminID)
	if err != nil {
		return false, err
	}
	return p != nil && p.SuperAdmin, nil
}
