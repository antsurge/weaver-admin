package service

import (
	"context"
	adminV1 "github.com/hypercoze/kratos-admin/api/gen/go/admin/service/v1"
	authenticationV1 "github.com/hypercoze/kratos-admin/api/gen/go/authentication/service/v1"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/biz"
)

type AuthenticationService struct {
	adminV1.UnimplementedAuthenticationServiceServer
	//adminV1.AuthenticationServiceHTTPServer
	//adminV1.AuthenticationServiceServer
	authenticationUc *biz.AuthenticationUsecase
}

func NewAuthenticationService(
	authenticationUc *biz.AuthenticationUsecase,
) *AuthenticationService {
	return &AuthenticationService{
		authenticationUc: authenticationUc,
	}
}

func (s *AuthenticationService) Login(ctx context.Context, req *authenticationV1.LoginRequest) (*authenticationV1.LoginResponse, error) {
	return &authenticationV1.LoginResponse{}, nil
}
