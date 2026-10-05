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

	menuUc           *biz.MenuUsecase
	roleUc           *biz.RoleUsecase
	apiMetadata      *openapi_scanner.Service
	dataPermissionUc *biz.DataPermissionUsecase
}

func NewPermissionService(
	menuUc *biz.MenuUsecase,
	roleUc *biz.RoleUsecase,
	apiMetadata *openapi_scanner.Service,
	dataPermissionUc *biz.DataPermissionUsecase,
) *PermissionService {
	return &PermissionService{
		menuUc:           menuUc,
		roleUc:           roleUc,
		apiMetadata:      apiMetadata,
		dataPermissionUc: dataPermissionUc,
	}
}

// 列表权限tree
func (s *PermissionService) MenuTree(ctx context.Context, req *permissionV1.MenuTreeRequest) (*permissionV1.MenuTreeResponse, error) {
	input := &biz.ListMenuRequest{}
	err := copierx.Copy(&input, &req)
	if err != nil {
		return nil, err
	}

	tree, err := s.menuUc.MenuTree(ctx, input)
	if err != nil {
		return nil, err
	}
	output := make([]*permissionV1.Menu, 0)
	err = copierx.Copy(&output, &tree)
	if err != nil {
		return nil, err
	}

	return &permissionV1.MenuTreeResponse{Items: output}, nil
}

// 获取权限详情（含接口权限）
func (s *PermissionService) GetMenu(ctx context.Context, req *permissionV1.GetMenuRequest) (*permissionV1.Menu, error) {
	permission, err := s.menuUc.GetMenu(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	output := &permissionV1.Menu{}
	err = copierx.Copy(&output, &permission)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// 创建权限
func (s *PermissionService) CreateMenu(ctx context.Context, req *permissionV1.CreateMenuRequest) (*permissionV1.Menu, error) {
	input := &biz.Menu{}
	err := copierx.Copy(&input, &req)
	if err != nil {
		return nil, err
	}
	permission, err := s.menuUc.CreateMenu(ctx, input)
	if err != nil {
		return nil, err
	}

	output := &permissionV1.Menu{}
	err = copierx.Copy(&output, &permission)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// 更新权限
func (s *PermissionService) UpdateMenu(ctx context.Context, req *permissionV1.UpdateMenuRequest) (*permissionV1.Menu, error) {
	input := &biz.Menu{}
	err := copierx.Copy(&input, &req)
	if err != nil {
		return nil, err
	}
	permission, err := s.menuUc.UpdateMenu(ctx, input)
	if err != nil {
		return nil, err
	}

	output := &permissionV1.Menu{}
	err = copierx.Copy(&output, &permission)
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

// 角色编码是否存在（用于表单失焦校验）
func (s *PermissionService) IsRoleCodeExists(ctx context.Context, req *permissionV1.IsRoleCodeExistsRequest) (*permissionV1.IsRoleFieldExistsResponse, error) {
	exists, err := s.roleUc.IsRoleCodeExists(ctx, req.Code, req.Id)
	if err != nil {
		return nil, err
	}
	return &permissionV1.IsRoleFieldExistsResponse{Exists: exists}, nil
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
	err = copierx.Copy(&output, &menus)
	if err != nil {
		return nil, err
	}

	return &permissionV1.ListMenusByRoleResponse{Items: output}, nil
}

// ====== RBAC 角色数据权限规则绑定方法 ======

// BindDataPermissionsForRole 为角色绑定数据权限规则（全量替换）
func (s *PermissionService) BindDataPermissionsForRole(
	ctx context.Context,
	req *permissionV1.BindDataPermissionsForRoleRequest,
) (*emptypb.Empty, error) {
	err := s.roleUc.BindDataPermissionsForRole(ctx, req.RoleId, req.DataPermissionIds)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ListDataPermissionsByRole 查询角色已绑定的数据权限规则列表
func (s *PermissionService) ListDataPermissionsByRole(
	ctx context.Context,
	req *permissionV1.ListDataPermissionsByRoleRequest,
) (*permissionV1.ListDataPermissionsByRoleResponse, error) {
	perms, err := s.roleUc.GetDataPermissionsByRole(ctx, req.RoleId)
	if err != nil {
		return nil, err
	}

	output := make([]*permissionV1.DataPermission, 0, len(perms))
	err = copierx.Copy(&output, &perms)
	if err != nil {
		return nil, err
	}

	return &permissionV1.ListDataPermissionsByRoleResponse{Items: output}, nil
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

// ====== 数据权限 ======

// ListDataPermission 数据权限规则列表（分页）
func (s *PermissionService) ListDataPermission(ctx context.Context, req *permissionV1.ListDataPermissionRequest) (*permissionV1.ListDataPermissionResponse, error) {
	input := biz.ListDataPermissionRequest{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}

	res, err := s.dataPermissionUc.List(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &permissionV1.ListDataPermissionResponse{}
	if res != nil {
		output.Total = int64(res.Total)
		err = copierx.Copy(&output.Items, &res.Data)
	}

	return output, err
}

// CreateDataPermission 创建数据权限规则
func (s *PermissionService) CreateDataPermission(ctx context.Context, req *permissionV1.CreateDataPermissionRequest) (*permissionV1.DataPermission, error) {
	input := biz.DataPermission{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}

	data, err := s.dataPermissionUc.Create(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &permissionV1.DataPermission{}
	err = copierx.Copy(output, data)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// UpdateDataPermission 更新数据权限规则
func (s *PermissionService) UpdateDataPermission(ctx context.Context, req *permissionV1.UpdateDataPermissionRequest) (*permissionV1.DataPermission, error) {
	input := biz.DataPermission{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}

	data, err := s.dataPermissionUc.Update(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &permissionV1.DataPermission{}
	err = copierx.Copy(output, data)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// UpdateDataPermissionStatus 更新数据权限规则状态
func (s *PermissionService) UpdateDataPermissionStatus(ctx context.Context, req *permissionV1.UpdateDataPermissionStatusRequest) (*emptypb.Empty, error) {
	err := s.dataPermissionUc.UpdateStatus(ctx, req.Id, req.Status)
	return nil, err
}

// DeleteDataPermission 删除数据权限规则（批量）
func (s *PermissionService) DeleteDataPermission(ctx context.Context, req *permissionV1.DeleteDataPermissionRequest) (*emptypb.Empty, error) {
	err := s.dataPermissionUc.Delete(ctx, req.Ids)
	return nil, err
}

// IsDataPermissionCodeExists 规则编码是否存在（用于表单失焦校验）
func (s *PermissionService) IsDataPermissionCodeExists(ctx context.Context, req *permissionV1.IsDataPermissionCodeExistsRequest) (*permissionV1.IsDataPermissionFieldExistsResponse, error) {
	exists, err := s.dataPermissionUc.IsCodeExists(ctx, req.Code, req.Id)
	if err != nil {
		return nil, err
	}
	return &permissionV1.IsDataPermissionFieldExistsResponse{Exists: exists}, nil
}
