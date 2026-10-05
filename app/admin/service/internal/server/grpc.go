package server

import (
	"github.com/antsurge/weaver-admin/app/admin/service/internal/biz"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/conf"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/data/openapi_scanner"
	"github.com/antsurge/weaver-admin/app/admin/service/internal/service"
	auth "github.com/antsurge/weaver-admin/pkg/middleware/auth"
	"github.com/antsurge/weaver-admin/pkg/middleware/localize"
	authUtils "github.com/antsurge/weaver-admin/pkg/utils/auth"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	jwt "github.com/golang-jwt/jwt/v5"

	adminV1 "github.com/antsurge/weaver-admin/api/gen/go/admin/service/v1"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(
	c *conf.Server,
	jwtConf *conf.JWT,
	tokenRepo biz.TokenRepo,
	authenticationService *service.AuthenticationService,
	permissionService *service.PermissionService,
	organizationService *service.OrganizationService,
	tenantService *service.TenantService,
	identityService *service.IdentityService,
	systemService *service.SystemService,
	fileService *service.FileService,
	messageService *service.MessageService,
	securityService *service.SecurityService,
	opsService *service.OpsService,
	lowcodeService *service.LowcodeService,

	recorder *biz.OperationLogRecorder,
	scanner *openapi_scanner.Service,

	logger log.Logger,
) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			// 国际化：错误统一按 Reason 翻译（语言取自 metadata 的 accept-language）
			localize.I18N(),
			recovery.Recovery(),
			NewOperationLogMiddleware(recorder, scanner),
			// 认证中间件（与 HTTP 同一白名单、同一 TokenStore 校验）。
			// 接口权限（authz）暂不对 gRPC 生效：接口绑定基于
			// OpenAPI/HTTP 路径（api_interface.code），gRPC 场景待单独建模。
			selector.Server(auth.Server(
				authUtils.WithSecret([]byte(jwtConf.AccessSecret)),
				authUtils.WithSigningMethod(jwt.SigningMethodHS256),
				authUtils.WithTokenStore(tokenRepo),
			)).Match(NewWhiteListMatcher()).Build(),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)

	adminV1.RegisterAuthenticationServiceServer(srv, authenticationService)
	adminV1.RegisterPermissionServiceServer(srv, permissionService)
	adminV1.RegisterOrganizationServer(srv, organizationService)
	adminV1.RegisterTenantServer(srv, tenantService)
	adminV1.RegisterIdentityServer(srv, identityService)
	adminV1.RegisterSystemServer(srv, systemService)
	adminV1.RegisterFileServer(srv, fileService)
	adminV1.RegisterMessageServer(srv, messageService)
	adminV1.RegisterSecurityServer(srv, securityService)
	adminV1.RegisterOpsServer(srv, opsService)
	adminV1.RegisterLowcodeServer(srv, lowcodeService)

	return srv
}
