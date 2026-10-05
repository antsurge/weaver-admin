package biz

import (
	"context"
	"sort"
	"strings"
	"time"

	authenticationV1 "github.com/antsurge/weaver-admin/api/gen/go/authentication/service/v1"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/conf"
	"github.com/antsurge/weaver-admin/pkg/metadata"
	"github.com/antsurge/weaver-admin/pkg/tenant"
	"github.com/antsurge/weaver-admin/pkg/utils/auth"
	"github.com/antsurge/weaver-admin/pkg/utils/crypto"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/exp/slices"
)

type TokenRepo interface {
	// 保存 token 会话（value 为 JSON：userID/type/ip/ua/loginAt），支持主动失效控制
	Save(ctx context.Context, token string, s *TokenSession, ttl time.Duration) error

	// 获取 token 对应的用户ID（用于验证 token）
	Get(ctx context.Context, token string) (string, error)

	// 删除 token（登出/强制下线单会话用）
	Delete(ctx context.Context, token string) error

	// 是否存在该token
	Exists(ctx context.Context, token string) (bool, error)

	// 列出全部在线会话（仅 access token，用于在线用户列表）
	ListSessions(ctx context.Context) ([]*TokenSession, error)

	// 删除指定用户的全部 token（强制下线用户用）
	DeleteByUser(ctx context.Context, userID string) error
}

// TokenSession 一次登录会话的元信息（持久化在 Redis token key 的 value 中）
type TokenSession struct {
	Token        string        // JWT 字符串（data 层回填）
	UserID       string        // 用户ID
	Type         string        // access / refresh
	IP           string        // 登录 IP
	UserAgent    string        // 浏览器/设备 UA
	LoginAt      int64         // 登录时间（Unix 秒）
	RemainingTTL time.Duration // 剩余有效时长（data 层回填，用于计算过期时间）
}

type CurrentUserInfoResponse struct {
	Id        string
	RealName  string
	MenuTree  []*Menu
	RoleCodes []string
	// 当前生效的数据权限范围（用户覆盖 > 角色并集；ALL/DEPT_AND_CHILD/DEPT/SELF）
	DataScope string
	// 归属部门名称（个人中心展示，无部门为空串）
	DepartmentName string
	// 角色名称列表（个人中心展示）
	RoleNames []string
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
	logUc         *LogUsecase
	jwtConf       *conf.JWT
	appConf       *conf.App
	securityUc    *SecurityUsecase
	tenantRepo    TenantRepo
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
	logUc *LogUsecase,
	jwtConf *conf.JWT,
	appConf *conf.App,
	securityUc *SecurityUsecase,
	tenantRepo TenantRepo,
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
		logUc:         logUc,
		jwtConf:       jwtConf,
		appConf:       appConf,
		securityUc:    securityUc,
		tenantRepo:    tenantRepo,
	}
}

// 登录
func (uc *AuthenticationUsecase) Login(ctx context.Context, req *authenticationV1.LoginRequest) (*authenticationV1.LoginResponse, error) {
	// 登录为匿名接口，需显式注入租户作用域：
	// 单租户模式固定为 default 租户；多租户模式由登录请求携带的租户编码解析并校验（见 resolveLoginTenant）。
	ctx, err := uc.resolveLoginTenant(ctx, req)
	if err != nil {
		return nil, err
	}

	// 登录日志：无论成败都记录，便于安全审计
	recordLogin := func(adminID, status, reason string) {
		ip, ua := clientInfoFromCtx(ctx)
		uc.logUc.RecordLogin(ctx, &LoginLog{
			Username:  req.GetUsername(),
			AdminID:   adminID,
			Status:    status,
			Reason:    reason,
			IP:        ip,
			UserAgent: ua,
		})
	}

	// 获取验证码并校验
	storedCaptcha, err := uc.captchaRepo.Get(ctx, req.CaptchaId)
	if err != nil {
		recordLogin("", LogStatusFail, "CAPTCHA_EXPIRED")
		return nil, authenticationV1.ErrorCaptchaExpired("CAPTCHA_EXPIRED")
	}
	// 一次性使用，删除验证码
	_ = uc.captchaRepo.Delete(ctx, req.CaptchaId)
	if storedCaptcha != strings.ToLower(req.Captcha) {
		recordLogin("", LogStatusFail, "CAPTCHA_INVALID")
		return nil, authenticationV1.ErrorCaptchaInvalid("CAPTCHA_INVALID")
	}

	// 登录失败次数限制：处于锁定期直接拒绝（先于凭据校验，避免继续暴力尝试）
	if locked, _, err := uc.securityUc.LoginLocked(ctx, req.Username); err != nil {
		uc.log.Errorf("check login lock failed: %v", err)
	} else if locked {
		recordLogin("", LogStatusFail, "ACCOUNT_LOCKED")
		return nil, authenticationV1.ErrorAccountLocked("ACCOUNT_LOCKED")
	}

	// 验证用户名和密码
	admin, err := uc.adminRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		recordLogin("", LogStatusFail, "BAD_REQUEST")
		return nil, authenticationV1.ErrorBadRequest("BAD_REQUEST")
	}
	if admin == nil {
		// 用户名不存在同样计入失败次数，避免账号枚举
		uc.securityUc.RecordLoginFailure(ctx, req.Username)
		recordLogin("", LogStatusFail, "INVALID_CREDENTIALS")
		return nil, authenticationV1.ErrorInvalidCredentials("INVALID_CREDENTIALS")
	}
	// 账号被禁用则拒绝登录
	if admin.Status == "disabled" {
		recordLogin(admin.ID, LogStatusFail, "ACCOUNT_DISABLED")
		return nil, authenticationV1.ErrorNoPermission("ACCOUNT_DISABLED")
	}
	// 校验密码
	if !crypto.CheckPasswordHash(req.Password, admin.Password) {
		uc.securityUc.RecordLoginFailure(ctx, req.Username)
		recordLogin(admin.ID, LogStatusFail, "INVALID_CREDENTIALS")
		return nil, authenticationV1.ErrorInvalidCredentials("INVALID_CREDENTIALS")
	}

	// 获取当前用户的角色
	roleIds, err := uc.adminRoleRepo.GetRoleIdsByAdminId(ctx, admin.ID)
	if err != nil {
		return nil, err
	}
	if len(roleIds) == 0 {
		recordLogin(admin.ID, LogStatusFail, "NO_PERMISSION")
		return nil, authenticationV1.ErrorNoPermission("NO_PERMISSION")
	}

	// 根据roleIds获取codes
	roleCodes, err := uc.roleRepo.GetCodesByIds(ctx, roleIds)
	if err != nil {
		return nil, err
	}
	// 若不是超级管理员则需要进一步测试
	if len(roleCodes) == 0 {
		recordLogin(admin.ID, LogStatusFail, "NO_PERMISSION")
		return nil, authenticationV1.ErrorNoPermission("NO_PERMISSION")
	}

	if exists := slices.Contains(roleCodes, uc.appConf.SuperAdminCode); !exists {
		// 根据角色id获取已启用的菜单（已启用的）
		munuIds, err := uc.roleMenuRepo.GetMenuIdsByRoleIds(ctx, roleIds)
		if err != nil {
			return nil, err
		}
		if len(munuIds) == 0 {
			recordLogin(admin.ID, LogStatusFail, "NO_PERMISSION")
			return nil, authenticationV1.ErrorNoPermission("NO_PERMISSION")
		}
	}

	// 登录成功：清零失败计数
	uc.securityUc.ResetLoginFailure(ctx, req.Username)

	// 登录成功
	recordLogin(admin.ID, LogStatusSuccess, "")

	// 生成 Access/Refresh
	return uc.GenerateTokens(ctx, admin.ID, admin.TenantID)
}

// resolveLoginTenant 解析登录请求的租户作用域。
//   - single 模式：固定注入 default 租户；
//   - multi 模式：由登录请求携带的租户编码解析租户，校验存在/启用/未过期后注入其租户ID；
//     任一校验失败均拒绝登录。
func (uc *AuthenticationUsecase) resolveLoginTenant(ctx context.Context, req *authenticationV1.LoginRequest) (context.Context, error) {
	if uc.appConf.GetTenantMode() != "multi" {
		return tenant.With(ctx, tenant.DefaultTenantID), nil
	}

	code := req.GetTenantCode()
	if code == "" {
		return nil, authenticationV1.ErrorBadRequest("TENANT_CODE_REQUIRED")
	}
	t, err := uc.tenantRepo.GetByCode(ctx, code)
	if err != nil {
		uc.log.Errorf("resolve tenant by code %q failed: %v", code, err)
		return nil, authenticationV1.ErrorBadRequest("TENANT_QUERY_FAILED")
	}
	if t == nil {
		return nil, authenticationV1.ErrorBadRequest("TENANT_NOT_FOUND")
	}
	if t.Status != "enabled" {
		return nil, authenticationV1.ErrorBadRequest("TENANT_DISABLED")
	}
	if t.ExpireAt != nil && time.Now().After(*t.ExpireAt) {
		return nil, authenticationV1.ErrorBadRequest("TENANT_EXPIRED")
	}
	return tenant.With(ctx, t.ID), nil
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
	return uc.GenerateTokens(ctx, claims.UserID, claims.TenantID)
}

// 生成Access+Refresh
// tenantID 登录时取自该管理员所属租户；刷新 token 时取自旧 claims。
func (uc *AuthenticationUsecase) GenerateTokens(ctx context.Context, userID, tenantID string) (*authenticationV1.LoginResponse, error) {
	// Token 有效期由安全策略（数据库）统一管理；读取失败或未配置时回退内置默认值
	accessTTL, refreshTTL := uc.securityUc.TokenTTLs(ctx)
	accessClaims := auth.NewAccessClaims(userID, tenantID, accessTTL, uc.jwtConf.Issuer)
	refreshClaims := auth.NewRefreshClaims(userID, tenantID, refreshTTL, uc.jwtConf.Issuer)
	accessToken, err := auth.GenerateToken(accessClaims, auth.WithSecret([]byte(uc.jwtConf.AccessSecret)), auth.WithSigningMethod(jwt.SigningMethodHS256))
	if err != nil {
		return nil, authenticationV1.ErrorBadRequest("BAD_REQUEST")
	}

	refreshToken, err := auth.GenerateToken(refreshClaims, auth.WithSecret([]byte(uc.jwtConf.RefreshSecret)), auth.WithSigningMethod(jwt.SigningMethodHS256))
	if err != nil {
		return nil, authenticationV1.ErrorBadRequest("BAD_REQUEST")
	}

	// 保存 token 用于主动失效（同时记录会话元信息，供在线用户列表使用）
	ip, ua := clientInfoFromCtx(ctx)
	loginAt := time.Now().Unix()
	base := &TokenSession{UserID: userID, IP: ip, UserAgent: ua, LoginAt: loginAt}

	accessSession := *base
	accessSession.Token = accessToken
	accessSession.Type = "access"
	if err := uc.tokenRepo.Save(ctx, accessToken, &accessSession, accessTTL); err != nil {
		uc.log.Warnf("save access token failed: %v", err)
		return nil, authenticationV1.ErrorBadRequest("BAD_REQUEST")
	}
	refreshSession := *base
	refreshSession.Token = refreshToken
	refreshSession.Type = "refresh"
	if err := uc.tokenRepo.Save(ctx, refreshToken, &refreshSession, refreshTTL); err != nil {
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

	menus, roleCodes, err := uc.currentUserMenus(ctx)
	if err != nil {
		return nil, err
	}

	// 解析当前生效的数据权限范围（用户绑定规则 > 角色并集；超管恒为 ALL）
	dataScope, err := uc.resolveDataScope(ctx, admin.ID)
	if err != nil {
		uc.log.Warnf("解析用户 %s 的数据权限范围失败: %v", admin.ID, err)
		dataScope = DataScopeSelf
	}

	// 个人中心展示：角色名称列表 + 归属部门名称（查询失败不阻塞主流程）
	roleNames := []string{}
	if roleIDs, err := uc.adminRepo.GetRoleIDsByAdmin(ctx, admin.ID); err == nil {
		if names, err := uc.roleRepo.GetNamesByIds(ctx, roleIDs); err == nil {
			roleNames = names
		}
	}
	departmentName := ""
	if admin.DepartmentID != "" {
		if deptNames, err := uc.adminRepo.GetDepartmentNamesByIDs(ctx, []string{admin.DepartmentID}); err == nil {
			departmentName = deptNames[admin.DepartmentID]
		}
	}

	return &CurrentUserInfoResponse{
		Id:             admin.ID,
		RealName:       admin.RealName,
		MenuTree:       buildMenuTree(menus),
		RoleCodes:      roleCodes,
		DataScope:      dataScope,
		DepartmentName: departmentName,
		RoleNames:      roleNames,
	}, nil
}

// resolveDataScope 计算用户生效的数据权限范围。
// 优先级：用户直接绑定的数据权限规则（多规则取权限最大者） >
//
//	角色绑定的数据权限规则（多角色取权限最大者，即并集效果）> 默认仅本人；
//
// 绑定超级管理员角色的用户恒为全部数据。
func (uc *AuthenticationUsecase) resolveDataScope(ctx context.Context, adminID string) (string, error) {
	// 1. 用户直接绑定的数据权限规则优先（比角色规则优先级更高）
	adminPerms, err := uc.adminRepo.GetDataPermissionsByAdminIDs(ctx, []string{adminID})
	if err != nil {
		return "", err
	}
	if perms := adminPerms[adminID]; len(perms) > 0 {
		effective := DataScopeSelf
		effectiveRank := DataScopeRank[effective]
		for _, p := range perms {
			scope := DataScopeFromRule(p.ScopeType)
			rank, ok := DataScopeRank[scope]
			if !ok {
				continue
			}
			if rank > effectiveRank {
				effective = scope
				effectiveRank = rank
			}
		}
		return effective, nil
	}

	// 2. 查询用户角色
	roleIDs, err := uc.adminRoleRepo.GetRoleIdsByAdminId(ctx, adminID)
	if err != nil {
		return "", err
	}
	if len(roleIDs) == 0 {
		return DataScopeSelf, nil
	}

	// 3. 绑定超管角色：恒为全部数据
	roleCodes, err := uc.roleRepo.GetCodesByIds(ctx, roleIDs)
	if err != nil {
		return "", err
	}
	if uc.appConf != nil && uc.appConf.SuperAdminCode != "" &&
		slices.Contains(roleCodes, uc.appConf.SuperAdminCode) {
		return DataScopeAll, nil
	}

	// 4. 按角色绑定的数据权限规则取权限最大者（并集效果）
	rolePerms, err := uc.roleRepo.GetDataPermissionsByRoleIDs(ctx, roleIDs)
	if err != nil {
		return "", err
	}
	effective := DataScopeSelf
	effectiveRank := DataScopeRank[effective]
	for _, perms := range rolePerms {
		for _, p := range perms {
			scope := DataScopeFromRule(p.ScopeType)
			rank, ok := DataScopeRank[scope]
			if !ok {
				continue
			}
			if rank > effectiveRank {
				effective = scope
				effectiveRank = rank
			}
		}
	}
	return effective, nil
}

// CurrentUserMenus 获取当前登录用户可见的菜单树
func (uc *AuthenticationUsecase) CurrentUserMenus(ctx context.Context) ([]*Menu, error) {
	menus, _, err := uc.currentUserMenus(ctx)
	if err != nil {
		return nil, err
	}
	return buildMenuTree(menus), nil
}

// currentUserMenus 查询当前用户的菜单平铺列表与角色编码（superAdmin 返回全部菜单）
func (uc *AuthenticationUsecase) currentUserMenus(ctx context.Context) ([]*Menu, []string, error) {
	adminID := metadata.GetAdminID(ctx)

	// 查询用户关联的角色ID列表
	roleIDs, err := uc.adminRepo.GetRoleIDsByAdmin(ctx, adminID)
	if err != nil {
		uc.log.Errorf("获取用户 %s 的角色失败: %v", adminID, err)
		return nil, nil, err
	}

	// 获取角色编码
	roleCodes, err := uc.roleRepo.GetCodesByIds(ctx, roleIDs)
	if err != nil {
		return nil, nil, err
	}

	// 如果用户没有绑定角色，返回空菜单
	if len(roleIDs) == 0 || len(roleCodes) == 0 {
		uc.log.Warnf("用户 %s 没有绑定任何角色", adminID)
		return nil, nil, authenticationV1.ErrorNoPermission("NO_PERMISSION")
	}

	var menus []*Menu
	if slices.Contains(roleCodes, uc.appConf.SuperAdminCode) {
		menus, err = uc.menuRepo.ListMenu(ctx, &ListMenuRequest{})
		if err != nil {
			uc.log.Errorf("查询菜单详情失败: %v", err)
			return nil, nil, err
		}
	} else {
		// 查询所有角色关联的菜单ID列表（去重）
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
			return nil, nil, authenticationV1.ErrorNoPermission("NO_PERMISSION")
		}

		// 根据菜单ID列表查询完整的菜单数据
		menuIDs := make([]string, 0, len(menuIDSet))
		for id := range menuIDSet {
			menuIDs = append(menuIDs, id)
		}
		menus, err = uc.menuRepo.GetMenusByIDs(ctx, menuIDs)
		if err != nil {
			uc.log.Errorf("查询菜单详情失败: %v", err)
			return nil, nil, err
		}
	}

	// menus 按照 weight 降序排序
	sort.Slice(menus, func(i, j int) bool {
		return menus[i].Weight > menus[j].Weight
	})

	return menus, roleCodes, nil
}
