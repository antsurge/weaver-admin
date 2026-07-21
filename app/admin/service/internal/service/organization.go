package service

import (
	"context"

	adminV1 "github.com/antsurge/weaver-admin/api/gen/go/admin/service/v1"
	organizationV1 "github.com/antsurge/weaver-admin/api/gen/go/organization/service/v1"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/pkg/enthelper"
	"github.com/antsurge/weaver-admin/pkg/utils/copierx"
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OrganizationService struct {
	adminV1.UnimplementedOrganizationServer

	departmentUc *biz.DepartmentUsecase
	positionUc   *biz.PositionUsecase
}

func NewOrganizationService(
	departmentUc *biz.DepartmentUsecase,
	positionUc *biz.PositionUsecase,
) *OrganizationService {
	return &OrganizationService{
		departmentUc: departmentUc,
		positionUc:   positionUc,
	}
}

// 部门tree
func (s *OrganizationService) DepartmentTree(ctx context.Context, req *organizationV1.DepartmentTreeRequest) (*organizationV1.DepartmentTreeResponse, error) {
	input := biz.DepartmentListResult{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}

	tree, err := s.departmentUc.DepartmentTree(ctx, &input)
	if err != nil {
		return nil, err
	}
	output := make([]*organizationV1.Department, 0)
	err = copier.Copy(&output, &tree)

	return &organizationV1.DepartmentTreeResponse{Items: output}, nil
}

func (s *OrganizationService) GetDepartment(ctx context.Context, req *organizationV1.GetDepartmentRequest) (*organizationV1.Department, error) {
	department, err := s.departmentUc.GetDepartment(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	output := &organizationV1.Department{}
	err = copier.Copy(&output, &department)

	return output, err
}

// 创建
func (s *OrganizationService) CreateDepartment(ctx context.Context, req *organizationV1.CreateDepartmentRequest) (*organizationV1.Department, error) {
	input := biz.Department{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}
	departmet, err := s.departmentUc.CreateDepartment(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &organizationV1.Department{}
	err = copier.Copy(output, departmet)

	return output, err
}

// 编辑
func (s *OrganizationService) UpdateDepartment(ctx context.Context, req *organizationV1.UpdateDepartmentRequest) (*organizationV1.Department, error) {
	input := biz.Department{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}
	departmet, err := s.departmentUc.UpdateDepartment(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &organizationV1.Department{}
	err = copier.Copy(output, departmet)

	return output, err
}

// 删除
func (s *OrganizationService) DeleteDepartment(ctx context.Context, req *organizationV1.DeleteDepartmentRequset) (*emptypb.Empty, error) {
	err := s.departmentUc.DeleteDepartment(ctx, req.Ids)
	return nil, err
}

// 更新状态
func (s *OrganizationService) UpdateDepartmentStatus(ctx context.Context, req *organizationV1.UpdateDepartmentStatusRequest) (*organizationV1.Department, error) {
	err := s.departmentUc.UpdateDepartmentStatus(ctx, req.Id, req.Status)
	return nil, err
}

// 岗位列表
func (s *OrganizationService) ListPosition(ctx context.Context, req *organizationV1.ListPositionRequest) (*organizationV1.ListPositionResponse, error) {
	input := biz.ListPositionRequest{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}

	opts := &biz.ListPositionOption{}
	opts.Sorts = enthelper.ConvertSorts(req.Sorts)

	list, err := s.positionUc.ListPosition(ctx, &input, opts)
	if err != nil {
		return nil, err
	}

	output := &organizationV1.ListPositionResponse{}
	if list != nil {
		err = copierx.Copy(output, list)
	}

	return output, err
}

// 获取岗位
func (s *OrganizationService) GetPosition(ctx context.Context, req *organizationV1.GetPositionRequest) (*organizationV1.Position, error) {
	position, err := s.positionUc.GetPosition(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	output := &organizationV1.Position{}
	err = copier.Copy(output, position)
	return output, err
}

// 岗位创建
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

// 岗位编辑
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

// 岗位更新状态
func (s *OrganizationService) UpdatePositionStatus(ctx context.Context, req *organizationV1.UpdatePositionStatusRequest) (*emptypb.Empty, error) {
	err := s.positionUc.UpdatePositionStatus(ctx, req.Id, req.Status)
	return nil, err
}

// 岗位删除
func (s *OrganizationService) DeletePosition(ctx context.Context, req *organizationV1.DeletePositionRequest) (*emptypb.Empty, error) {
	err := s.positionUc.DeletePosition(ctx, req.Ids)
	return nil, err
}

func (s *OrganizationService) IsPositionNameExists(ctx context.Context, req *organizationV1.IsPositionNameExistsRequest) (*organizationV1.IsPositionFieldExistsResponse, error) {
	exists, err := s.positionUc.IsPositionNameExists(ctx, req.Name, req.Id)
	if err != nil {
		return nil, err
	}
	return &organizationV1.IsPositionFieldExistsResponse{Exists: exists}, nil
}

func (s *OrganizationService) IsPositionCodeExists(ctx context.Context, req *organizationV1.IsPositionCodeExistsRequest) (*organizationV1.IsPositionFieldExistsResponse, error) {
	exists, err := s.positionUc.IsPositionCodeExists(ctx, req.Code, req.Id)
	if err != nil {
		return nil, err
	}
	return &organizationV1.IsPositionFieldExistsResponse{Exists: exists}, nil
}
