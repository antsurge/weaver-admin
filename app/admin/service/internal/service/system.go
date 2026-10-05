package service

import (
	"context"
	"encoding/json"

	adminV1 "github.com/antsurge/weaver-admin/api/gen/go/admin/service/v1"
	systemV1 "github.com/antsurge/weaver-admin/api/gen/go/system/service/v1"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data"
	"github.com/antsurge/weaver-admin/pkg/utils/copierx"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/emptypb"
)

type SystemService struct {
	adminV1.UnimplementedSystemServer

	apiInterfaceUc *biz.ApiInterfaceUsecase
	dictTypeUc     *biz.DictTypeUsecase
	dictDataUc     *biz.DictDataUsecase
	configUc       *biz.ConfigUsecase
	log            *log.Helper
}

func NewSystemService(
	apiInterfaceUc *biz.ApiInterfaceUsecase,
	dictTypeUc *biz.DictTypeUsecase,
	dictDataUc *biz.DictDataUsecase,
	configUc *biz.ConfigUsecase,
	logger log.Logger,
) *SystemService {
	return &SystemService{
		apiInterfaceUc: apiInterfaceUc,
		dictTypeUc:     dictTypeUc,
		dictDataUc:     dictDataUc,
		configUc:       configUc,
		log:            log.NewHelper(logger),
	}
}

// ListApiInterface 分页查询接口列表
func (s *SystemService) ListApiInterface(ctx context.Context, req *systemV1.ListApiInterfaceRequest) (*systemV1.ListApiInterfaceResponse, error) {
	input := &biz.ListApiInterfaceRequest{}
	var err error
	err = copierx.Copy(input, req)
	if err != nil {
		return nil, err
	}
	if req.CurrentPage != 0 {
		input.CurrentPage = int(req.CurrentPage)
	}
	if req.PageSize != 0 {
		input.PageSize = int(req.PageSize)
	}

	res, err := s.apiInterfaceUc.List(ctx, input)
	if err != nil {
		return nil, err
	}

	output := &systemV1.ListApiInterfaceResponse{}
	if res != nil {
		output.Total = int64(res.Total)
		output.Items = make([]*systemV1.ApiInterface, 0, len(res.Data))
		for _, item := range res.Data {
			pbItem := &systemV1.ApiInterface{}
			if err := copierx.Copy(pbItem, item); err != nil {
				return nil, err
			}
			output.Items = append(output.Items, pbItem)
		}
	}

	return output, nil
}

// ListApiInterfaceOptions 接口筛选选项（服务名/标签）
func (s *SystemService) ListApiInterfaceOptions(ctx context.Context, req *systemV1.ListApiInterfaceOptionsRequest) (*systemV1.ListApiInterfaceOptionsResponse, error) {
	res, err := s.apiInterfaceUc.ListOptions(ctx)
	if err != nil {
		return nil, err
	}

	output := &systemV1.ListApiInterfaceOptionsResponse{}
	if res != nil {
		output.Services = res.Services
		output.Tags = res.Tags
	}
	return output, nil
}

// CreateApiInterface 手动新增接口
func (s *SystemService) CreateApiInterface(ctx context.Context, req *systemV1.CreateApiInterfaceRequest) (*systemV1.ApiInterface, error) {
	input := &biz.ApiInterface{}
	if err := copierx.Copy(input, req); err != nil {
		return nil, err
	}

	res, err := s.apiInterfaceUc.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	output := &systemV1.ApiInterface{}
	if err := copierx.Copy(output, res); err != nil {
		return nil, err
	}
	return output, nil
}

// UpdateApiInterface 手动编辑接口
func (s *SystemService) UpdateApiInterface(ctx context.Context, req *systemV1.UpdateApiInterfaceRequest) (*systemV1.ApiInterface, error) {
	input := &biz.ApiInterface{}
	if err := copierx.Copy(input, req); err != nil {
		return nil, err
	}

	res, err := s.apiInterfaceUc.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	output := &systemV1.ApiInterface{}
	if err := copierx.Copy(output, res); err != nil {
		return nil, err
	}
	return output, nil
}

// ImportApiInterface 导入 openapi.yaml 文件（按 code 对比：已存在更新、不存在创建）
func (s *SystemService) ImportApiInterface(ctx context.Context, req *systemV1.ImportApiInterfaceRequest) (*systemV1.ImportApiInterfaceResponse, error) {
	// 解析文件内容
	items, err := biz.ParseOpenAPIFile(req.File)
	if err != nil {
		return nil, err
	}

	// 批量导入
	result, err := s.apiInterfaceUc.Import(ctx, items)
	if err != nil {
		return nil, err
	}

	return &systemV1.ImportApiInterfaceResponse{
		Total:    int64(result.Total),
		Imported: int64(result.Imported),
		Updated:  int64(result.Updated),
		Skipped:  int64(result.Skipped),
	}, nil
}

// DeleteApiInterface 批量删除接口
func (s *SystemService) DeleteApiInterface(ctx context.Context, req *systemV1.DeleteApiInterfaceRequest) (*emptypb.Empty, error) {
	err := s.apiInterfaceUc.Delete(ctx, req.Ids)
	return nil, err
}

// ====== 字典 ======

func (s *SystemService) ListDictType(ctx context.Context, req *systemV1.ListDictTypeRequest) (*systemV1.ListDictTypeResponse, error) {
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

	output := &systemV1.ListDictTypeResponse{}
	if res != nil {
		output.Total = int64(res.Total)
		err = copierx.Copy(&output.Items, &res.Data)
	}

	return output, err
}

func (s *SystemService) CreateDictType(ctx context.Context, req *systemV1.CreateDictTypeRequest) (*systemV1.DictType, error) {
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

	output := &systemV1.DictType{}
	err = copier.Copy(output, dictType)
	if err != nil {
		return nil, err
	}

	return output, err
}

func (s *SystemService) UpdateDictType(ctx context.Context, req *systemV1.UpdateDictTypeRequest) (*systemV1.DictType, error) {
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

	output := &systemV1.DictType{}
	err = copier.Copy(output, dictType)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (s *SystemService) UpdateDictTypeStatus(ctx context.Context, req *systemV1.UpdateDictTypeStatusRequest) (*systemV1.DictType, error) {
	err := s.dictTypeUc.UpdateDictTypeStatus(ctx, req.Id, req.Status)
	return nil, err
}

func (s *SystemService) DeleteDictType(ctx context.Context, req *systemV1.DeleteDictTypeRequset) (*emptypb.Empty, error) {
	err := s.dictTypeUc.DeleteDictType(ctx, req.Ids)
	return nil, err
}

// GetDictTypeByCode 根据字典编码获取字典类型及字典数据（供业务页面消费）
func (s *SystemService) GetDictTypeByCode(ctx context.Context, req *systemV1.GetDictTypeByCodeRequest) (*systemV1.DictType, error) {
	dictType, err := s.dictTypeUc.GetDictTypeByCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}
	if dictType == nil {
		return nil, nil
	}

	output := &systemV1.DictType{}
	if err = copierx.Copy(output, dictType); err != nil {
		return nil, err
	}

	// DictData 需要手动转换（Extension: map -> JSON string）
	if len(dictType.DictData) > 0 {
		output.DictData = make([]*systemV1.DictData, 0, len(dictType.DictData))
		for _, item := range dictType.DictData {
			dataItem := &systemV1.DictData{}
			if err = copier.Copy(dataItem, item); err != nil {
				return nil, err
			}
			dataItem.Extension = extensionToJSONString(item.Extension)
			output.DictData = append(output.DictData, dataItem)
		}
	}

	return output, nil
}

func (s *SystemService) ListDictData(ctx context.Context, req *systemV1.ListDictDataRequest) (*systemV1.ListDictDataResponse, error) {
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

	output := &systemV1.ListDictDataResponse{}
	if res != nil {
		output.Total = int64(res.Total)
		output.Items = make([]*systemV1.DictData, 0, len(res.Data))
		for _, item := range res.Data {
			dataItem := &systemV1.DictData{}
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

func (s *SystemService) CreateDictData(ctx context.Context, req *systemV1.CreateDictDataRequest) (*systemV1.DictData, error) {
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

	output := &systemV1.DictData{}
	err = copier.Copy(output, dictData)
	if err != nil {
		return nil, err
	}

	output.Extension = extensionToJSONString(dictData.Extension)
	return output, nil
}

func (s *SystemService) UpdateDictData(ctx context.Context, req *systemV1.UpdateDictDataRequest) (*systemV1.DictData, error) {
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

	output := &systemV1.DictData{}
	err = copier.Copy(output, dictData)
	if err != nil {
		return nil, err
	}

	output.Extension = extensionToJSONString(dictData.Extension)
	return output, nil
}

func (s *SystemService) UpdateDictDataStatus(ctx context.Context, req *systemV1.UpdateDictDataStatusRequest) (*systemV1.DictData, error) {
	err := s.dictDataUc.UpdateDictDataStatus(ctx, req.Id, req.Status)
	return nil, err
}

func (s *SystemService) DeleteDictData(ctx context.Context, req *systemV1.DeleteDictDataRequset) (*emptypb.Empty, error) {
	err := s.dictDataUc.DeleteDictData(ctx, req.Ids)
	return nil, err
}

// ====== 参数设置 ======

func (s *SystemService) ListConfig(ctx context.Context, req *systemV1.ListConfigRequest) (*systemV1.ListConfigResponse, error) {
	input := biz.ListConfigRequest{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}

	res, err := s.configUc.List(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &systemV1.ListConfigResponse{}
	if res != nil {
		output.Total = int64(res.Total)
		err = copierx.Copy(&output.Items, &res.Data)
	}

	return output, err
}

func (s *SystemService) GetConfigByKey(ctx context.Context, req *systemV1.GetConfigByKeyRequest) (*systemV1.Config, error) {
	config, err := s.configUc.GetConfigByKey(ctx, req.Key)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, nil
	}

	output := &systemV1.Config{}
	err = copierx.Copy(output, config)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (s *SystemService) CreateConfig(ctx context.Context, req *systemV1.CreateConfigRequest) (*systemV1.Config, error) {
	input := biz.SysConfig{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}

	config, err := s.configUc.CreateConfig(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &systemV1.Config{}
	err = copierx.Copy(output, config)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (s *SystemService) UpdateConfig(ctx context.Context, req *systemV1.UpdateConfigRequest) (*systemV1.Config, error) {
	input := biz.SysConfig{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}

	config, err := s.configUc.UpdateConfig(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &systemV1.Config{}
	err = copierx.Copy(output, config)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (s *SystemService) UpdateConfigStatus(ctx context.Context, req *systemV1.UpdateConfigStatusRequest) (*emptypb.Empty, error) {
	err := s.configUc.UpdateConfigStatus(ctx, req.Id, req.Status)
	return nil, err
}

func (s *SystemService) DeleteConfig(ctx context.Context, req *systemV1.DeleteConfigRequest) (*emptypb.Empty, error) {
	err := s.configUc.DeleteConfig(ctx, req.Ids)
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
