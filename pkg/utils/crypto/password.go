package crypto

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a plaintext password using the bcrypt algorithm.
func HashPassword(password string) (string, error) {
	// GenerateFromPassword automatically handles salt generation.
	// The recommended cost is bcrypt.DefaultCost (10),
	// which provides a good balance between security and performance.
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a plaintext password with a bcrypt hash
// and returns true if they match.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
