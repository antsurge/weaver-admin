package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/conf"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/service"

	adminV1 "github.com/hypercoze/kratos-admin/api/gen/go/admin/service/v1"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(
	c *conf.Server,
	authenticationService *service.AuthenticationService,
	permissionService *service.PermissionService,
	organizationService *service.OrganizationService,
	dictionaryService *service.DictionaryService,

	logger log.Logger,
) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
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
	adminV1.RegisterDictionaryServer(srv, dictionaryService)
	return srv
}
