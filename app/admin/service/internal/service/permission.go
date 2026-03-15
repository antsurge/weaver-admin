package service

import (
	"context"

	adminV1 "github.com/hypercoze/kratos-admin/api/gen/go/admin/service/v1"
	permissionV1 "github.com/hypercoze/kratos-admin/api/gen/go/permission/service/v1"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/biz"
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PermissionService struct {
	adminV1.UnimplementedPermissionServiceServer

	permissionUc *biz.PermissionUsecase
}

func NewPermissionService(
	permissionUc *biz.PermissionUsecase,
) *PermissionService {
	return &PermissionService{
		permissionUc: permissionUc,
	}
}

// 列表权限tree
func (s *PermissionService) PermissionTree(ctx context.Context, req *permissionV1.PermissionTreeRequest) (*permissionV1.PermissionTreeResponse, error) {
	tree, err := s.permissionUc.PermissionTree(ctx, req)
	if err != nil {
		return nil, err
	}
	output := make([]*permissionV1.Permission, 0)
	err = copier.Copy(&output, &tree)

	return &permissionV1.PermissionTreeResponse{Data: output}, nil
}

// 创建权限
func (s *PermissionService) CreatePermission(ctx context.Context, req *permissionV1.CreatePermissionRequest) (*permissionV1.CreatePermissionResponse, error) {
	input := &biz.Permission{}
	var err error
	err = copier.Copy(&input, &req)
	if err != nil {
		return nil, err
	}
	permission, err := s.permissionUc.CreatePermission(ctx, input)
	if err != nil {
		return nil, err
	}

	output := &permissionV1.Permission{}
	err = copier.Copy(&output, &permission)
	if err != nil {
		return nil, err
	}
	return &permissionV1.CreatePermissionResponse{Permission: output}, nil
}

// 更新权限
func (s *PermissionService) UpdatePermission(ctx context.Context, req *permissionV1.UpdatePermissionRequest) (*permissionV1.UpdatePermissionResponse, error) {
	input := &biz.Permission{}
	var err error
	err = copier.Copy(&input, &req)
	if err != nil {
		return nil, err
	}
	permission, err := s.permissionUc.UpdatePermission(ctx, input)

	output := &permissionV1.Permission{}
	err = copier.Copy(&output, &permission)
	if err != nil {
		return nil, err
	}
	return &permissionV1.UpdatePermissionResponse{Permission: output}, nil
}

// 删除数据
func (s *PermissionService) DeletePermission(ctx context.Context, req *permissionV1.DeletePermissionRequest) (*emptypb.Empty, error) {
	err := s.permissionUc.DeletePermission(ctx, req.Ids)
	return nil, err
}

// 更新权限状态
func (s *PermissionService) UpdatePermissionStatus(ctx context.Context, req *permissionV1.UpdatePermissionStatusRequest) (*emptypb.Empty, error) {
	err := s.permissionUc.UpdatePermissionStatus(ctx, req.Id, req.Status)
	return nil, err
}
