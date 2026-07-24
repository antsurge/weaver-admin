package service

import (
	"context"

	adminV1 "github.com/antsurge/weaver-admin/api/gen/go/admin/service/v1"
	authenticationV1 "github.com/antsurge/weaver-admin/api/gen/go/authentication/service/v1"
	permissionV1 "github.com/antsurge/weaver-admin/api/gen/go/permission/service/v1"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/pkg/utils/copierx"
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthenticationService struct {
	adminV1.UnimplementedAuthenticationServiceServer

	authenticationUc *biz.AuthenticationUsecase
	captchaUc        *biz.CaptchaUsecase
}

func NewAuthenticationService(
	authenticationUc *biz.AuthenticationUsecase,
	captchaUc *biz.CaptchaUsecase,
) *AuthenticationService {
	return &AuthenticationService{
		authenticationUc: authenticationUc,
		captchaUc:        captchaUc,
	}
}

// 获取验证码
func (s *AuthenticationService) GetCaptcha(ctx context.Context, _ *emptypb.Empty) (*authenticationV1.GetCaptchaResponse, error) {
	return s.captchaUc.GetCaptcha(ctx)
}

// 用户登录
func (s *AuthenticationService) Login(ctx context.Context, req *authenticationV1.LoginRequest) (*authenticationV1.LoginResponse, error) {
	return s.authenticationUc.Login(ctx, req)
}

// 刷新认证数据
func (s *AuthenticationService) RefreshToken(ctx context.Context, req *authenticationV1.RefreshTokenRequest) (*authenticationV1.LoginResponse, error) {
	return s.authenticationUc.RefreshToken(ctx, req)
}

// 退出
func (s *AuthenticationService) Logout(ctx context.Context, req *authenticationV1.RefreshTokenRequest) (*emptypb.Empty, error) {
	err := s.authenticationUc.Logout(ctx, req)
	return nil, err
}

// 获取验证码
func (s *AuthenticationService) CurrentUserInfo(ctx context.Context, _ *emptypb.Empty) (*authenticationV1.CurrentUserInfoResponse, error) {
	data, err := s.authenticationUc.CurrentUserInfo(ctx)
	if err != nil {
		return nil, err
	}
	output := &authenticationV1.CurrentUserInfoResponse{}
	err = copierx.Copy(output, data)

	return output, nil
}

// CurrentUserMenus 获取当前用户的菜单（根据用户角色返回绑定的菜单树）
func (s *AuthenticationService) CurrentUserMenus(ctx context.Context, _ *emptypb.Empty) (*authenticationV1.CurrentUserMenusResponse, error) {
	menus, err := s.authenticationUc.CurrentUserMenus(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为 Proto 消息
	output := make([]*permissionV1.Menu, 0)
	err = copier.Copy(&output, &menus)
	if err != nil {
		return nil, err
	}

	return &authenticationV1.CurrentUserMenusResponse{Items: output}, nil
}
