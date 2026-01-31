package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// BaseClaims 通用业务 Claims
type BaseClaims struct {
	UserID string `json:"user_id"`
	JTI    string `json:"jti"`
	jwt.RegisteredClaims
}

func GenerateJTI() string {
	return uuid.NewString()
}

func NewBaseClaims(
	userID string,
	expire time.Duration,
	issuer string,
) *BaseClaims {
	now := time.Now()

	return &BaseClaims{
		UserID: userID,
		JTI:    GenerateJTI(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(now.Add(expire)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
}
