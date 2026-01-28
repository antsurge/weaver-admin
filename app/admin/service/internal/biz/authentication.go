package biz

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	v1 "github.com/hypercoze/kratos-admin/api/gen/go/authentication/service/v1"
	"strings"
)

type AuthenticationUsecase struct {
	log       *log.Helper
	adminRepo AdminRepo
}

func NewAuthenticationUsecase(
	logger log.Logger,
	adminRepo AdminRepo,
) *AuthenticationUsecase {
	return &AuthenticationUsecase{
		log:       log.NewHelper(logger),
		adminRepo: adminRepo,
	}
}

// 登录
func (uc *AuthenticationUsecase) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	// 参数处理
	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	fmt.Println(username, password)

	return &v1.LoginResponse{}, nil
}
