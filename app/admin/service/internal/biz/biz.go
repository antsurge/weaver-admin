package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewAdminUseCase,
	NewAuthenticationUsecase,
	NewCaptchaUsecase,

	NewMenuUsecase,
	NewRoleUseCase,

	NewDepartmentUsecase,
	NewPositionUsecase,

	NewDictTypeUsecase,
	NewDictDataUsecase,
)
