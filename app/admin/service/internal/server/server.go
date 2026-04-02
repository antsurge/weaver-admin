package server

import (
	"github.com/google/wire"
	"github.com/hypercoze/kratos-admin/app/admin/service/internal/handler"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(
	NewGRPCServer,
	NewHTTPServer,

	handler.ProviderSet,
)
