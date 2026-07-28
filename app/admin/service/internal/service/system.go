package service

import (
	"context"

	adminV1 "github.com/antsurge/weaver-admin/api/gen/go/admin/service/v1"
	systemV1 "github.com/antsurge/weaver-admin/api/gen/go/system/service/v1"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/pkg/utils/copierx"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

type SystemService struct {
	adminV1.UnimplementedSystemServer

	apiInterfaceUc *biz.ApiInterfaceUsecase
	log            *log.Helper
}

func NewSystemService(
	apiInterfaceUc *biz.ApiInterfaceUsecase,
	logger log.Logger,
) *SystemService {
	return &SystemService{
		apiInterfaceUc: apiInterfaceUc,
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

// ImportApiInterface 导入 openapi.yaml 文件
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
		Skipped:  int64(result.Skipped),
	}, nil
}

// DeleteApiInterface 批量删除接口
func (s *SystemService) DeleteApiInterface(ctx context.Context, req *systemV1.DeleteApiInterfaceRequest) (*emptypb.Empty, error) {
	err := s.apiInterfaceUc.Delete(ctx, req.Ids)
	return nil, err
}