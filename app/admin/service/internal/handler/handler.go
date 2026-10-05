package handler

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewOrganizationHandler,
	NewSystemHandler,
	NewAdminHandler,
	NewFileHandler,
	NewNotificationHandler,
	NewCodegenHandler,
	NewProfileHandler,
)
