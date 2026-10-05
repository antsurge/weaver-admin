package biz

import (
	"context"
	"time"

	commonV1 "github.com/antsurge/weaver-admin/api/gen/go/common/v1"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/conf"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/crypto"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

// AdminDefaultPasswordFallback 新建管理员时若未配置 app.admin_default_password，则回退到此默认密码
const AdminDefaultPasswordFallback = "123456"

// 管理员状态枚举值
const (
	StatusEnabled  = "enabled"
	StatusDisabled = "disabled"
)

// MinPasswordLength 密码最小长度
const MinPasswordLength = 6

// Admin 是 biz 层的领域模型 (Domain Object)
// 这里的字段对应你 Proto 中定义的 Admin 消息
type Admin struct {
	ID        string
	TenantID  string // 所属租户
	RealName  string
	Username  string
	Email     string
	Phone     string
	Avatar    string
	Password  string // 用于创建时传递，查询时通常为空
	Status    string // enabled/disabled
	CreatedAt time.Time
	UpdatedAt time.Time
	// 归属部门ID
	DepartmentID string `json:"departmentId"`
	// 归属部门名称（列表展示用）
	DepartmentName string `json:"departmentName"`
	// 绑定的数据权限规则ID列表（空 = 跟随角色）
	DataPermissionIDs []string `json:"dataPermissionIds"`
	// 绑定的数据权限规则名称列表（列表展示用；空 = 跟随角色）
	DataPermissionNames []string `json:"dataPermissionNames"`
	// 关联的角色ID列表
	RoleIDs []string `json:"roleIds"`
	// 关联的角色名称列表（用于列表展示）
	RoleNames []string `json:"roleNames"`
	// 是否超级管理员（绑定超管角色）
	IsSuperAdmin bool `json:"isSuperAdmin"`
}

type ListAdminRequest struct {
	enthelper.PaginationParams
	// 过滤条件（均为可选，空字符串表示不过滤）
	Username     string
	RealName     string
	Phone        string
	Email        string
	Status       string
	DepartmentID string
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
	// FindByIDs 批量按ID查询（小批量场景，如日志列表补齐操作人姓名）
	FindByIDs(ctx context.Context, ids []string) ([]*Admin, error)
	// DeleteAdmin 批量软删除用户并清理角色绑定；superAdminCode 用于保护绑定超管角色的用户
	DeleteAdmin(ctx context.Context, ids []string, superAdminCode string) error
	// UpdatePassword 更新指定用户密码（已哈希）
	UpdatePassword(ctx context.Context, adminID, hashedPassword string) error
	// UpdateStatus 更新指定用户状态（enabled/disabled）
	UpdateStatus(ctx context.Context, adminID, status string) error
	// BatchUpdateStatus 批量更新用户状态
	BatchUpdateStatus(ctx context.Context, adminIDs []string, status string) error

	// ====== 唯一性校验方法 ======

	// ExistsUsername 判断用户名是否已存在（excludeID 非空时排除该用户，用于更新场景）
	ExistsUsername(ctx context.Context, username, excludeID string) (bool, error)
	// ExistsPhone 判断手机号是否已存在（excludeID 非空时排除该用户）
	ExistsPhone(ctx context.Context, phone, excludeID string) (bool, error)
	// ExistsEmail 判断邮箱是否已存在（excludeID 非空时排除该用户）
	ExistsEmail(ctx context.Context, email, excludeID string) (bool, error)

	// ====== 角色关联方法 ======

	// BindRolesForAdmin 为用户绑定角色（全量替换：先删除所有现有绑定，再批量插入新绑定）
	BindRolesForAdmin(ctx context.Context, adminID string, roleIDs []string) error

	// GetRoleIDsByAdmin 获取用户关联的角色ID列表
	GetRoleIDsByAdmin(ctx context.Context, adminID string) ([]string, error)

	// ====== 数据权限规则关联方法 ======

	// BindDataPermissionsForAdmin 为用户绑定数据权限规则（全量替换）
	BindDataPermissionsForAdmin(ctx context.Context, adminID string, dataPermissionIDs []string) error

	// GetDataPermissionIDsByAdmin 获取用户绑定的数据权限规则ID列表
	GetDataPermissionIDsByAdmin(ctx context.Context, adminID string) ([]string, error)

	// GetDataPermissionsByAdminIDs 批量获取多个用户绑定的数据权限规则（返回 map[adminID][]*DataPermission）
	GetDataPermissionsByAdminIDs(ctx context.Context, adminIDs []string) (map[string][]*DataPermission, error)

	// GetDepartmentNamesByIDs 批量查询部门名称（返回 map[deptID]name，用于个人中心展示归属部门）
	GetDepartmentNamesByIDs(ctx context.Context, deptIDs []string) (map[string]string, error)
}

// AdminUseCase 是业务逻辑控制器
type AdminUseCase struct {
	repo          AdminRepo
	roleRepo      RoleRepo
	tenantUsecase *TenantUsecase
	appConf       *conf.App
	log           *log.Helper
}

// NewAdminUseCase 构造函数，由 Wire 注入 Repo
func NewAdminUseCase(repo AdminRepo, roleRepo RoleRepo, tenantUsecase *TenantUsecase, appConf *conf.App, logger log.Logger) *AdminUseCase {
	return &AdminUseCase{repo: repo, roleRepo: roleRepo, tenantUsecase: tenantUsecase, appConf: appConf, log: log.NewHelper(logger)}
}

func (uc *AdminUseCase) ListAdmin(ctx context.Context, params *ListAdminRequest) (*ListAdminResponse, error) {
	res, err := uc.repo.ListAdmin(ctx, params)
	if err != nil {
		return nil, err
	}

	// 标记超级管理员账号（前端据此隐藏删除等操作）
	uc.markSuperAdmins(ctx, res.Items)

	return res, nil
}

// markSuperAdmins 为列表中的用户填充 IsSuperAdmin 标记
func (uc *AdminUseCase) markSuperAdmins(ctx context.Context, admins []*Admin) {
	if uc.appConf == nil || uc.appConf.SuperAdminCode == "" || len(admins) == 0 {
		return
	}
	for _, admin := range admins {
		if len(admin.RoleIDs) == 0 {
			continue
		}
		codes, err := uc.roleRepo.GetCodesByIds(ctx, admin.RoleIDs)
		if err != nil {
			continue
		}
		for _, code := range codes {
			if code == uc.appConf.SuperAdminCode {
				admin.IsSuperAdmin = true
				break
			}
		}
	}
}

func (uc *AdminUseCase) CreateAdmin(ctx context.Context, admin *Admin) (*Admin, error) {
	// 状态枚举校验（空值默认启用）
	if admin.Status == "" {
		admin.Status = StatusEnabled
	}
	if !isValidStatus(admin.Status) {
		return nil, errors.BadRequest("INVALID_STATUS", "状态值非法，仅支持 enabled/disabled")
	}

	// 唯一性校验（用户名 / 手机号 / 邮箱）
	if err := uc.checkUnique(ctx, admin, ""); err != nil {
		return nil, err
	}

	// 租户用户数配额校验（max_users>0 时强制）
	if uc.tenantUsecase != nil {
		if err := uc.tenantUsecase.CheckQuota(ctx, admin.TenantID, "user"); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	admin.ID = uuid.GenerateXID()
	admin.CreatedAt = now
	admin.UpdatedAt = now

	// 密码为空则使用默认密码（默认密码来自 app.admin_default_password 配置）
	if admin.Password == "" {
		admin.Password = uc.defaultPassword()
	}

	// 密码 bcrypt 哈希后入库（登录校验使用 CheckPasswordHash，存储必须是哈希值）
	hashed, err := crypto.HashPassword(admin.Password)
	if err != nil {
		return nil, err
	}
	admin.Password = hashed

	err = uc.repo.CreateAdmin(ctx, admin)
	if err != nil {
		return nil, err
	}

	// 绑定角色（如果有）
	if len(admin.RoleIDs) > 0 {
		if err := uc.repo.BindRolesForAdmin(ctx, admin.ID, admin.RoleIDs); err != nil {
			return nil, err
		}
	}

	// 绑定数据权限规则（如果有）
	if len(admin.DataPermissionIDs) > 0 {
		if err := uc.repo.BindDataPermissionsForAdmin(ctx, admin.ID, admin.DataPermissionIDs); err != nil {
			return nil, err
		}
	}

	return admin, nil
}

func (uc *AdminUseCase) UpdateAdmin(ctx context.Context, admin *Admin) (*Admin, error) {
	if admin.ID == "" {
		return nil, errors.BadRequest("BAD_REQUEST", "用户ID不能为空")
	}

	// 状态枚举校验（更新时字段可选，非空才校验）
	if admin.Status != "" && !isValidStatus(admin.Status) {
		return nil, errors.BadRequest("INVALID_STATUS", "状态值非法，仅支持 enabled/disabled")
	}

	// 唯一性校验（用户名 / 手机号 / 邮箱，排除自身）
	if err := uc.checkUnique(ctx, admin, admin.ID); err != nil {
		return nil, err
	}

	now := time.Now()
	admin.UpdatedAt = now

	// 密码非空表示修改密码，需要哈希；空表示不修改（data 层会跳过密码更新）
	if admin.Password != "" {
		hashed, err := crypto.HashPassword(admin.Password)
		if err != nil {
			return nil, err
		}
		admin.Password = hashed
	}

	err := uc.repo.UpdateAdmin(ctx, admin)
	if err != nil {
		return nil, err
	}

	// 仅当显式携带角色列表时才重新绑定（全量替换）。
	// roleIds 为 nil 表示"不修改角色"，避免前端未传时误清空用户已有角色。
	if admin.RoleIDs != nil {
		if err := uc.repo.BindRolesForAdmin(ctx, admin.ID, admin.RoleIDs); err != nil {
			return nil, err
		}
	}

	// 仅当显式携带数据权限规则列表时才重新绑定（全量替换）。
	// dataPermissionIDs 为 nil 表示"不修改"，避免前端未传时误清空用户已有规则。
	if admin.DataPermissionIDs != nil {
		if err := uc.repo.BindDataPermissionsForAdmin(ctx, admin.ID, admin.DataPermissionIDs); err != nil {
			return nil, err
		}
	}

	return admin, nil
}

// UpdateStatus 启用/禁用指定用户
func (uc *AdminUseCase) UpdateStatus(ctx context.Context, adminID, status string) error {
	if adminID == "" || status == "" {
		return errors.BadRequest("BAD_REQUEST", "参数不能为空")
	}
	if !isValidStatus(status) {
		return errors.BadRequest("INVALID_STATUS", "状态值非法，仅支持 enabled/disabled")
	}

	// 超级管理员账号不允许被禁用，避免系统失去唯一超管入口
	if status == StatusDisabled {
		admin, err := uc.repo.FindByID(ctx, adminID)
		if err != nil {
			return err
		}
		if admin != nil && uc.isSuperAdmin(ctx, adminID) {
			return errors.BadRequest("CANNOT_DISABLE_SUPERADMIN", "超级管理员账号不允许禁用")
		}
	}

	return uc.repo.UpdateStatus(ctx, adminID, status)
}

// BatchUpdateStatus 批量启用/禁用用户
func (uc *AdminUseCase) BatchUpdateStatus(ctx context.Context, adminIDs []string, status string) error {
	if len(adminIDs) == 0 || status == "" {
		return errors.BadRequest("BAD_REQUEST", "参数不能为空")
	}
	if !isValidStatus(status) {
		return errors.BadRequest("INVALID_STATUS", "状态值非法，仅支持 enabled/disabled")
	}

	// 禁用操作需排除超级管理员账号，避免系统失去唯一超管入口
	ids := adminIDs
	if status == StatusDisabled {
		ids = make([]string, 0, len(adminIDs))
		for _, id := range adminIDs {
			if id == "" || uc.isSuperAdmin(ctx, id) {
				continue
			}
			ids = append(ids, id)
		}
		if len(ids) == 0 {
			return errors.BadRequest("CANNOT_DISABLE_SUPERADMIN", "超级管理员账号不允许禁用")
		}
	}

	return uc.repo.BatchUpdateStatus(ctx, ids, status)
}

// isSuperAdmin 判断指定用户是否绑定了超级管理员角色
func (uc *AdminUseCase) isSuperAdmin(ctx context.Context, adminID string) bool {
	if uc.appConf == nil || uc.appConf.SuperAdminCode == "" {
		return false
	}
	roleIDs, err := uc.repo.GetRoleIDsByAdmin(ctx, adminID)
	if err != nil || len(roleIDs) == 0 {
		return false
	}
	codes, err := uc.roleRepo.GetCodesByIds(ctx, roleIDs)
	if err != nil {
		return false
	}
	for _, code := range codes {
		if code == uc.appConf.SuperAdminCode {
			return true
		}
	}
	return false
}

// checkUnique 校验用户名/手机号/邮箱唯一性；excludeID 非空时排除该用户（更新场景）
func (uc *AdminUseCase) checkUnique(ctx context.Context, admin *Admin, excludeID string) error {
	if admin.Username != "" {
		exists, err := uc.repo.ExistsUsername(ctx, admin.Username, excludeID)
		if err != nil {
			return err
		}
		if exists {
			return errors.BadRequest("USERNAME_ALREADY_EXISTS", "用户名已存在")
		}
	}
	if admin.Phone != "" {
		exists, err := uc.repo.ExistsPhone(ctx, admin.Phone, excludeID)
		if err != nil {
			return err
		}
		if exists {
			return errors.BadRequest("PHONE_ALREADY_EXISTS", "手机号已存在")
		}
	}
	if admin.Email != "" {
		exists, err := uc.repo.ExistsEmail(ctx, admin.Email, excludeID)
		if err != nil {
			return err
		}
		if exists {
			return errors.BadRequest("EMAIL_ALREADY_EXISTS", "邮箱已存在")
		}
	}
	return nil
}

// isValidStatus 校验状态枚举值
func isValidStatus(status string) bool {
	return status == StatusEnabled || status == StatusDisabled
}

// IsUsernameExists 用户名是否存在（excludeID 非空时排除该用户，用于编辑场景）
func (uc *AdminUseCase) IsUsernameExists(ctx context.Context, username, excludeID string) (bool, error) {
	if username == "" {
		return false, nil
	}
	return uc.repo.ExistsUsername(ctx, username, excludeID)
}

// GetAdminWithRoles 获取用户详情（包含角色ID列表与数据权限规则ID列表）
func (uc *AdminUseCase) GetAdminWithRoles(ctx context.Context, id string) (*Admin, error) {
	admin, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, commonV1.ErrorNotFoundRecord("NOT_FOUND_RECORD")
	}

	// 填充角色ID列表
	roleIDs, err := uc.repo.GetRoleIDsByAdmin(ctx, id)
	if err != nil {
		uc.log.Warnf("获取用户角色失败: %v", err)
		roleIDs = []string{}
	}

	admin.RoleIDs = roleIDs
	admin.IsSuperAdmin = uc.isSuperAdmin(ctx, id)

	// 填充数据权限规则ID列表
	dpIDs, err := uc.repo.GetDataPermissionIDsByAdmin(ctx, id)
	if err != nil {
		uc.log.Warnf("获取用户数据权限规则失败: %v", err)
		dpIDs = []string{}
	}
	admin.DataPermissionIDs = dpIDs

	return admin, nil
}

// DeleteAdmin 删除用户（软删除）
// operatorID 为当前操作者 ID：禁止删除自己；绑定超级管理员角色的用户禁止删除
func (uc *AdminUseCase) DeleteAdmin(ctx context.Context, ids []string, operatorID string) error {
	for _, id := range ids {
		if id != "" && id == operatorID {
			return errors.BadRequest("CANNOT_DELETE_SELF", "不能删除当前登录账号")
		}
	}
	return uc.repo.DeleteAdmin(ctx, ids, uc.appConf.SuperAdminCode)
}

// ResetPassword 管理员重置指定用户密码
func (uc *AdminUseCase) ResetPassword(ctx context.Context, adminID, password string) error {
	if adminID == "" || password == "" {
		return errors.BadRequest("BAD_REQUEST", "参数不能为空")
	}
	if len(password) < MinPasswordLength {
		return errors.BadRequest("PASSWORD_TOO_SHORT", "密码长度不能少于6位")
	}
	hashed, err := crypto.HashPassword(password)
	if err != nil {
		return err
	}
	return uc.repo.UpdatePassword(ctx, adminID, hashed)
}

// defaultPassword 返回新建管理员时的默认密码：
// 优先使用 app.admin_default_password 配置；未配置时回退到内置默认值。
func (uc *AdminUseCase) defaultPassword() string {
	if uc.appConf != nil && uc.appConf.AdminDefaultPassword != "" {
		return uc.appConf.AdminDefaultPassword
	}
	return AdminDefaultPasswordFallback
}
