package auth_test

import (
	"testing"
	"time"

	"helpdesk-backend/internal/auth"
	"helpdesk-backend/internal/model"
)

// TestPasswordHashing menguji keamanan proses hashing dan verifikasi password bcrypt.
func TestPasswordHashing(t *testing.T) {
	password := "SecretPassword123!"

	hash1, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() returned unexpected error: %v", err)
	}

	if hash1 == "" || hash1 == password {
		t.Fatalf("HashPassword() returned invalid hash result: %s", hash1)
	}

	// Verifikasi bahwa password benar menghasilkan kecocokan
	if !auth.CheckPassword(password, hash1) {
		t.Errorf("CheckPassword() failed to verify valid password")
	}

	// Verifikasi bahwa password salah ditolak
	if auth.CheckPassword("WrongPassword!", hash1) {
		t.Errorf("CheckPassword() unexpectedly accepted incorrect password")
	}

	// Verifikasi salt: dua kali hashing password yang sama menghasilkan hash berbeda
	hash2, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() second call returned error: %v", err)
	}
	if hash1 == hash2 {
		t.Errorf("HashPassword() should use random salt and generate different hashes")
	}
}

// TestJWTTokenLifecycle menguji siklus hidup pembuatan, validasi klaim, dan proteksi JWT.
func TestJWTTokenLifecycle(t *testing.T) {
	secret := "test-secret-key-for-unit-testing"
	wrongSecret := "another-secret-key-different"

	user := &model.User{
		ID:    42,
		Email: "tester@example.com",
		Role:  model.RoleAgent,
	}

	// 1. Pembuatan Token Valid
	token, err := auth.GenerateToken(user, secret, 1*time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken() returned unexpected error: %v", err)
	}
	if token == "" {
		t.Fatalf("GenerateToken() returned empty token string")
	}

	// 2. Validasi Token dengan Secret yang Benar
	claims, err := auth.ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateToken() returned unexpected error: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("claims.UserID = %d, want %d", claims.UserID, user.ID)
	}
	if claims.Email != user.Email {
		t.Errorf("claims.Email = %s, want %s", claims.Email, user.Email)
	}
	if claims.Role != user.Role {
		t.Errorf("claims.Role = %s, want %s", claims.Role, user.Role)
	}

	// 3. Validasi dengan Secret yang Salah (Harus Ditolak)
	_, err = auth.ValidateToken(token, wrongSecret)
	if err == nil {
		t.Errorf("ValidateToken() with wrong secret should have failed")
	}

	// 4. Validasi Token Kadaluarsa (Harus Ditolak)
	expiredToken, err := auth.GenerateToken(user, secret, -1*time.Minute)
	if err != nil {
		t.Fatalf("Failed to generate expired token: %v", err)
	}
	_, err = auth.ValidateToken(expiredToken, secret)
	if err == nil {
		t.Errorf("ValidateToken() with expired token should have failed")
	}
}
