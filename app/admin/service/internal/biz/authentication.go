package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	authenticationV1 "github.com/hypercoze/kratos-admin/api/gen/go/authentication/service/v1"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/conf"
	"github.com/hypercoze/kratos-admin/pkg/utils/auth"
	"github.com/hypercoze/kratos-admin/pkg/utils/crypto"
	"time"
)

type TokenRepo interface {
	// 保存 token，支持主动失效控制
	Save(ctx context.Context, token string, userID string, ttl time.Duration) error

	// 获取 token 对应的用户ID（可选，用于验证 token）
	Get(ctx context.Context, token string) (string, error)

	// 删除 token（登出用，可选）
	Delete(ctx context.Context, token string) error

	// 是否存在该token
	Exists(ctx context.Context, token string) (bool, error)
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
	//storedCaptcha, err := uc.captchaRepo.Get(ctx, req.CaptchaId)
	//if err != nil {
	//	// TODO:记录日志
	//	return nil, authenticationV1.ErrorCaptchaExpired("CAPTCHA_EXPIRED")
	//}
	//// 一次性使用，删除验证码
	//_ = uc.captchaRepo.Delete(ctx, req.CaptchaId)
	//if storedCaptcha != strings.ToLower(req.Captcha) {
	//	// TODO:记录日志
	//	return nil, authenticationV1.ErrorCaptchaInvalid("CAPTCHA_INVALID")
	//}

	// 验证用户名和密码
	admin, err := uc.adminRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		// TODO:记录日志
		return nil, authenticationV1.ErrorBadRequest("BAD_REQUEST")
	}
	if admin == nil {
		return nil, authenticationV1.ErrorInvalidCredentials("INVALID_CREDENTIALS")
	}

	// 校验密码
	if !crypto.CheckPasswordHash(req.Password, admin.Password) {
		return nil, authenticationV1.ErrorInvalidCredentials("INVALID_CREDENTIALS")
	}

	// 生成 Access/Refresh
	return uc.GenerateTokens(ctx, admin.ID)
}

func (uc *AuthenticationUsecase) RefreshToken(ctx context.Context, req *authenticationV1.RefreshTokenRequest) (*authenticationV1.LoginResponse, error) {
	// 解析 refresh token
	claims := &auth.BaseClaims{}
	_, err := auth.ParseToken(req.RefreshToken, claims,
		auth.WithSecret([]byte(uc.jwtConf.RefreshSecret)),
		auth.WithSigningMethod(jwt.SigningMethodHS256),
	)
	if err != nil || claims.Type != "refresh" {
		return nil, authenticationV1.ErrorInvalidToken("INVALID_TOKEN")
	}

	exists, err := uc.tokenRepo.Exists(ctx, req.RefreshToken)
	if !exists || err != nil {
		return nil, authenticationV1.ErrorInvalidToken("INVALID_TOKEN")
	}
	_ = uc.tokenRepo.Delete(ctx, req.RefreshToken)
	_ = uc.tokenRepo.Delete(ctx, req.AccessToken)

	// 生成 Access/Refresh
	return uc.GenerateTokens(ctx, claims.UserID)
}

// 生成Access+Refresh
func (uc *AuthenticationUsecase) GenerateTokens(ctx context.Context, userID string) (*authenticationV1.LoginResponse, error) {
	// 生成 Access/Refresh
	accessTTL := uc.jwtConf.AccessTtl.AsDuration()
	refreshTTL := uc.jwtConf.RefreshTtl.AsDuration()
	accessClaims := auth.NewAccessClaims(userID, accessTTL, uc.jwtConf.Issuer)
	refreshClaims := auth.NewRefreshClaims(userID, refreshTTL, uc.jwtConf.Issuer)
	accessToken, err := auth.GenerateToken(accessClaims, auth.WithSecret([]byte(uc.jwtConf.AccessSecret)), auth.WithSigningMethod(jwt.SigningMethodHS256))
	if err != nil {
		return nil, authenticationV1.ErrorBadRequest("BAD_REQUEST")
	}

	refreshToken, err := auth.GenerateToken(refreshClaims, auth.WithSecret([]byte(uc.jwtConf.RefreshSecret)), auth.WithSigningMethod(jwt.SigningMethodHS256))
	if err != nil {
		return nil, authenticationV1.ErrorBadRequest("BAD_REQUEST")
	}

	// 保存 token 用于主动失效
	if err := uc.tokenRepo.Save(ctx, accessToken, userID, accessTTL); err != nil {
		uc.log.Warnf("save access token failed: %v", err)
		return nil, authenticationV1.ErrorBadRequest("BAD_REQUEST")
	}
	if err := uc.tokenRepo.Save(ctx, refreshToken, userID, refreshTTL); err != nil {
		uc.log.Warnf("save refresh token failed: %v", err)
		return nil, authenticationV1.ErrorBadRequest("BAD_REQUEST")
	}

	return &authenticationV1.LoginResponse{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresIn:  int64(accessTTL.Seconds()),
		RefreshTokenExpiresIn: int64(refreshTTL.Seconds()),
		TokenType:             "Bearer",
		UserId:                userID,
	}, nil
}

func (uc *AuthenticationUsecase) Logout(ctx context.Context, req *authenticationV1.RefreshTokenRequest) error {
	_ = uc.tokenRepo.Delete(ctx, req.RefreshToken)
	_ = uc.tokenRepo.Delete(ctx, req.AccessToken)

	return nil
}
