package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// BaseClaims 通用业务 Claims
type BaseClaims struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"`
	JTI    string `json:"jti"`
	jwt.RegisteredClaims
}

func GenerateJTI() string {
	return uuid.NewString()
}

func NewAccessClaims(userID string, expire time.Duration, issuer string) *BaseClaims {
	now := time.Now()
	return &BaseClaims{
		UserID: userID,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(now.Add(expire)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
}

func NewRefreshClaims(userID string, expire time.Duration, issuer string) *BaseClaims {
	now := time.Now()
	return &BaseClaims{
		UserID: userID,
		Type:   "refresh",
		JTI:    GenerateJTI(), // 生成唯一的id
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(now.Add(expire)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
}
