package server

import (
	"context"
	"time"

	adminV1 "github.com/antsurge/weaver-admin/api/gen/go/admin/service/v1"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/conf"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/openapi_scanner"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/handler"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/handler/pb"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/service"
	"github.com/antsurge/weaver-admin/pkg/metrics"
	apimetricsMw "github.com/antsurge/weaver-admin/pkg/middleware/apimetrics"
	"github.com/antsurge/weaver-admin/pkg/middleware/auth"
	authzMw "github.com/antsurge/weaver-admin/pkg/middleware/authz"
	"github.com/antsurge/weaver-admin/pkg/middleware/bufvalidate"
	"github.com/antsurge/weaver-admin/pkg/middleware/demo"
	ipfilterMw "github.com/antsurge/weaver-admin/pkg/middleware/ipfilter"
	"github.com/antsurge/weaver-admin/pkg/middleware/localize"
	tenantguard "github.com/antsurge/weaver-admin/pkg/middleware/tenantguard"
	authUtils "github.com/antsurge/weaver-admin/pkg/utils/auth"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/golang-jwt/jwt/v5"
)

func NewWhiteListMatcher() selector.MatchFunc {
	whiteList := make(map[string]struct{})
	whiteList["/admin.service.v1.AuthenticationService/Login"] = struct{}{}
	whiteList["/admin.service.v1.AuthenticationService/Logout"] = struct{}{}
	whiteList["/admin.service.v1.AuthenticationService/GetCaptcha"] = struct{}{}
	whiteList["/admin.service.v1.AuthenticationService/RefreshToken"] = struct{}{}
	// 文件访问：私有桶返回的是带签名的临时地址，本身已具备访问控制，
	// 且 <img> 等标签无法携带 Authorization 头，故放行。
	whiteList["/admin.service.v1.File/GetFile"] = struct{}{}
	// 注意：个人中心（GET/PUT /admin/v1/current-user 等）不放入认证白名单——
	// 其 handler 依赖 auth 中间件解析 JWT 后注入的 adminID，放行会导致
	// GetAdminID 为空、登录用户也无法访问。仅需在 authz 中 WithBypass（登录即可）
	return func(ctx context.Context, operation string) bool {
		if _, ok := whiteList[operation]; ok {
			return false
		}
		return true
	}
}

func NewDemoWhiteListMatcher() selector.MatchFunc {
	whiteList := make(map[string]struct{})
	whiteList["/admin.service.v1.AuthenticationService/Login"] = struct{}{}
	return func(ctx context.Context, operation string) bool {
		if _, ok := whiteList[operation]; ok {
			return false
		}
		return true
	}
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	c *conf.Server,
	jwtConf *conf.JWT,
	appConf *conf.App,
	tokenRepo biz.TokenRepo,
	authzUsecase *biz.AuthzUseCase,
	securityUsecase *biz.SecurityUsecase,
	tenantUsecase *biz.TenantUsecase,
	apiMetrics *metrics.Collector,
	authenticationService *service.AuthenticationService,
	permissionService *service.PermissionService,
	organizationService *service.OrganizationService,
	tenantService *service.TenantService,
	identityService *service.IdentityService,
	systemService *service.SystemService,
	fileService *service.FileService,
	messageService *service.MessageService,
	opsService *service.OpsService,
	securityService *service.SecurityService,
	lowcodeService *service.LowcodeService,

	recorder *biz.OperationLogRecorder,
	scanner *openapi_scanner.Service,

	organizationHandler *handler.OrganizationHandler,
	systemHandler *handler.SystemHandler,
	adminHandler *handler.AdminHandler,
	fileHandler *handler.FileHandler,
	notificationHandler *handler.NotificationHandler,
	codegenHandler *handler.CodegenHandler,
	profileHandler *handler.ProfileHandler,

	logger log.Logger,
) *http.Server {

	// 接口权限查询器：biz.AuthzUseCase -> authz.Checker 适配
	authzChecker := authzMw.CheckerFunc(func(ctx context.Context, adminID string) (*authzMw.Permission, error) {
		p, err := authzUsecase.GetPermission(ctx, adminID)
		if err != nil {
			return nil, err
		}
		return &authzMw.Permission{SuperAdmin: p.SuperAdmin, Codes: p.Codes}, nil
	})

	// 租户守卫：校验请求期租户状态（enabled/未过期）+ 平台超管跨租户切换。
	// 依赖 auth 注入的 adminID 与 tenantID，必须位于 auth 之后。
	tenantGuard := tenantguard.Server(
		tenantguard.WithTenantProvider(tenantUsecase),
		tenantguard.WithSuperAdminChecker(authzUsecase),
		tenantguard.WithSuperadminCrossRead(appConf.TenantSuperadminCrossRead),
		tenantguard.WithTenantMode(appConf.TenantMode),
		tenantguard.WithCacheTTL(30*time.Second),
	)

	// IP 白/黑名单校验器：策略读取自带缓存；策略 off 或查询失败时放行（fail-open）。
	ipChecker := ipfilterMw.CheckerFunc(func(ctx context.Context, ip string) (bool, error) {
		return securityUsecase.IsIPAllowed(ctx, ip)
	})

	// 接口权限中间件（RBAC 执行层）：默认不挂载（authz_enabled=false，方便权限数据
	// 未就绪时联调）；开启后 superAdmin 直通、其余按角色绑定接口权限匹配，自服务接口
	// 豁免。必须位于 auth 之后（依赖 adminID）。
	mws := []middleware.Middleware{
		localize.I18N(),
		// 接口指标采集：置于链路最前，被拦截的请求同样计入统计
		apimetricsMw.Server(
			apimetricsMw.WithCollector(apiMetrics),
			apimetricsMw.WithSkipPathPrefix("/admin/v1/notifications/stream"),
		),
		// IP 访问控制：位于 auth 之前，登录/验证码等白名单接口同样受控
		ipfilterMw.Server(ipfilterMw.WithChecker(ipChecker)),
		selector.Server(auth.Server(
			authUtils.WithSecret([]byte(jwtConf.AccessSecret)),
			authUtils.WithSigningMethod(jwt.SigningMethodHS256),
			authUtils.WithTokenStore(tokenRepo),
		)).Match(NewWhiteListMatcher()).Build(),
		// 租户守卫：校验请求期租户状态 + 平台超管跨租户切换；登录等白名单接口无租户上下文自动放行
		tenantGuard,
	}
	if appConf.AuthzEnabled {
		mws = append(mws,
			authzMw.Server(
				authzMw.WithChecker(authzChecker),
				authzMw.WithBypass("POST", "/admin/v1/logout"),
				authzMw.WithBypass("GET", "/admin/v1/current-user", "/admin/v1/current-user/menus"),
				// 个人中心：修改自己资料/密码为自服务接口，登录即可访问
				authzMw.WithBypass("PUT", "/admin/v1/current-user", "/admin/v1/current-user/password"),
				// 文件上传（头像等）：上传为通用能力，登录即可使用，无需绑定接口权限码
				authzMw.WithBypass("POST", "/admin/v1/file/upload"),
				// 消息通知：我的消息列表 / 未读数 / SSE 推送流为自服务接口，
				// 登录即可访问，无需绑定接口权限码（仍要求登录）。
				authzMw.WithBypass("GET", "/admin/v1/notifications", "/admin/v1/notifications/unread-count", "/admin/v1/notifications/stream"),
			),
		)
	}
	mws = append(mws,
		recovery.Recovery(),
		// 操作日志采集：放在 recovery 之后，能拿到真实耗时与错误
		NewOperationLogMiddleware(recorder, scanner),
		bufvalidate.BufValidator(),
		selector.Server(demo.DemoReadonly(appConf.IsDemo)).
			Match(NewDemoWhiteListMatcher()).Build(),
	)

	var opts = []http.ServerOption{
		http.Middleware(mws...),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	// 注意：自定义 handler 中带字面量路径的路由（如 /admin/v1/admin/export）必须先于
	// 参数化路由（如 /admin/v1/admin/{id}）注册，否则会被 {id} 抢先匹配，导致
	// "export" 被当作 id 查询而返回 NOT_FOUND_RECORD。
	pb.RegisterAdminHandlerServer(srv, adminHandler)
	pb.RegisterOrganizationHandlerServer(srv, organizationHandler)
	pb.RegisterSystemHandlerServer(srv, systemHandler)
	pb.RegisterFileHandlerServer(srv, fileHandler)
	pb.RegisterMessageHandlerServer(srv, notificationHandler)
	pb.RegisterCodegenHandlerServer(srv, codegenHandler)
	pb.RegisterProfileHandlerServer(srv, profileHandler)

	adminV1.RegisterAuthenticationServiceHTTPServer(srv, authenticationService)
	adminV1.RegisterPermissionServiceHTTPServer(srv, permissionService)
	adminV1.RegisterOrganizationHTTPServer(srv, organizationService)
	adminV1.RegisterTenantHTTPServer(srv, tenantService)
	adminV1.RegisterIdentityHTTPServer(srv, identityService)
	adminV1.RegisterSystemHTTPServer(srv, systemService)
	adminV1.RegisterFileHTTPServer(srv, fileService)
	adminV1.RegisterMessageHTTPServer(srv, messageService)
	adminV1.RegisterSecurityHTTPServer(srv, securityService)
	adminV1.RegisterOpsHTTPServer(srv, opsService)
	adminV1.RegisterLowcodeHTTPServer(srv, lowcodeService)

	return srv
}
