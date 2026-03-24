package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hypercoze/kratos-admin/pkg/enthelper"
)

type DictType struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Code        string     `json:"code"`
	Status      string     `json:"status"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`

	DictData []*DictData `json:"dictData"`
}

type ListDictTypeRequest struct {
	enthelper.PaginationParams
}

type ListDictTypeResponse struct {
	Data  []*DictType
	Total int
}

type DictTypeRepo interface {
	ListDictType(context.Context, *ListDictTypeRequest) (*ListDictTypeResponse, error)
	CreateDictType(context.Context, *DictType) error
}

type DictTypeUsecase struct {
	repo         DictTypeRepo
	dictDataRepo DictDataRepo
	log          *log.Helper
}

func NewDictTypeUsecase(
	repo DictTypeRepo,
	dictDataRepo DictDataRepo,
	logger log.Logger,
) *DictTypeUsecase {
	return &DictTypeUsecase{
		repo:         repo,
		dictDataRepo: dictDataRepo,
		log:          log.NewHelper(logger),
	}
}

func (uc *DictTypeUsecase) List(ctx context.Context, req *ListDictTypeRequest) (*ListDictTypeResponse, error) {
	return uc.repo.ListDictType(ctx, req)
}

// 填充字典明细数据
func (uc *DictTypeUsecase) FillDictData(ctx context.Context, items ...*DictType) error {
	if len(items) == 0 {
		return nil
	}
	ids := make([]string, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}

	dataRes, err := uc.dictDataRepo.List(ctx, &ListDictDataRequest{
		DictTypeIds: ids,
	})
	if err != nil {
		return err
	}
	if len(dataRes.Data) == 0 {
		return nil
	}

	// 组装成map
	dataMap := make(map[string][]*DictData)
	for _, item := range dataRes.Data {
		key := item.DictTypeId
		if _, ok := dataMap[key]; !ok {
			dataMap[key] = make([]*DictData, 0)
		}
		dataMap[key] = append(dataMap[key], item)
	}

	for _, item := range items {
		if data, ok := dataMap[item.ID]; ok {
			item.DictData = data
		}
	}

	return nil
}
