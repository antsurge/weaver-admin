package biz

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

// 数据权限范围常量（服务端按用户/角色绑定的数据权限规则计算下发，用于兼容旧协议字段）
const (
	DataScopeAll       = "ALL"
	DataScopeDeptChild = "DEPT_AND_CHILD"
	DataScopeDept      = "DEPT"
	DataScopeSelf      = "SELF"
)

// DataScopeRank 数据范围权限大小排序（数值越大权限越大，用于多角色取并集）
var DataScopeRank = map[string]int{
	DataScopeAll:       4,
	DataScopeDeptChild: 3,
	DataScopeDept:      2,
	DataScopeSelf:      1,
}

// 数据权限规则 scope_type 取值（规则库定义）
const (
	ScopeTypeAllData  = "1" // 全部数据
	ScopeTypeCustom   = "2" // 自定义（指定部门/角色）
	ScopeTypeSelfDept = "3" // 本部门数据
	ScopeTypeSubDept  = "4" // 本部门及以下
	ScopeTypeSelfOnly = "5" // 仅本人数据
)

// DataScopeFromRule 将数据权限规则 scope_type 映射为数据范围四值（用于 current-user 下发与兼容）
// 自定义范围（指定部门/角色）在合并时视为“本部门及以下”级别，交由执行层按规则精确过滤。
func DataScopeFromRule(scopeType string) string {
	switch scopeType {
	case ScopeTypeAllData:
		return DataScopeAll
	case ScopeTypeSelfDept:
		return DataScopeDept
	case ScopeTypeSubDept, ScopeTypeCustom:
		return DataScopeDeptChild
	case ScopeTypeSelfOnly:
		return DataScopeSelf
	}
	return DataScopeSelf
}

type DataPermission struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Code      string     `json:"code"`
	ScopeType string     `json:"scopeType"`
	DeptIds   string     `json:"deptIds"`
	RoleIds   string     `json:"roleIds"`
	Status    string     `json:"status"`
	Remark    string     `json:"remark"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

type ListDataPermissionRequest struct {
	enthelper.PaginationParams
	Name      string `form:"name" query:"name"`
	Code      string `form:"code" query:"code"`
	ScopeType string `form:"scopeType" query:"scopeType"`
	Status    string `form:"status" query:"status"`
}

type ListDataPermissionOption struct {
	enthelper.QueryOption
}

type ListDataPermissionResponse struct {
	Data  []*DataPermission
	Total int
}

type DataPermissionRepo interface {
	ListDataPermission(context.Context, *ListDataPermissionRequest, ...*ListDataPermissionOption) (*ListDataPermissionResponse, error)
	CreateDataPermission(context.Context, *DataPermission) error
	UpdateDataPermission(context.Context, *DataPermission) error
	DeleteDataPermission(context.Context, []string) error
	UpdateDataPermissionStatus(context.Context, string, string) error
	IsDataPermissionCodeExists(context.Context, string, string) (bool, error)

	// ListEnabledDataPermissionsByIDs 按ID批量查询启用的数据权限规则
	ListEnabledDataPermissionsByIDs(context.Context, []string) ([]*DataPermission, error)
}

type DataPermissionUsecase struct {
	repo DataPermissionRepo
	log  *log.Helper
}

func NewDataPermissionUsecase(repo DataPermissionRepo, logger log.Logger) *DataPermissionUsecase {
	return &DataPermissionUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *DataPermissionUsecase) List(ctx context.Context, req *ListDataPermissionRequest) (*ListDataPermissionResponse, error) {
	return uc.repo.ListDataPermission(ctx, req)
}

func (uc *DataPermissionUsecase) Create(ctx context.Context, req *DataPermission) (*DataPermission, error) {
	data := req

	// 唯一性校验（规则编码）
	if err := uc.validateCodeUnique(ctx, data, ""); err != nil {
		return nil, err
	}

	data.ID = uuid.GenerateXID()
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	if data.Status == "" {
		data.Status = "enabled"
	}
	if data.ScopeType == "" {
		data.ScopeType = "1"
	}
	if data.DeptIds == "" {
		data.DeptIds = "[]"
	}
	if data.RoleIds == "" {
		data.RoleIds = "[]"
	}

	err := uc.repo.CreateDataPermission(ctx, data)
	return data, err
}

func (uc *DataPermissionUsecase) Update(ctx context.Context, req *DataPermission) (*DataPermission, error) {
	data := req

	// 唯一性校验（规则编码，编辑时排除自身）
	if err := uc.validateCodeUnique(ctx, data, data.ID); err != nil {
		return nil, err
	}

	data.UpdatedAt = time.Now()

	err := uc.repo.UpdateDataPermission(ctx, data)
	return data, err
}

func (uc *DataPermissionUsecase) Delete(ctx context.Context, ids []string) error {
	return uc.repo.DeleteDataPermission(ctx, ids)
}

func (uc *DataPermissionUsecase) UpdateStatus(ctx context.Context, id, status string) error {
	return uc.repo.UpdateDataPermissionStatus(ctx, id, status)
}

// IsCodeExists 判断规则编码是否存在；excludeID 非空时排除该记录（编辑场景）
func (uc *DataPermissionUsecase) IsCodeExists(ctx context.Context, code, excludeID string) (bool, error) {
	return uc.repo.IsDataPermissionCodeExists(ctx, code, excludeID)
}

// validateCodeUnique 校验规则编码唯一性；excludeID 非空时排除自身（编辑场景）
func (uc *DataPermissionUsecase) validateCodeUnique(ctx context.Context, d *DataPermission, excludeID string) error {
	exists, err := uc.IsCodeExists(ctx, d.Code, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return errors.BadRequest("DATA_PERMISSION_CODE_EXISTS", "规则编码已存在")
	}
	return nil
}
