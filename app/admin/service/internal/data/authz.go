package data

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/conf"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/admin"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/apiinterface"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/menu"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/menuapipermission"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/ent/role"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// authzRepo 鉴权数据查询实现
type authzRepo struct {
	data *Data
	log  *log.Helper
}

// NewAuthzRepo 创建鉴权数据查询仓库
func NewAuthzRepo(data *Data, logger log.Logger) biz.AuthzRepo {
	return &authzRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// GetPermissionByAdmin 查询用户接口权限。
// 链路：admin_roles → roles(启用) 判定超级管理员；role_menus → menus(启用)
// → api_permissions 生成权限码（"METHOD|path"）。
func (r *authzRepo) GetPermissionByAdmin(
	ctx context.Context,
	adminID string,
	superAdminCode string,
) (*biz.AuthzPermission, error) {
	p := &biz.AuthzPermission{}

	// 1. 用户的启用角色
	roles, err := r.data.db.Role.Query().
		Where(
			role.HasAdminsWith(admin.IDEQ(adminID)),
			role.StatusEQ(role.StatusEnabled),
		).
		All(ctx)
	if err != nil {
		r.log.Errorf("authz query roles error, adminID=%s: %v", adminID, err)
		return nil, err
	}
	if len(roles) == 0 {
		return p, nil
	}

	// 2. 超级管理员判定
	roleIDs := make([]string, 0, len(roles))
	for _, rg := range roles {
		if superAdminCode != "" && rg.Code == superAdminCode {
			p.SuperAdmin = true
			return p, nil
		}
		roleIDs = append(roleIDs, rg.ID)
	}

	// 3. 角色绑定的启用菜单
	menus, err := r.data.db.Menu.Query().
		Where(
			menu.HasRolesWith(role.IDIn(roleIDs...)),
			menu.StatusEQ(menu.StatusEnabled),
		).
		All(ctx)
	if err != nil {
		r.log.Errorf("authz query menus error, adminID=%s: %v", adminID, err)
		return nil, err
	}
	if len(menus) == 0 {
		return p, nil
	}

	// 4. 菜单绑定的接口 code（menu_api_permission）→ 实时反查 api_interface 生成权限码
	menuIDs := make([]string, 0, len(menus))
	for _, m := range menus {
		menuIDs = append(menuIDs, m.ID)
	}

	codeSet := make(map[string]struct{})
	bindings, err := r.data.db.MenuApiPermission.Query().
		Where(menuapipermission.MenuIDIn(menuIDs...)).
		All(ctx)
	if err != nil {
		r.log.Errorf("authz query menu api bindings error, adminID=%s: %v", adminID, err)
		return nil, err
	}
	if len(bindings) > 0 {
		codes := make([]string, 0, len(bindings))
		for _, b := range bindings {
			codes = append(codes, b.APICode)
		}
		apis, err := r.data.db.ApiInterface.Query().
			Where(apiinterface.CodeIn(codes...)).
			All(ctx)
		if err != nil {
			r.log.Errorf("authz query api interfaces error, adminID=%s: %v", adminID, err)
			return nil, err
		}
		for _, ap := range apis {
			if ap.Method == "" || ap.Path == "" {
				continue
			}
			codeSet[strings.ToUpper(ap.Method)+"|"+ap.Path] = struct{}{}
		}
	}

	p.Codes = make([]string, 0, len(codeSet))
	for code := range codeSet {
		p.Codes = append(p.Codes, code)
	}

	return p, nil
}

// authzCache 权限 Redis 缓存实现
type authzCache struct {
	data *Data
	conf *conf.Data
	log  *log.Helper
	ttl  time.Duration
}

// NewAuthzCache 创建权限缓存
func NewAuthzCache(data *Data, logger log.Logger, c *conf.Data) biz.AuthzCache {
	return &authzCache{
		data: data,
		conf: c,
		log:  log.NewHelper(logger),
		ttl:  60 * time.Second, // 权限变更最长 60s 延迟生效
	}
}

func (c *authzCache) key(adminID string) string {
	return fmt.Sprintf("%sauthz:%s", c.conf.Redis.KeyPrefix, adminID)
}

func (c *authzCache) Get(ctx context.Context, adminID string) (*biz.AuthzPermission, error) {
	val, err := c.data.redis.Get(ctx, c.key(adminID)).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("authz cache miss, adminID=%s", adminID)
	}
	if err != nil {
		return nil, err
	}
	p := &biz.AuthzPermission{}
	if err := json.Unmarshal([]byte(val), p); err != nil {
		return nil, err
	}
	return p, nil
}

func (c *authzCache) Set(ctx context.Context, adminID string, p *biz.AuthzPermission) error {
	buf, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return c.data.redis.Set(ctx, c.key(adminID), buf, c.ttl).Err()
}
