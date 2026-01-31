package biz

import "github.com/go-kratos/kratos/v2/errors"

// 认证相关错误（避免泄露“用户名是否存在”细节，可视需求调整）
var (
	ErrInvalidArgument = errors.BadRequest("AUTH_INVALID_ARGUMENT", "invalid argument")
	ErrUnauthorized    = errors.Unauthorized("AUTH_UNAUTHORIZED", "invalid username or password")
	ErrInternal        = errors.InternalServer("AUTH_INTERNAL", "internal error")
)

