package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/golang-jwt/jwt/v5"
	adminV1 "github.com/hypercoze/kratos-admin/api/gen/go/admin/service/v1"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/conf"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/service"
	"github.com/hypercoze/kratos-admin/pkg/middleware/auth"
	"github.com/hypercoze/kratos-admin/pkg/middleware/bufvalidate"
	"github.com/hypercoze/kratos-admin/pkg/middleware/localize"
	authUtils "github.com/hypercoze/kratos-admin/pkg/utils/auth"
)

func NewWhiteListMatcher() selector.MatchFunc {
	whiteList := make(map[string]struct{})
	whiteList["/admin.service.v1.AuthenticationService/Login"] = struct{}{}
	whiteList["/admin.service.v1.AuthenticationService/GetCaptcha"] = struct{}{}
	whiteList["/admin.service.v1.AuthenticationService/RefreshToken"] = struct{}{}
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
	authenticationService *service.AuthenticationService,
	permissionService *service.PermissionService,
	organizationService *service.OrganizationService,
	dictionaryService *service.DictionaryService,

	logger log.Logger,
) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			localize.I18N(),
			selector.Server(auth.Server(
				authUtils.WithSecret([]byte(jwtConf.AccessSecret)),
				authUtils.WithSigningMethod(jwt.SigningMethodHS256),
			)).Match(NewWhiteListMatcher()).Build(),
			recovery.Recovery(),
			bufvalidate.BufValidator(),
		),
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
	adminV1.RegisterAuthenticationServiceHTTPServer(srv, authenticationService)
	adminV1.RegisterPermissionServiceHTTPServer(srv, permissionService)
	adminV1.RegisterOrganizationHTTPServer(srv, organizationService)
	adminV1.RegisterDictionaryHTTPServer(srv, dictionaryService)
	return srv
}
