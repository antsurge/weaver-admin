package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestJWT(t *testing.T) {
	secret := []byte("test-secret")
	userID := uuid.NewString()

	// -------------------------
	// 1. GenerateToken
	// -------------------------
	claims := NewBaseClaims(userID, 2*time.Hour, "test-service")

	tokenStr, err := GenerateToken(claims, WithSecret(secret))
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	// -------------------------
	// 2. ParseToken
	// -------------------------
	parsedClaims := &BaseClaims{}
	token, err := ParseToken(tokenStr, parsedClaims, WithSecret(secret))
	assert.NoError(t, err)
	assert.True(t, token.Valid)
	assert.Equal(t, userID, parsedClaims.UserID)

	// -------------------------
	// 3. ValidateToken
	// -------------------------
	err = ValidateToken(tokenStr, parsedClaims, WithSecret(secret))
	assert.NoError(t, err)

	// -------------------------
	// 4. RemainingTTL
	// -------------------------
	ttl, err := RemainingTTL(parsedClaims)
	assert.NoError(t, err)
	assert.True(t, ttl > 0 && ttl <= 2*time.Hour)

	// -------------------------
	// 5. Expired token scenario
	// -------------------------
	expiredClaims := NewBaseClaims(userID, -1*time.Hour, "test-service")
	expiredTokenStr, err := GenerateToken(expiredClaims, WithSecret(secret))
	assert.NoError(t, err)

	err = ValidateToken(expiredTokenStr, &BaseClaims{}, WithSecret(secret))
	assert.ErrorIs(t, err, ErrTokenExpired)

	// -------------------------
	// 6. Wrong secret scenario
	// -------------------------
	wrongSecret := []byte("wrong-secret")
	err = ValidateToken(tokenStr, &BaseClaims{}, WithSecret(wrongSecret))
	assert.ErrorIs(t, err, ErrTokenInvalid)
}
