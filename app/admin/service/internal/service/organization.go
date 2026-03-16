package service

import (
	"context"

	adminV1 "github.com/hypercoze/kratos-admin/api/gen/go/admin/service/v1"
	organizationV1 "github.com/hypercoze/kratos-admin/api/gen/go/organization/service/v1"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/biz"
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OrganizationService struct {
	adminV1.UnimplementedOrganizationServer

	organizationUc *biz.OrganizationUsecase
	positionUc     *biz.PositionUsecase
}

func NewOrganizationService(
	organizationUc *biz.OrganizationUsecase,
	positionUc *biz.PositionUsecase,
) *OrganizationService {
	return &OrganizationService{
		organizationUc: organizationUc,
		positionUc:     positionUc,
	}
}

// 组织tree
func (s *OrganizationService) OrganizationTree(ctx context.Context, req *organizationV1.OrganizationTreeRequest) (*organizationV1.OrganizationTreeResponse, error) {
	tree, err := s.organizationUc.OrganizationTree(ctx)
	if err != nil {
		return nil, err
	}
	output := make([]*organizationV1.Organization, 0)
	err = copier.Copy(&output, &tree)

	return &organizationV1.OrganizationTreeResponse{Data: output}, nil
}

// 创建
func (s *OrganizationService) CreateOrganization(ctx context.Context, req *organizationV1.CreateOrganizationRequest) (*organizationV1.Organization, error) {
	input := biz.Organization{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}
	organization, err := s.organizationUc.CreateOrganization(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &organizationV1.Organization{}
	err = copier.Copy(output, organization)

	return output, err
}

// 编辑
func (s *OrganizationService) UpdateOrganization(ctx context.Context, req *organizationV1.UpdateOrganizationRequest) (*organizationV1.Organization, error) {
	input := biz.Organization{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}
	organization, err := s.organizationUc.UpdateOrganization(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &organizationV1.Organization{}
	err = copier.Copy(output, organization)

	return output, err
}

// 删除
func (s *OrganizationService) DeleteOrganization(ctx context.Context, req *organizationV1.DeleteOrganizationRequset) (*emptypb.Empty, error) {
	err := s.organizationUc.DeleteOrganization(ctx, req.Ids)
	return nil, err
}

// 更新状态
func (s *OrganizationService) UpdateOrganizationStatus(ctx context.Context, req *organizationV1.UpdateOrganizationStatusRequest) (*organizationV1.Organization, error) {
	err := s.organizationUc.UpdateOrganizationStatus(ctx, req.Id, req.Status)
	return nil, err
}

// 岗位列表
func (s *OrganizationService) ListPosition(ctx context.Context, req *organizationV1.ListPositionRequest) (*organizationV1.ListPositionResponse, error) {
	list, err := s.positionUc.ListPosition(ctx, &biz.ListPositionRequest{
		PaginationParam: biz.PaginationParam{
			Page:     1,
			PageSize: 10,
		},
	})
	if err != nil {
		return nil, err
	}

	output := &organizationV1.ListPositionResponse{}
	if list != nil {
		err = copier.Copy(output, list)
	}

	return output, err
}

func (s *OrganizationService) CreatePosition(ctx context.Context, req *organizationV1.CreatePositionRequest) (*organizationV1.Position, error) {
	input := biz.Position{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}
	position, err := s.positionUc.CreatePosition(ctx, &input)
	if err != nil {
		return nil, err
	}
	output := &organizationV1.Position{}
	err = copier.Copy(output, position)
	return output, err
}

func (s *OrganizationService) UpdatePosition(ctx context.Context, req *organizationV1.UpdatePositionRequest) (*organizationV1.Position, error) {
	input := biz.Position{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}

	position, err := s.positionUc.UpdatePosition(ctx, &input)
	if err != nil {
		return nil, err
	}
	output := &organizationV1.Position{}
	err = copier.Copy(output, position)
	return output, err
}
