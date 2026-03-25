package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/hypercoze/kratos-admin/pkg/enthelper"
	"github.com/hypercoze/kratos-admin/pkg/utils/uuid"
)

type DictType struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Code      string     `json:"code"`
	Status    string     `json:"status"`
	Remark    string     `json:"remark"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`

	DictData []*DictData `json:"dictData"`
}

type ListDictTypeRequest struct {
	enthelper.PaginationParams
	Name   string `form:"name" query:"name"`
	Code   string `form:"code" query:"code"`
	Status string `form:"status" query:"status"`
}

type ListDictTypeOption struct {
	enthelper.QueryOption
}

type ListDictTypeResponse struct {
	Data  []*DictType
	Total int
}

type DictTypeRepo interface {
	ListDictType(context.Context, *ListDictTypeRequest, ...*ListDictTypeOption) (*ListDictTypeResponse, error)
	CreateDictType(context.Context, *DictType) error
	UpdateDictType(context.Context, *DictType) error
	DeleteDictType(context.Context, []string) error
	UpdateDictTypeStatus(context.Context, string, string) error
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

func (uc *DictTypeUsecase) CreateDictType(ctx context.Context, req *DictType) (*DictType, error) {
	dictType := req
	dictType.ID = uuid.GenerateXID()
	dictType.CreatedAt = time.Now()
	dictType.UpdatedAt = time.Now()

	err := uc.repo.CreateDictType(ctx, dictType)
	return dictType, err
}

func (uc *DictTypeUsecase) UpdateDictType(ctx context.Context, req *DictType) (*DictType, error) {
	dictType := req
	dictType.UpdatedAt = time.Now()

	err := uc.repo.UpdateDictType(ctx, dictType)
	return dictType, err
}

func (uc *DictTypeUsecase) DeleteDictType(ctx context.Context, ids []string) error {
	return uc.repo.DeleteDictType(ctx, ids)
}

func (uc *DictTypeUsecase) UpdateDictTypeStatus(ctx context.Context, id, status string) error {
	return uc.repo.UpdateDictTypeStatus(ctx, id, status)
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
		key := item.DictTypeID
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
