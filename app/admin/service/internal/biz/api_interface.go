package biz

import (
	"context"
	"time"

	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/uuid"
	"github.com/go-kratos/kratos/v2/log"
)

// ApiInterface 从 openapi.yaml 导入的接口信息
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
	Total    int
	Imported int
	Skipped  int
}

// ApiInterfaceRepo 接口
type ApiInterfaceRepo interface {
	ListApiInterface(context.Context, *ListApiInterfaceRequest) (*ListApiInterfaceResponse, error)
	UpsertApiInterface(context.Context, *ApiInterface) error
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

// Import 批量导入（upsert by code）
func (uc *ApiInterfaceUsecase) Import(ctx context.Context, items []*ApiInterface) (*ImportResult, error) {
	result := &ImportResult{Total: len(items)}
	for _, item := range items {
		item.ID = uuid.GenerateXID()
		item.CreatedAt = time.Now()
		item.UpdatedAt = time.Now()
		err := uc.repo.UpsertApiInterface(ctx, item)
		if err != nil {
			return nil, err
		}
		result.Imported++
	}
	result.Skipped = result.Total - result.Imported
	return result, nil
}

// Delete 批量删除
func (uc *ApiInterfaceUsecase) Delete(ctx context.Context, ids []string) error {
	return uc.repo.DeleteApiInterface(ctx, ids)
}