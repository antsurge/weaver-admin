package crypto

import "testing"

func TestHashAndCheckPassword_Success(t *testing.T) {
	password := "P@ssw0rd123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Fatal("hash should not be empty")
	}

	if !CheckPasswordHash(password, hash) {
		t.Fatal("password should match hash")
	}
}

func TestCheckPasswordHash_Failure(t *testing.T) {
	password := "correct-password"
	wrongPassword := "wrong-password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if CheckPasswordHash(wrongPassword, hash) {
		t.Fatal("password should NOT match hash")
	}
}

func TestHashPassword_DifferentHashesForSamePassword(t *testing.T) {
	password := "same-password"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash1 == hash2 {
		t.Fatal("bcrypt should generate different hashes for same password")
	}

	if !CheckPasswordHash(password, hash1) {
		t.Fatal("password should match hash1")
	}

	if !CheckPasswordHash(password, hash2) {
		t.Fatal("password should match hash2")
	}
}

func TestHashAndCheckPassword_EmptyPassword(t *testing.T) {
	password := ""

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed for empty password: %v", err)
	}

	if !CheckPasswordHash(password, hash) {
		t.Fatal("empty password should match hash")
	}
}
