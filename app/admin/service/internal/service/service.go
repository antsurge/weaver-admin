package service

import "github.com/google/wire"

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	NewAuthenticationService,
	NewPermissionService,

	NewOrganizationService,

	NewTenantService,

	NewIdentityService,

	NewSystemService,
	NewFileService,

	NewMessageService,

	NewSecurityService,

	NewOpsService,

	NewLowcodeService,
)
