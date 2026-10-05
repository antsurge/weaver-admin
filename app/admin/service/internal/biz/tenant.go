package biz

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/tenant"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

// Tenant 租户领域模型（平台级）。
type Tenant struct {
	ID           string     // 租户ID（租户隔离字段 tenant_id 的取值）
	Code         string     // 租户编码（登录用，创建后不可修改）
	Name         string     // 租户名称
	Status       string     // enabled=启用 disabled=禁用
	ExpireAt     *time.Time // 到期时间（为空表示永不过期）
	MaxUsers     int        // 最大用户数（0=不限）
	MaxRoles     int        // 最大角色数（0=不限）
	ContactName  string     // 联系人
	ContactPhone string     // 联系电话
	Remark       string     // 备注
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

// TenantRepo 租户仓库（平台级表，不参与租户隔离）。
type TenantRepo interface {
	// GetByCode 按租户编码查询租户（不含已删除）。
	GetByCode(ctx context.Context, code string) (*Tenant, error)
	// ListTenant 租户分页列表。
	ListTenant(ctx context.Context, params *ListTenantRequest, opts ...*ListTenantOption) (*ListTenantResponse, error)
	// GetTenantByID 按 ID 查询租户（不含已删除）。
	GetTenantByID(ctx context.Context, id string) (*Tenant, error)
	// CreateTenant 创建租户。
	CreateTenant(ctx context.Context, t *Tenant) error
	// UpdateTenant 更新租户（普通字段）。
	UpdateTenant(ctx context.Context, t *Tenant) error
	// DeleteTenant 批量软删除租户。
	DeleteTenant(ctx context.Context, ids []string) error
	// UpdateTenantStatus 更新租户状态。
	UpdateTenantStatus(ctx context.Context, id, status string) error
	// ExistsTenantCode 判断租户编码是否存在（excludeID 非空时排除该租户，用于更新场景）。
	ExistsTenantCode(ctx context.Context, code, excludeID string) (bool, error)
	// GetUsage 统计租户下各资源当前用量（用户/角色/菜单/部门等）。
	GetUsage(ctx context.Context, tenantID string) (*TenantUsage, error)
}

type ListTenantRequest struct {
	enthelper.PaginationParams
	Name   string `form:"name" query:"name"`
	Code   string `form:"code" query:"code"`
	Status string `form:"status" query:"status"`
}

type ListTenantOption struct {
	enthelper.QueryOption
}

type ListTenantResponse struct {
	Data  []*Tenant
	Total int
}

// TenantUsecase 租户业务逻辑控制器。
type TenantUsecase struct {
	repo TenantRepo
	log  *log.Helper
}

func NewTenantUsecase(repo TenantRepo, logger log.Logger) *TenantUsecase {
	return &TenantUsecase{repo: repo, log: log.NewHelper(logger)}
}

// List 租户分页列表。
func (uc *TenantUsecase) List(ctx context.Context, req *ListTenantRequest) (*ListTenantResponse, error) {
	return uc.repo.ListTenant(ctx, req)
}

// GetTenant 租户详情。
func (uc *TenantUsecase) GetTenant(ctx context.Context, id string) (*Tenant, error) {
	t, err := uc.repo.GetTenantByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, errors.BadRequest("TENANT_NOT_FOUND", "租户不存在")
	}
	return t, nil
}

// CreateTenant 创建租户。
func (uc *TenantUsecase) CreateTenant(ctx context.Context, req *Tenant) (*Tenant, error) {
	// 状态默认启用
	if req.Status == "" {
		req.Status = StatusEnabled
	}
	if !isValidStatus(req.Status) {
		return nil, errors.BadRequest("INVALID_STATUS", "状态值非法，仅支持 enabled/disabled")
	}

	// 编码唯一性校验
	exists, err := uc.repo.ExistsTenantCode(ctx, req.Code, "")
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.BadRequest("CODE_ALREADY_EXISTS", "租户编码已存在")
	}

	now := time.Now()
	req.ID = uuid.GenerateXID()
	req.CreatedAt = now
	req.UpdatedAt = now

	if err := uc.repo.CreateTenant(ctx, req); err != nil {
		uc.log.Errorf("create tenant failed: %v", err)
		return nil, errors.BadRequest("CREATE_TENANT_FAIL", "创建租户失败")
	}
	return req, nil
}

// UpdateTenant 更新租户。
func (uc *TenantUsecase) UpdateTenant(ctx context.Context, req *Tenant) (*Tenant, error) {
	if req.ID == "" {
		return nil, errors.BadRequest("BAD_REQUEST", "租户ID不能为空")
	}

	// 校验租户存在
	old, err := uc.repo.GetTenantByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if old == nil {
		return nil, errors.BadRequest("TENANT_NOT_FOUND", "租户不存在")
	}

	// 默认租户只允许修改基础信息，不允许被禁用/过期
	if old.ID == tenant.DefaultTenantID {
		if req.Status == StatusDisabled {
			return nil, errors.BadRequest("CANNOT_DISABLE_DEFAULT", "默认租户不允许禁用")
		}
		req.Code = old.Code // 编码不可变
	}

	// 状态非空才校验（复用老租户的状态）
	if req.Status != "" && !isValidStatus(req.Status) {
		return nil, errors.BadRequest("INVALID_STATUS", "状态值非法，仅支持 enabled/disabled")
	}

	req.UpdatedAt = time.Now()
	if err := uc.repo.UpdateTenant(ctx, req); err != nil {
		uc.log.Errorf("update tenant failed: %v", err)
		return nil, errors.BadRequest("UPDATE_TENANT_FAIL", "更新租户失败")
	}

	return uc.repo.GetTenantByID(ctx, req.ID)
}

// UpdateTenantStatus 启用/禁用租户。
func (uc *TenantUsecase) UpdateTenantStatus(ctx context.Context, id, status string) error {
	if id == "" || status == "" {
		return errors.BadRequest("BAD_REQUEST", "参数不能为空")
	}
	if !isValidStatus(status) {
		return errors.BadRequest("INVALID_STATUS", "状态值非法，仅支持 enabled/disabled")
	}

	// 默认租户不允许禁用，避免系统失去默认租户入口
	if id == tenant.DefaultTenantID && status == StatusDisabled {
		return errors.BadRequest("CANNOT_DISABLE_DEFAULT", "默认租户不允许禁用")
	}

	// 校验租户存在
	old, err := uc.repo.GetTenantByID(ctx, id)
	if err != nil {
		return err
	}
	if old == nil {
		return errors.BadRequest("TENANT_NOT_FOUND", "租户不存在")
	}

	return uc.repo.UpdateTenantStatus(ctx, id, status)
}

// DeleteTenant 删除租户（软删除）。默认租户不允许删除。
func (uc *TenantUsecase) DeleteTenant(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return errors.BadRequest("BAD_REQUEST", "参数不能为空")
	}
	for _, id := range ids {
		if id == tenant.DefaultTenantID {
			return errors.BadRequest("CANNOT_DELETE_DEFAULT", "默认租户不允许删除")
		}
	}
	return uc.repo.DeleteTenant(ctx, ids)
}
