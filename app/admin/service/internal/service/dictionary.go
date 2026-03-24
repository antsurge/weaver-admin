package service

import (
	"context"

	adminV1 "github.com/hypercoze/kratos-admin/api/gen/go/admin/service/v1"
	dictionaryV1 "github.com/hypercoze/kratos-admin/api/gen/go/dictionary/service/v1"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/biz"
	"github.com/jinzhu/copier"
)

type DictionaryService struct {
	adminV1.UnimplementedDictionaryServer

	dictTypeUc *biz.DictTypeUsecase
}

func NewDictionaryService(
	dictTypeUc *biz.DictTypeUsecase,
) *DictionaryService {
	return &DictionaryService{
		dictTypeUc: dictTypeUc,
	}
}

func (s *DictionaryService) ListDictType(ctx context.Context, req *dictionaryV1.ListDictTypeRequest) (*dictionaryV1.ListDictTypeResponse, error) {
	res, err := s.dictTypeUc.List(ctx, &biz.ListDictTypeRequest{})
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
