package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hypercoze/kratos-admin/pkg/enthelper"
)

type DictData struct {
	ID          string         `json:"id"`
	DictTypeId  string         `json:"dictTypeId"`
	Label       string         `json:"label"`
	Value       string         `json:"value"`
	Weight      int            `json:"weight"`
	Status      string         `json:"status"`
	Extension   map[string]any `json:"extension"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   *time.Time     `json:"deletedAt,omitempty"`
}

type ListDictDataRequest struct {
	enthelper.PaginationParams
	DictTypeIds []string `json:"dictTypeIds"`
}
type ListDictDataResponse struct {
	Data  []*DictData
	Total int
}

type DictDataRepo interface {
	List(context.Context, *ListDictDataRequest) (*ListDictDataResponse, error)
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
