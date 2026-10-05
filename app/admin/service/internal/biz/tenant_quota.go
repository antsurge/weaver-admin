package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
)

// TenantUsage 租户资源用量统计（用于配额展示与强制校验）
type TenantUsage struct {
	TenantID string `json:"tenantId"`
	// 当前用户数
	Admins int `json:"admins"`
	// 当前角色数
	Roles int `json:"roles"`
	// 当前菜单数
	Menus int `json:"menus"`
	// 当前部门数
	Departments int `json:"departments"`
	// 当前岗位数
	Positions int `json:"positions"`
	// 当前字典类型数
	DictTypes int `json:"dictTypes"`
	// 当前定时任务数
	Jobs int `json:"jobs"`
	// 当前通知数
	Notifications int `json:"notifications"`
}

// TenantQuota 租户配额（配合 Tenant 的 MaxUsers/MaxRoles 使用）
type TenantQuota struct {
	TenantID string `json:"tenantId"`
	MaxUsers int    `json:"maxUsers"`
	MaxRoles int    `json:"maxRoles"`
}

// GetUsage 统计租户下各资源当前用量。
// 实现上通过租户隔离上下文将查询限定在指定租户作用域内。
func (uc *TenantUsecase) GetUsage(ctx context.Context, tenantID string) (*TenantUsage, error) {
	return uc.repo.GetUsage(ctx, tenantID)
}

// CheckQuota 校验租户配额是否允许继续创建指定资源。
// resource 取值：user / role
func (uc *TenantUsecase) CheckQuota(ctx context.Context, tenantID, resource string) error {
	if tenantID == "" {
		return nil
	}

	t, err := uc.repo.GetTenantByID(ctx, tenantID)
	if err != nil {
		return err
	}
	if t == nil {
		return errors.BadRequest("TENANT_NOT_FOUND", "租户不存在")
	}
	// 租户被禁用时不允许新增资源
	if t.Status == StatusDisabled {
		return errors.BadRequest("TENANT_DISABLED", "租户已禁用，无法新增资源")
	}

	usage, err := uc.repo.GetUsage(ctx, tenantID)
	if err != nil {
		return err
	}

	switch resource {
	case "user":
		if t.MaxUsers > 0 && usage.Admins >= t.MaxUsers {
			return errors.BadRequest("QUOTA_EXCEEDED",
				"租户用户数已达上限（当前 "+itoa(usage.Admins)+"/"+itoa(t.MaxUsers)+"），请联系平台管理员提升配额")
		}
	case "role":
		if t.MaxRoles > 0 && usage.Roles >= t.MaxRoles {
			return errors.BadRequest("QUOTA_EXCEEDED",
				"租户角色数已达上限（当前 "+itoa(usage.Roles)+"/"+itoa(t.MaxRoles)+"），请联系平台管理员提升配额")
		}
	}
	return nil
}

// IsTenantUsable 判断租户当前是否可用：存在、未删除、enabled 且未过期。
// 用于请求期租户状态守卫（tenantguard 中间件）。
func (uc *TenantUsecase) IsTenantUsable(ctx context.Context, tenantID string) (bool, error) {
	if tenantID == "" {
		return false, nil
	}

	t, err := uc.repo.GetTenantByID(ctx, tenantID)
	if err != nil {
		return false, err
	}
	if t == nil {
		return false, nil
	}
	if t.Status == StatusDisabled {
		return false, nil
	}
	if t.ExpireAt != nil && !t.ExpireAt.After(time.Now()) {
		return false, nil
	}
	return true, nil
}

// itoa 简易 int->string（避免引入 strconv 依赖到频繁调用路径）
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var b [20]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		b[pos] = '-'
	}
	return string(b[pos:])
}
