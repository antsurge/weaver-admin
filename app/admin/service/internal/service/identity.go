package service

import (
	"context"
	"strings"

	adminV1 "github.com/antsurge/weaver-admin/api/gen/go/admin/service/v1"
	identityV1 "github.com/antsurge/weaver-admin/api/gen/go/identity/service/v1"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/pkg/metadata"
	"github.com/antsurge/weaver-admin/pkg/utils/copierx"
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/emptypb"
)

type IdentityService struct {
	adminV1.UnimplementedIdentityServer

	adminUc *biz.AdminUseCase
	fileUc  *biz.FileUsecase
}

func NewIdentityService(adminUc *biz.AdminUseCase, fileUc *biz.FileUsecase) *IdentityService {
	return &IdentityService{adminUc: adminUc, fileUc: fileUc}
}

// fillAvatarURL 根据对象键补全头像的可访问地址，供前端直接使用。
// 转换失败（如存储未启用）时保持为空，不影响其它字段返回。
func (s *IdentityService) fillAvatarURL(ctx context.Context, output *identityV1.Admin, objectKey string) {
	if output == nil || objectKey == "" {
		return
	}
	// 已是完整 URL（http/https）时直接透传，兼容历史数据
	if strings.HasPrefix(objectKey, "http://") || strings.HasPrefix(objectKey, "https://") {
		output.AvatarUrl = objectKey
		return
	}
	accessURL, err := s.fileUc.GetAccessURL(ctx, objectKey)
	if err != nil {
		return
	}
	output.AvatarUrl = accessURL
}

func (s *IdentityService) ListAdmin(ctx context.Context, req *identityV1.ListAdminRequest) (*identityV1.ListAdminResponse, error) {
	input := biz.ListAdminRequest{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}
	// proto 分页字段名（page）与 biz 的 PaginationParams（currentPage）不一致，需显式映射
	input.CurrentPage = int(req.Page)
	input.PageSize = int(req.PageSize)

	res, err := s.adminUc.ListAdmin(ctx, &input)
	if err != nil {
		return nil, err
	}

	output := &identityV1.ListAdminResponse{}
	if res != nil {
		err = copierx.Copy(&output, res)
		if err != nil {
			return nil, err
		}
		// 为每个用户的头像对象键补全可访问地址
		for _, item := range output.Items {
			if item == nil {
				continue
			}
			s.fillAvatarURL(ctx, item, item.Avatar)
		}
	}

	return output, nil
}

// GetAdmin 获取用户详情（包含角色ID列表）
func (s *IdentityService) GetAdmin(ctx context.Context, req *identityV1.GetAdminRequest) (*identityV1.Admin, error) {
	admin, err := s.adminUc.GetAdminWithRoles(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	output := &identityV1.Admin{}
	err = copierx.Copy(output, admin)
	if err != nil {
		return nil, err
	}
	s.fillAvatarURL(ctx, output, output.Avatar)

	return output, nil
}

// CreateAdmin 创建用户（支持同时绑定角色）
func (s *IdentityService) CreateAdmin(ctx context.Context, req *identityV1.CreateAdminRequest) (*identityV1.Admin, error) {
	input := biz.Admin{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}
	admin, err := s.adminUc.CreateAdmin(ctx, &input)
	if err != nil {
		return nil, err
	}
	output := &identityV1.Admin{}
	err = copierx.Copy(output, admin)
	if err != nil {
		return nil, err
	}
	s.fillAvatarURL(ctx, output, output.Avatar)

	return output, nil
}

// UpdateAdmin 更新用户（支持同时重新绑定角色）
func (s *IdentityService) UpdateAdmin(ctx context.Context, req *identityV1.UpdateAdminRequest) (*identityV1.Admin, error) {
	input := biz.Admin{}
	var err error
	err = copier.Copy(&input, req)
	if err != nil {
		return nil, err
	}
	admin, err := s.adminUc.UpdateAdmin(ctx, &input)
	if err != nil {
		return nil, err
	}
	output := &identityV1.Admin{}
	err = copierx.Copy(output, admin)
	if err != nil {
		return nil, err
	}
	s.fillAvatarURL(ctx, output, output.Avatar)

	return output, err
}

// DeleteAdmin 删除用户（批量软删除；禁止删除自己与超管账号）
func (s *IdentityService) DeleteAdmin(ctx context.Context, req *identityV1.DeleteAdminRequest) (*emptypb.Empty, error) {
	err := s.adminUc.DeleteAdmin(ctx, req.Ids, metadata.GetAdminID(ctx))
	return nil, err
}

// ResetPassword 重置指定用户密码
func (s *IdentityService) ResetPassword(ctx context.Context, req *identityV1.ResetPasswordRequest) (*emptypb.Empty, error) {
	err := s.adminUc.ResetPassword(ctx, req.Id, req.Password)
	return nil, err
}

// UpdateAdminStatus 启用/禁用指定用户
func (s *IdentityService) UpdateAdminStatus(ctx context.Context, req *identityV1.UpdateAdminStatusRequest) (*emptypb.Empty, error) {
	err := s.adminUc.UpdateStatus(ctx, req.Id, req.Status)
	return nil, err
}

// BatchUpdateAdminStatus 批量启用/禁用用户
func (s *IdentityService) BatchUpdateAdminStatus(ctx context.Context, req *identityV1.BatchUpdateAdminStatusRequest) (*emptypb.Empty, error) {
	err := s.adminUc.BatchUpdateStatus(ctx, req.Ids, req.Status)
	return nil, err
}

// UsernameExists 校验用户名是否存在（编辑时通过 id 排除自身）
func (s *IdentityService) UsernameExists(ctx context.Context, req *identityV1.UsernameExistsRequest) (*identityV1.UsernameExistsResponse, error) {
	exists, err := s.adminUc.IsUsernameExists(ctx, req.Username, req.Id)
	if err != nil {
		return nil, err
	}
	return &identityV1.UsernameExistsResponse{Exists: exists}, nil
}
