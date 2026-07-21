package service

import (
	"context"
	"encoding/json"

	adminV1 "github.com/antsurge/weaver-admin/api/gen/go/admin/service/v1"
	dictionaryV1 "github.com/antsurge/weaver-admin/api/gen/go/dictionary/service/v1"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data"
	"github.com/antsurge/weaver-admin/pkg/utils/copierx"
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/emptypb"
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
		output.Total = int64(res.Total)
		err = copierx.Copy(&output.Items, &res.Data)
	}

	return output, err
}

func (s *DictionaryService) CreateDictType(ctx context.Context, req *dictionaryV1.CreateDictTypeRequest) (*dictionaryV1.DictType, error) {
	input := biz.DictType{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}
	dictType, err := s.dictTypeUc.CreateDictType(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &dictionaryV1.DictType{}
	err = copier.Copy(output, dictType)
	if err != nil {
		return nil, err
	}

	return output, err
}

func (s *DictionaryService) UpdateDictType(ctx context.Context, req *dictionaryV1.UpdateDictTypeRequest) (*dictionaryV1.DictType, error) {
	input := biz.DictType{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}

	dictType, err := s.dictTypeUc.UpdateDictType(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &dictionaryV1.DictType{}
	err = copier.Copy(output, dictType)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (s *DictionaryService) UpdateDictTypeStatus(ctx context.Context, req *dictionaryV1.UpdateDictTypeStatusRequest) (*dictionaryV1.DictType, error) {
	err := s.dictTypeUc.UpdateDictTypeStatus(ctx, req.Id, req.Status)
	return nil, err
}

func (s *DictionaryService) DeleteDictType(ctx context.Context, req *dictionaryV1.DeleteDictTypeRequset) (*emptypb.Empty, error) {
	err := s.dictTypeUc.DeleteDictType(ctx, req.Ids)
	return nil, err
}

func (s *DictionaryService) ListDictData(ctx context.Context, req *dictionaryV1.ListDictDataRequest) (*dictionaryV1.ListDictDataResponse, error) {
	input := biz.ListDictDataRequest{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}

	res, err := s.dictDataUc.List(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &dictionaryV1.ListDictDataResponse{}
	if res != nil {
		output.Total = int64(res.Total)
		output.Items = make([]*dictionaryV1.DictData, 0, len(res.Data))
		for _, item := range res.Data {
			dataItem := &dictionaryV1.DictData{}
			err = copier.Copy(dataItem, item)
			if err != nil {
				return nil, err
			}

			dataItem.Extension = extensionToJSONString(item.Extension)
			output.Items = append(output.Items, dataItem)
		}
	}

	return output, nil
}

func (s *DictionaryService) CreateDictData(ctx context.Context, req *dictionaryV1.CreateDictDataRequest) (*dictionaryV1.DictData, error) {
	input := biz.DictData{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}

	input.Extension, err = data.ExtensionFromJSONString(req.Extension)
	if err != nil {
		return nil, err
	}

	dictData, err := s.dictDataUc.CreateDictData(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &dictionaryV1.DictData{}
	err = copier.Copy(output, dictData)
	if err != nil {
		return nil, err
	}

	output.Extension = extensionToJSONString(dictData.Extension)
	return output, nil
}

func (s *DictionaryService) UpdateDictData(ctx context.Context, req *dictionaryV1.UpdateDictDataRequest) (*dictionaryV1.DictData, error) {
	input := biz.DictData{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}

	input.Extension, err = data.ExtensionFromJSONString(req.Extension)
	if err != nil {
		return nil, err
	}

	dictData, err := s.dictDataUc.UpdateDictData(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &dictionaryV1.DictData{}
	err = copier.Copy(output, dictData)
	if err != nil {
		return nil, err
	}

	output.Extension = extensionToJSONString(dictData.Extension)
	return output, nil
}

func (s *DictionaryService) UpdateDictDataStatus(ctx context.Context, req *dictionaryV1.UpdateDictDataStatusRequest) (*dictionaryV1.DictData, error) {
	err := s.dictDataUc.UpdateDictDataStatus(ctx, req.Id, req.Status)
	return nil, err
}

func (s *DictionaryService) DeleteDictData(ctx context.Context, req *dictionaryV1.DeleteDictDataRequset) (*emptypb.Empty, error) {
	err := s.dictDataUc.DeleteDictData(ctx, req.Ids)
	return nil, err
}

func extensionToJSONString(value map[string]any) string {
	if len(value) == 0 {
		return ""
	}
	b, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(b)
}
