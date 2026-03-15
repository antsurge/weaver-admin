package service

import (
	"context"

	adminV1 "github.com/hypercoze/kratos-admin/api/gen/go/admin/service/v1"
	authenticationV1 "github.com/hypercoze/kratos-admin/api/gen/go/authentication/service/v1"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/biz"
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
	return s.authenticationUc.CurrentUserInfo(ctx)
}
