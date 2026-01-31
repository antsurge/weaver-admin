package biz

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	authenticationV1 "github.com/hypercoze/kratos-admin/api/gen/go/authentication/service/v1"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/conf"
	"github.com/hypercoze/kratos-admin/pkg/utils/auth"
	"github.com/hypercoze/kratos-admin/pkg/utils/crypto"
	"strings"
	"time"
)

type TokenRepo interface {
	// 保存 token，支持主动失效控制
	Save(ctx context.Context, token string, userID string, ttl time.Duration) error

	// 获取 token 对应的用户ID（可选，用于验证 token）
	GetUserID(ctx context.Context, token string) (string, error)

	// 删除 token（登出用，可选）
	Delete(ctx context.Context, token string) error
}

type AuthenticationUsecase struct {
	log         *log.Helper
	adminRepo   AdminRepo
	captchaRepo CaptchaRepo
	tokenRepo   TokenRepo
	jwtConf     *conf.JWT
}

func NewAuthenticationUsecase(
	logger log.Logger,
	adminRepo AdminRepo,
	captchaRepo CaptchaRepo,
	tokenRepo TokenRepo,
	jwtConf *conf.JWT,
) *AuthenticationUsecase {
	return &AuthenticationUsecase{
		log:         log.NewHelper(logger),
		adminRepo:   adminRepo,
		captchaRepo: captchaRepo,
		tokenRepo:   tokenRepo,
		jwtConf:     jwtConf,
	}
}

// 登录
func (uc *AuthenticationUsecase) Login(ctx context.Context, req *authenticationV1.LoginRequest) (*authenticationV1.LoginResponse, error) {
	// 获取验证码并校验
	storedCaptcha, err := uc.captchaRepo.Get(ctx, req.CaptchaId)
	if err != nil {
		return nil, fmt.Errorf("captcha expired or not found: %w", err)
	}
	// 一次性使用，删除验证码
	_ = uc.captchaRepo.Delete(ctx, req.CaptchaId)
	if storedCaptcha != strings.ToLower(req.Captcha) {
		return nil, fmt.Errorf("captcha invalid")
	}

	// 查询用户信息
	admin, err := uc.adminRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	if admin == nil {
		return nil, fmt.Errorf("user not found")
	}

	// 校验密码
	if !crypto.CheckPasswordHash(req.Password, admin.Password) {
		return nil, fmt.Errorf("password error")
	}
	// 生成 Token
	accessTTL := uc.jwtConf.AccessTtl.AsDuration()
	refreshTTL := uc.jwtConf.RefreshTtl.AsDuration()

	// 生成 Access Token Claims
	accessClaims := auth.NewBaseClaims(admin.ID, accessTTL, uc.jwtConf.Issuer)
	accessToken, err := auth.GenerateToken(
		accessClaims,
		auth.WithSecret([]byte(uc.jwtConf.Secret)),
		auth.WithSigningMethod(jwt.SigningMethodHS256))
	if err != nil {
		return nil, fmt.Errorf("generate access token failed: %w", err)
	}

	// 生成 Refresh Token Claims
	refreshClaims := auth.NewBaseClaims(admin.ID, refreshTTL, uc.jwtConf.Issuer)
	refreshToken, err := auth.GenerateToken(
		refreshClaims,
		auth.WithSecret([]byte(uc.jwtConf.Secret)),
		auth.WithSigningMethod(jwt.SigningMethodHS256))
	if err != nil {
		return nil, fmt.Errorf("generate refresh token failed: %w", err)
	}

	// 可选：保存 token 用于主动失效
	if err := uc.tokenRepo.Save(ctx, accessToken, admin.ID, accessTTL); err != nil {
		uc.log.Warnf("save access token failed: %v", err)
	}
	if err := uc.tokenRepo.Save(ctx, refreshToken, admin.ID, refreshTTL); err != nil {
		uc.log.Warnf("save refresh token failed: %v", err)
	}

	// 返回响应
	return &authenticationV1.LoginResponse{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresIn:  int64(accessTTL.Seconds()),
		RefreshTokenExpiresIn: int64(refreshTTL.Seconds()),
		TokenType:             "Bearer",
		UserId:                admin.ID,
	}, nil
}
