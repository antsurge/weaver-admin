package service

import (
	"context"

	adminV1 "github.com/antsurge/weaver-admin/api/gen/go/admin/service/v1"
	permissionV1 "github.com/antsurge/weaver-admin/api/gen/go/permission/service/v1"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/openapi_scanner"
	"github.com/antsurge/weaver-admin/pkg/utils/copierx"
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PermissionService struct {
	adminV1.UnimplementedPermissionServiceServer

	menuUc      *biz.MenuUsecase
	roleUc      *biz.RoleUsecase
	apiMetadata *openapi_scanner.Service
}

func NewPermissionService(
	menuUc *biz.MenuUsecase,
	roleUc *biz.RoleUsecase,
	apiMetadata *openapi_scanner.Service,
) *PermissionService {
	return &PermissionService{
		menuUc:      menuUc,
		roleUc:      roleUc,
		apiMetadata: apiMetadata,
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

// ====== 角色相关方法 ======

// 角色列表
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

// 获取角色详情（包含菜单ID列表）
func (s *PermissionService) GetRole(ctx context.Context, req *permissionV1.GetRoleRequest) (*permissionV1.Role, error) {
	// 使用 GetRoleWithMenus 获取带菜单的角色详情
	role, err := s.roleUc.GetRoleWithMenus(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	output := &permissionV1.Role{}
	err = copierx.Copy(output, role)

	return output, err
}

// 创建角色（支持同时绑定菜单）
func (s *PermissionService) CreateRole(ctx context.Context, req *permissionV1.CreateRoleRequest) (*permissionV1.Role, error) {
	input := &biz.Role{}
	var err error
	err = copier.Copy(&input, &req)
	if err != nil {
		return nil, err
	}

	// 如果有菜单ID，使用 CreateRoleWithMenus
	if len(input.MenuIDs) > 0 {
		role, err := s.roleUc.CreateRoleWithMenus(ctx, input)
		if err != nil {
			return nil, err
		}
		output := &permissionV1.Role{}
		err = copierx.Copy(output, role)
		return output, err
	}

	// 普通创建（不绑定菜单）
	role, err := s.roleUc.CreateRole(ctx, input)
	if err != nil {
		return nil, err
	}
	output := &permissionV1.Role{}
	err = copierx.Copy(output, role)
	return output, err
}

// 更新角色（支持同时重新绑定菜单）
func (s *PermissionService) UpdateRole(ctx context.Context, req *permissionV1.UpdateRoleRequest) (*permissionV1.Role, error) {
	input := &biz.Role{}
	var err error
	err = copier.Copy(&input, &req)
	if err != nil {
		return nil, err
	}

	// 使用 UpdateRoleWithMenus 更新角色并重新绑定菜单
	role, err := s.roleUc.UpdateRoleWithMenus(ctx, input)
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

// ====== RBAC 角色菜单绑定方法 ======

// BindMenusForRole 为角色绑定菜单（全量替换）
func (s *PermissionService) BindMenusForRole(
	ctx context.Context,
	req *permissionV1.BindMenusForRoleRequest,
) (*emptypb.Empty, error) {
	err := s.roleUc.BindMenusForRole(ctx, req.RoleId, req.MenuIds)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ListMenusByRole 查询角色的菜单树
func (s *PermissionService) ListMenusByRole(
	ctx context.Context,
	req *permissionV1.ListMenusByRoleRequest,
) (*permissionV1.ListMenusByRoleResponse, error) {
	menus, err := s.roleUc.GetMenusByRole(ctx, req.RoleId)
	if err != nil {
		return nil, err
	}

	output := make([]*permissionV1.Menu, 0)
	err = copier.Copy(&output, &menus)
	if err != nil {
		return nil, err
	}

	return &permissionV1.ListMenusByRoleResponse{Items: output}, nil
}

// ListApiMetadata 查询所有接口元数据（来自 openapi.yaml 扫描）
func (s *PermissionService) ListApiMetadata(
	ctx context.Context,
	req *permissionV1.ListApiMetadataRequest,
) (*permissionV1.ListApiMetadataResponse, error) {
	groups := s.apiMetadata.Metadata()
	items := make([]*permissionV1.ApiMetadata, 0, len(groups))
	for _, g := range groups {
		endpoints := make([]*permissionV1.ApiEndpoint, 0, len(g.Endpoints))
		for _, e := range g.Endpoints {
			endpoints = append(endpoints, &permissionV1.ApiEndpoint{
				Method:  e.Method,
				Path:    e.Path,
				Summary: e.Summary,
			})
		}
		items = append(items, &permissionV1.ApiMetadata{
			Service:   g.Service,
			Tag:       g.Tag,
			Endpoints: endpoints,
		})
	}
	return &permissionV1.ListApiMetadataResponse{Items: items}, nil
}
