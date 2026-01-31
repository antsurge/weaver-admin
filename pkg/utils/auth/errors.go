package auth

import (
	"github.com/go-kratos/kratos/v2/errors"
	"net/http"
)

var (
	ErrTokenExpired = errors.New(
		http.StatusUnauthorized,
		"JWT_TOKEN_EXPIRED",
		"jwt token expired",
	)

	ErrTokenInvalid = errors.New(
		http.StatusUnauthorized,
		"JWT_TOKEN_INVALID",
		"jwt token invalid",
	)

	ErrTokenMalformed = errors.New(
		http.StatusUnauthorized,
		"JWT_TOKEN_MALFORMED",
		"jwt token malformed",
	)

	ErrSigningMethod = errors.New(
		http.StatusUnauthorized,
		"JWT_SIGNING_METHOD_INVALID",
		"unexpected jwt signing method",
	)
)
