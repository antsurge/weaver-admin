package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewAdminUseCase,
	NewAuthenticationUsecase,
	NewCaptchaUsecase,

	NewPermissionUsecase,

	NewOrganizationUsecase,
	NewPositionUsecase,

	NewDictTypeUsecase,
	NewDictDataUsecase,
)
