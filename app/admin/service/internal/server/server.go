package server

import (
	"github.com/antsurge/weaver-admin/app/admin/service/internal/handler"
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(
	NewGRPCServer,
	NewHTTPServer,

	handler.ProviderSet,
)
