package service

import (
	"context"

	adminV1 "github.com/hypercoze/kratos-admin/api/gen/go/admin/service/v1"
	dictionaryV1 "github.com/hypercoze/kratos-admin/api/gen/go/dictionary/service/v1"
	organizationV1 "github.com/hypercoze/kratos-admin/api/gen/go/organization/service/v1"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/biz"
	"github.com/jinzhu/copier"
)

type DictionaryService struct {
	adminV1.UnimplementedDictionaryServer

	dictTypeUc *biz.DictTypeUsecase
	dictDataUc *biz.DictDataUsecase
}

func NewDictionaryService(
	dictTypeUc *biz.DictTypeUsecase,
	dictDataUc *biz.DictDataUsecase,
) *DictionaryService {
	return &DictionaryService{
		dictTypeUc: dictTypeUc,
		dictDataUc: dictDataUc,
	}
}

func (s *DictionaryService) ListDictType(ctx context.Context, req *dictionaryV1.ListDictTypeRequest) (*dictionaryV1.ListDictTypeResponse, error) {
	input := biz.ListDictTypeRequest{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}

	res, err := s.dictTypeUc.List(ctx, &input)
	if err != nil {
		return nil, err
	}

	// 填充data数据
	if err := s.dictTypeUc.FillDictData(ctx, res.Data...); err != nil {
		return nil, err
	}

	output := &dictionaryV1.ListDictTypeResponse{}
	if res != nil {
		err = copier.Copy(output, res)
	}

	return output, nil
}

func (s *DictionaryService) CreateDictType(ctx context.Context, req *dictionaryV1.CreateDictTypeRequest) (*dictionaryV1.DictType, error) {
	input := biz.Position{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}
	position, err := s.dictTypeUc.CreateDictType(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &dictionaryV1.DictType{}
	err = copier.Copy(output, position)
	return output, err
}
