package auth

import (
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

/**
 * GenerateToken
 * @Description: Generates a JWT token string using the provided claims and options.
 *                It handles the signing with the configured secret and signing method.
 * @param claims The JWT claims to include in the token (jwt.MapClaims).
 * @param opts   Optional configurations such as secret, expiration, issuer, and signing method.
 * @return string The signed JWT token string.
 * @return error  Returns an error if signing fails or secret is not provided.
 */
func GenerateToken(claims jwt.Claims, opts ...Option) (string, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	if len(o.Secret) == 0 {
		return "", ErrTokenInvalid
	}

	token := jwt.NewWithClaims(o.SigningMethod, claims)
	return token.SignedString(o.Secret)
}

/**
 * ParseToken
 * @Description: Parses and validates a JWT token string into the provided claims.
 *                It checks the signing method and verifies the token signature using the configured secret.
 * @param tokenStr The JWT token string to parse.
 * @param claims   The JWT claims structure to populate (must implement jwt.Claims).
 * @param opts     Optional configurations such as secret, signing method, and issuer.
 * @return *jwt.Token The parsed JWT token object.
 * @return error     Returns an error if the token is invalid, malformed, expired, or the signing method is unexpected.
 */
func ParseToken(tokenStr string, claims jwt.Claims, opts ...Option) (*jwt.Token, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	return jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != o.SigningMethod.Alg() {
				return nil, ErrSigningMethod
			}
			return o.Secret, nil
		},
	)
}

/**
 * ValidateToken
 * @Description: Validates a JWT token string by parsing it into the provided claims
 *                and checking its signature, expiration, and overall validity.
 *                Returns a semantic error if the token is expired, malformed, or otherwise invalid.
 * @param tokenString The JWT token string to validate.
 * @param claims      The JWT claims structure to populate (must implement jwt.Claims).
 * @param opts        Optional configurations such as secret, signing method, and issuer.
 * @return error      Returns a Kratos-style error (ErrTokenExpired, ErrTokenMalformed, ErrTokenInvalid)
 *                    if the token fails validation, otherwise returns nil.
 */
func ValidateToken(tokenString string, claims jwt.Claims, opts ...Option) error {
	token, err := ParseToken(tokenString, claims, opts...)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return ErrTokenMalformed
		}
		return ErrTokenInvalid
	}

	if !token.Valid {
		return ErrTokenInvalid
	}

	return nil
}

/**
 * RemainingTTL
 * @Description: Calculates the remaining time-to-live (TTL) of a JWT token based on its expiration time.
 *                This is a generic function that works with any type implementing jwt.Claims.
 * @param claims The JWT claims instance (any type implementing jwt.Claims) to check the expiration time.
 * @return time.Duration The duration until the token expires. Returns a negative value if already expired.
 * @return error         Returns ErrTokenInvalid if the expiration time cannot be determined or claims are invalid.
 */
func RemainingTTL[T jwt.Claims](claims T) (time.Duration, error) {
	exp, err := claims.GetExpirationTime()
	if err != nil || exp == nil {
		return 0, ErrTokenInvalid
	}
	return time.Until(exp.Time), nil
}
