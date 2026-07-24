package biz

import (
	"context"
	"strings"
	"time"

	authenticationV1 "github.com/antsurge/weaver-admin/api/gen/go/authentication/service/v1"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/conf"
	"github.com/antsurge/weaver-admin/pkg/metadata"
	"github.com/antsurge/weaver-admin/pkg/utils/auth"
	"github.com/antsurge/weaver-admin/pkg/utils/crypto"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
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

type CurrentUserInfoResponse struct {
	Id        string
	RealName  string
	MenusTree []*Menu
}

type AuthenticationUsecase struct {
	log           *log.Helper
	adminRepo     AdminRepo
	captchaRepo   CaptchaRepo
	tokenRepo     TokenRepo
	roleRepo      RoleRepo
	menuRepo      MenuRepo
	adminRoleRepo AdminRoleRepo
	roleMenuRepo  RoleMenuRepo
	jwtConf       *conf.JWT
}

func NewAuthenticationUsecase(
	logger log.Logger,
	adminRepo AdminRepo,
	captchaRepo CaptchaRepo,
	tokenRepo TokenRepo,
	roleRepo RoleRepo,
	menuRepo MenuRepo,
	adminRoleRepo AdminRoleRepo,
	roleMenuRepo RoleMenuRepo,
	jwtConf *conf.JWT,
) *AuthenticationUsecase {
	return &AuthenticationUsecase{
		log:           log.NewHelper(logger),
		adminRepo:     adminRepo,
		captchaRepo:   captchaRepo,
		tokenRepo:     tokenRepo,
		roleRepo:      roleRepo,
		menuRepo:      menuRepo,
		adminRoleRepo: adminRoleRepo,
		roleMenuRepo:  roleMenuRepo,
		jwtConf:       jwtConf,
	}
}

// 登录
func (uc *AuthenticationUsecase) Login(ctx context.Context, req *authenticationV1.LoginRequest) (*authenticationV1.LoginResponse, error) {
	// 获取验证码并校验
	storedCaptcha, err := uc.captchaRepo.Get(ctx, req.CaptchaId)
	if err != nil {
		// TODO:记录日志
		return nil, authenticationV1.ErrorCaptchaExpired("CAPTCHA_EXPIRED")
	}
	// 一次性使用，删除验证码
	_ = uc.captchaRepo.Delete(ctx, req.CaptchaId)
	if storedCaptcha != strings.ToLower(req.Captcha) {
		// TODO:记录日志
		return nil, authenticationV1.ErrorCaptchaInvalid("CAPTCHA_INVALID")
	}

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

	// 获取当前用户的角色
	roleIds, err := uc.adminRoleRepo.GetRoleIdsByAdminId(ctx, admin.ID)
	if err != nil {
		return nil, err
	}
	if len(roleIds) == 0 {
		return nil, authenticationV1.ErrorNoPermission("NO_PERMISSION")
	}

	// 根据角色id获取已启用的菜单（已启用的）
	munuIds, err := uc.roleMenuRepo.GetMenuIdsByRoleIds(ctx, roleIds)
	if err != nil {
		return nil, err
	}
	if len(munuIds) == 0 {
		return nil, authenticationV1.ErrorNoPermission("NO_PERMISSION")
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

func (uc *AuthenticationUsecase) CurrentUserInfo(ctx context.Context) (*CurrentUserInfoResponse, error) {
	// 验证用户名和密码
	admin, err := uc.adminRepo.FindByID(ctx, metadata.GetAdminID(ctx))
	if err != nil {
		// TODO:记录日志
		return nil, authenticationV1.ErrorInvalidToken("INVALID_TOKEN")
	}
	if admin == nil {
		return nil, authenticationV1.ErrorInvalidToken("INVALID_TOKEN")
	}

	// 2. 查询用户关联的角色ID列表
	roleIDs, err := uc.adminRepo.GetRoleIDsByAdmin(ctx, admin.ID)
	if err != nil {
		uc.log.Errorf("获取用户 %s 的角色失败: %v", admin.ID, err)
		return nil, err
	}

	// 如果用户没有绑定角色，返回空菜单
	if len(roleIDs) == 0 {
		uc.log.Warnf("用户 %s 没有绑定任何角色", admin.ID)
		return nil, authenticationV1.ErrorNoPermission("NO_PERMISSION")
	}

	// 3. 查询所有角色关联的菜单ID列表（去重）
	menuIDSet := make(map[string]bool)
	for _, roleID := range roleIDs {
		menuIDs, err := uc.roleRepo.GetMenuIDsByRole(ctx, roleID)
		if err != nil {
			uc.log.Warnf("获取角色 %s 的菜单失败: %v", roleID, err)
			continue
		}
		for _, menuID := range menuIDs {
			menuIDSet[menuID] = true
		}
	}

	// 如果没有关联任何菜单，返回空
	if len(menuIDSet) == 0 {
		return nil, authenticationV1.ErrorNoPermission("NO_PERMISSION")
	}

	// 4. 根据菜单ID列表查询完整的菜单数据
	menuIDs := make([]string, 0, len(menuIDSet))
	for id := range menuIDSet {
		menuIDs = append(menuIDs, id)
	}

	menus, err := uc.menuRepo.GetMenusByIDs(ctx, menuIDs)
	if err != nil {
		uc.log.Errorf("查询菜单详情失败: %v", err)
		return nil, err
	}

	// 5. 构建树形结构
	tree := buildMenuTree(menus)

	return &CurrentUserInfoResponse{
		Id:        admin.ID,
		RealName:  admin.RealName,
		MenusTree: tree,
	}, nil
}

// CurrentUserMenus 获取当前用户的菜单（根据用户角色返回绑定的菜单树）
func (uc *AuthenticationUsecase) CurrentUserMenus(ctx context.Context) ([]*Menu, error) {
	// 1. 获取当前用户ID
	userID := metadata.GetAdminID(ctx)

	// 2. 查询用户关联的角色ID列表
	roleIDs, err := uc.adminRepo.GetRoleIDsByAdmin(ctx, userID)
	if err != nil {
		uc.log.Errorf("获取用户 %s 的角色失败: %v", userID, err)
		return nil, err
	}

	// 如果用户没有绑定角色，返回空菜单
	if len(roleIDs) == 0 {
		uc.log.Warnf("用户 %s 没有绑定任何角色", userID)
		return []*Menu{}, nil
	}

	// 3. 查询所有角色关联的菜单ID列表（去重）
	menuIDSet := make(map[string]bool)
	for _, roleID := range roleIDs {
		menuIDs, err := uc.roleRepo.GetMenuIDsByRole(ctx, roleID)
		if err != nil {
			uc.log.Warnf("获取角色 %s 的菜单失败: %v", roleID, err)
			continue
		}
		for _, menuID := range menuIDs {
			menuIDSet[menuID] = true
		}
	}

	// 如果没有关联任何菜单，返回空
	if len(menuIDSet) == 0 {
		return []*Menu{}, nil
	}

	// 4. 根据菜单ID列表查询完整的菜单数据
	menuIDs := make([]string, 0, len(menuIDSet))
	for id := range menuIDSet {
		menuIDs = append(menuIDs, id)
	}

	menus, err := uc.menuRepo.GetMenusByIDs(ctx, menuIDs)
	if err != nil {
		uc.log.Errorf("查询菜单详情失败: %v", err)
		return nil, err
	}

	// 5. 构建树形结构
	tree := buildMenuTree(menus)
	return tree, nil
}
