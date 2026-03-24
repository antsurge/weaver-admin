package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hypercoze/kratos-admin/pkg/enthelper"
	"github.com/hypercoze/kratos-admin/pkg/utils/uuid"
)

type DictData struct {
	ID         string         `json:"id"`
	DictTypeID string         `json:"dictTypeID"`
	Label      string         `json:"label"`
	Value      string         `json:"value"`
	Weight     int            `json:"weight"`
	Status     string         `json:"status"`
	Extension  map[string]any `json:"extension"`
	Remark     string         `json:"remark"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  *time.Time     `json:"deletedAt,omitempty"`
}

type ListDictDataRequest struct {
	enthelper.PaginationParams
	Label       string   `form:"label" query:"label"`
	Value       string   `form:"value" query:"value"`
	Status      string   `form:"status" query:"status"`
	DictTypeIds []string `json:"dictTypeIds"`
}
type ListDictDataResponse struct {
	Data  []*DictData
	Total int
}

type DictDataRepo interface {
	List(context.Context, *ListDictDataRequest) (*ListDictDataResponse, error)
	Create(context.Context, *DictData) error
	Update(context.Context, *DictData) error
	Delete(context.Context, []string) error
	UpdateStatus(context.Context, string, string) error
}

type DictDataUsecase struct {
	repo DictDataRepo
	log  *log.Helper
}

func NewDictDataUsecase(repo DictDataRepo, logger log.Logger) *DictDataUsecase {
	return &DictDataUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *DictDataUsecase) List(ctx context.Context, req *ListDictDataRequest) (*ListDictDataResponse, error) {
	return uc.repo.List(ctx, req)
}

func (uc *DictDataUsecase) CreateDictData(ctx context.Context, req *DictData) (*DictData, error) {
	dictData := req
	dictData.ID = uuid.GenerateXID()
	dictData.CreatedAt = time.Now()
	dictData.UpdatedAt = time.Now()

	err := uc.repo.Create(ctx, dictData)
	return dictData, err
}

func (uc *DictDataUsecase) UpdateDictData(ctx context.Context, req *DictData) (*DictData, error) {
	dictData := req
	dictData.UpdatedAt = time.Now()

	err := uc.repo.Update(ctx, dictData)
	return dictData, err
}

func (uc *DictDataUsecase) DeleteDictData(ctx context.Context, ids []string) error {
	return uc.repo.Delete(ctx, ids)
}

func (uc *DictDataUsecase) UpdateDictDataStatus(ctx context.Context, id, status string) error {
	return uc.repo.UpdateStatus(ctx, id, status)
}
