package biz

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

// 错误定义
var (
	// ErrApiInterfaceExists 接口已存在（service|method|path 相同，即 code 相同）
	ErrApiInterfaceExists = errors.BadRequest("API_INTERFACE_EXISTS", "接口已存在，请检查服务名/方法/路径")
)

// ApiInterface 接口信息
type ApiInterface struct {
	ID        string    `json:"id"`
	Service   string    `json:"service"`
	Tag       string    `json:"tag"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	Summary   string    `json:"summary"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ListApiInterfaceRequest 分页查询请求
type ListApiInterfaceRequest struct {
	enthelper.PaginationParams
	Service string `form:"service" query:"service"`
	Tag     string `form:"tag" query:"tag"`
	Method  string `form:"method" query:"method"`
	Path    string `form:"path" query:"path"`
	Summary string `form:"summary" query:"summary"`
}

// ListApiInterfaceResponse 分页查询响应
type ListApiInterfaceResponse struct {
	Data  []*ApiInterface
	Total int
}

// ImportResult 导入结果
type ImportResult struct {
	Total    int `json:"total"`
	Imported int `json:"imported"`
	Updated  int `json:"updated"`
	Skipped  int `json:"skipped"`
}

// ApiInterfaceOptions 筛选选项（下拉数据源）
type ApiInterfaceOptions struct {
	Services []string
	Tags     []string
}

// ApiInterfaceRepo 接口
type ApiInterfaceRepo interface {
	ListApiInterface(context.Context, *ListApiInterfaceRequest) (*ListApiInterfaceResponse, error)
	ListApiInterfaceOptions(context.Context) (*ApiInterfaceOptions, error)
	GetApiInterfaceByCode(context.Context, string) (*ApiInterface, error)
	CreateApiInterface(context.Context, *ApiInterface) (*ApiInterface, error)
	UpdateApiInterface(context.Context, *ApiInterface) (*ApiInterface, error)
	// UpsertApiInterfaces 按 code 对比导入：已存在的更新、不存在的创建，返回导入结果
	UpsertApiInterfaces(context.Context, []*ApiInterface) (*ImportResult, error)
	DeleteApiInterface(context.Context, []string) error
}

// ApiInterfaceUsecase 用例
type ApiInterfaceUsecase struct {
	repo ApiInterfaceRepo
	log  *log.Helper
}

// NewApiInterfaceUsecase 创建用例
func NewApiInterfaceUsecase(repo ApiInterfaceRepo, logger log.Logger) *ApiInterfaceUsecase {
	return &ApiInterfaceUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// List 分页查询
func (uc *ApiInterfaceUsecase) List(ctx context.Context, req *ListApiInterfaceRequest) (*ListApiInterfaceResponse, error) {
	return uc.repo.ListApiInterface(ctx, req)
}

// ListOptions 查询去重后的服务名/标签选项
func (uc *ApiInterfaceUsecase) ListOptions(ctx context.Context) (*ApiInterfaceOptions, error) {
	return uc.repo.ListApiInterfaceOptions(ctx)
}

// Create 手动新增接口
func (uc *ApiInterfaceUsecase) Create(ctx context.Context, item *ApiInterface) (*ApiInterface, error) {
	if err := validateApiInterface(item); err != nil {
		return nil, err
	}
	item.Code = buildApiInterfaceCode(item.Service, item.Method, item.Path)

	// 唯一键冲突检查
	exist, err := uc.repo.GetApiInterfaceByCode(ctx, item.Code)
	if err != nil {
		return nil, err
	}
	if exist != nil {
		return nil, ErrApiInterfaceExists
	}

	now := time.Now()
	item.ID = uuid.GenerateXID()
	item.CreatedAt = now
	item.UpdatedAt = now
	return uc.repo.CreateApiInterface(ctx, item)
}

// Update 手动编辑接口
func (uc *ApiInterfaceUsecase) Update(ctx context.Context, item *ApiInterface) (*ApiInterface, error) {
	if err := validateApiInterface(item); err != nil {
		return nil, err
	}
	if item.ID == "" {
		return nil, errors.BadRequest("API_INTERFACE_REQUIRED", "接口 ID 不能为空")
	}
	item.Code = buildApiInterfaceCode(item.Service, item.Method, item.Path)

	// 唯一键冲突检查（排除自身）
	exist, err := uc.repo.GetApiInterfaceByCode(ctx, item.Code)
	if err != nil {
		return nil, err
	}
	if exist != nil && exist.ID != item.ID {
		return nil, ErrApiInterfaceExists
	}

	item.UpdatedAt = time.Now()
	return uc.repo.UpdateApiInterface(ctx, item)
}

// Import 按 code 比对导入：已存在的更新、不存在的创建（不删除任何已有数据）。
// 手动新增的接口不受影响。
func (uc *ApiInterfaceUsecase) Import(ctx context.Context, items []*ApiInterface) (*ImportResult, error) {
	for _, item := range items {
		item.ID = uuid.GenerateXID()
		item.Code = buildApiInterfaceCode(item.Service, item.Method, item.Path)
		item.CreatedAt = time.Now()
		item.UpdatedAt = time.Now()
	}
	return uc.repo.UpsertApiInterfaces(ctx, items)
}

// Delete 批量删除
func (uc *ApiInterfaceUsecase) Delete(ctx context.Context, ids []string) error {
	return uc.repo.DeleteApiInterface(ctx, ids)
}

// validateApiInterface 校验手动新增/编辑的必填字段
func validateApiInterface(item *ApiInterface) error {
	if item.Service == "" {
		return errors.BadRequest("API_INTERFACE_INVALID", "服务名不能为空")
	}
	if item.Method == "" {
		return errors.BadRequest("API_INTERFACE_INVALID", "HTTP 方法不能为空")
	}
	if item.Path == "" {
		return errors.BadRequest("API_INTERFACE_INVALID", "接口路径不能为空")
	}
	return nil
}

// buildApiInterfaceCode 生成业务唯一键：service|METHOD|path
func buildApiInterfaceCode(service, method, path string) string {
	return fmt.Sprintf("%s|%s|%s", service, strings.ToUpper(method), path)
}
