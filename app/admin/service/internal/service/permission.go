package service

import (
	"context"

	adminV1 "github.com/hypercoze/kratos-admin/api/gen/go/admin/service/v1"
	permissionV1 "github.com/hypercoze/kratos-admin/api/gen/go/permission/service/v1"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/biz"
	"github.com/hypercoze/kratos-admin/pkg/utils/copierx"
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PermissionService struct {
	adminV1.UnimplementedPermissionServiceServer

	menuUc *biz.MenuUsecase
	roleUc *biz.RoleUsecase
}

func NewPermissionService(
	menuUc *biz.MenuUsecase,
	roleUc *biz.RoleUsecase,
) *PermissionService {
	return &PermissionService{
		menuUc: menuUc,
		roleUc: roleUc,
	}
}

// 列表权限tree
func (s *PermissionService) MenuTree(ctx context.Context, req *permissionV1.MenuTreeRequest) (*permissionV1.MenuTreeResponse, error) {
	input := &biz.ListMenuRequest{}
	var err error
	err = copier.Copy(&input, &req)

	tree, err := s.menuUc.MenuTree(ctx, input)
	if err != nil {
		return nil, err
	}
	output := make([]*permissionV1.Menu, 0)
	err = copier.Copy(&output, &tree)

	return &permissionV1.MenuTreeResponse{Items: output}, nil
}

// 创建权限
func (s *PermissionService) CreateMenu(ctx context.Context, req *permissionV1.CreateMenuRequest) (*permissionV1.Menu, error) {
	input := &biz.Menu{}
	var err error
	err = copier.Copy(&input, &req)
	if err != nil {
		return nil, err
	}
	permission, err := s.menuUc.CreateMenu(ctx, input)
	if err != nil {
		return nil, err
	}

	output := &permissionV1.Menu{}
	err = copier.Copy(&output, &permission)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// 更新权限
func (s *PermissionService) UpdateMenu(ctx context.Context, req *permissionV1.UpdateMenuRequest) (*permissionV1.Menu, error) {
	input := &biz.Menu{}
	var err error
	err = copier.Copy(&input, &req)
	if err != nil {
		return nil, err
	}
	permission, err := s.menuUc.UpdateMenu(ctx, input)
	if err != nil {
		return nil, err
	}

	output := &permissionV1.Menu{}
	err = copier.Copy(&output, &permission)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// 删除数据
func (s *PermissionService) DeleteMenu(ctx context.Context, req *permissionV1.DeleteMenuRequest) (*emptypb.Empty, error) {
	err := s.menuUc.DeleteMenu(ctx, req.Ids)
	return nil, err
}

// 更新权限状态
func (s *PermissionService) UpdateMenuStatus(ctx context.Context, req *permissionV1.UpdateMenuStatusRequest) (*emptypb.Empty, error) {
	err := s.menuUc.UpdateMenuStatus(ctx, req.Id, req.Status)
	return nil, err
}

// 角色
func (s *PermissionService) ListRole(ctx context.Context, req *permissionV1.ListRoleRequest) (*permissionV1.ListRoleResponse, error) {
	input := &biz.ListRoleRequest{}
	var err error
	err = copier.Copy(&input, &req)
	if err != nil {
		return nil, err
	}

	list, err := s.roleUc.ListRole(ctx, input)
	if err != nil {
		return nil, err
	}
	output := &permissionV1.ListRoleResponse{}
	if list != nil {
		err = copierx.Copy(output, list)
	}

	return output, nil
}

func (s *PermissionService) GetRole(ctx context.Context, req *permissionV1.GetRoleRequest) (*permissionV1.Role, error) {
	role, err := s.roleUc.GetRole(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	output := &permissionV1.Role{}
	err = copierx.Copy(output, role)

	return output, nil
}

func (s *PermissionService) CreateRole(ctx context.Context, req *permissionV1.CreateRoleRequest) (*permissionV1.Role, error) {
	input := &biz.Role{}
	var err error
	err = copier.Copy(&input, &req)
	if err != nil {
		return nil, err
	}
	role, err := s.roleUc.CreateRole(ctx, input)
	if err != nil {
		return nil, err
	}
	output := &permissionV1.Role{}
	err = copierx.Copy(output, role)
	return output, err
}

func (s *PermissionService) UpdateRole(ctx context.Context, req *permissionV1.UpdateRoleRequest) (*permissionV1.Role, error) {
	input := &biz.Role{}
	var err error
	err = copier.Copy(&input, &req)
	if err != nil {
		return nil, err
	}
	role, err := s.roleUc.UpdateRole(ctx, input)
	if err != nil {
		return nil, err
	}
	output := &permissionV1.Role{}
	err = copierx.Copy(output, role)
	return output, err
}

func (s *PermissionService) UpdateRoleStatus(ctx context.Context, req *permissionV1.UpdateRoleStatusRequest) (*emptypb.Empty, error) {
	err := s.roleUc.UpdateRoleStatus(ctx, req.Id, req.Status)
	return nil, err
}

func (s *PermissionService) DeleteRole(ctx context.Context, req *permissionV1.DeleteRoleRequest) (*emptypb.Empty, error) {
	err := s.roleUc.DeleteRole(ctx, req.Ids)
	return nil, err
}
