package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewAdminUseCase,
	NewAuthenticationUsecase,
	NewCaptchaUsecase,

	NewMenuUsecase,
	NewRoleUsecase,

	NewDepartmentUsecase,
	NewPositionUsecase,

	NewDictTypeUsecase,
	NewDictDataUsecase,

	NewTenantUsecase,

	NewApiInterfaceUsecase,

	NewAuthzUseCase,

	NewFileUsecase,

	NewNotificationHub,
	NewNotificationUsecase,

	NewLogUsecase,
	NewOperationLogRecorder,

	NewOnlineUsecase,

	NewConfigUsecase,

	NewSecurityUsecase,

	NewDataPermissionUsecase,

	NewJobScheduler,
	NewHTTPExecutor,
	wire.Bind(new(JobExecutor), new(*HTTPExecutor)),
	NewCronjobUsecase,

	NewApiMetricsCollector,
	NewMonitorUsecase,

	NewFormSchemaUsecase,
	NewFormSubmissionUsecase,
	NewCodegenUsecase,
)
