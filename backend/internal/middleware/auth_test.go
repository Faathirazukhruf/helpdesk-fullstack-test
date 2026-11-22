package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"helpdesk-backend/internal/auth"
	"helpdesk-backend/internal/middleware"
	"helpdesk-backend/internal/model"
)

// TestAuthMiddleware menguji pemeriksaan token pada header Authorization oleh AuthMiddleware.
func TestAuthMiddleware(t *testing.T) {
	secret := "test-secret-key-12345"
	authMW := middleware.AuthMiddleware(secret)

	// Dummy handler yang mengembalikan claims user dari request context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.GetUserClaims(r.Context())
		if !ok {
			http.Error(w, "No claims in context", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(claims)
	})

	handlerToTest := authMW(testHandler)

	t.Run("Tanpa Authorization Header (Harus 401 Unauthorized)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
		rec := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status code = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("Format Header Bukan Bearer (Harus 401 Unauthorized)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
		req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
		rec := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status code = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("Token Palsu atau Rusak (Harus 401 Unauthorized)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid.jwt.token")
		rec := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status code = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("Token Valid (Harus 200 OK dan Klaim Tersedia di Context)", func(t *testing.T) {
		user := &model.User{ID: 10, Email: "user@example.com", Role: model.RoleRequester}
		token, err := auth.GenerateToken(user, secret, time.Hour)
		if err != nil {
			t.Fatalf("GenerateToken failed: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status code = %d, want %d", rec.Code, http.StatusOK)
		}

		var claims auth.Claims
		if err := json.NewDecoder(rec.Body).Decode(&claims); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if claims.UserID != user.ID {
			t.Errorf("claims.UserID = %d, want %d", claims.UserID, user.ID)
		}
	})
}

// TestRequireRole menguji penegakan pembatasan hak akses berbasis peran (RBAC).
func TestRequireRole(t *testing.T) {
	secret := "test-secret-key-12345"
	authMW := middleware.AuthMiddleware(secret)
	adminOnlyMW := middleware.RequireRole(model.RoleAdmin)

	targetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Welcome Admin!"))
	})

	// Rantai middleware: AuthMiddleware -> RequireRole(Admin) -> TargetHandler
	chain := authMW(adminOnlyMW(targetHandler))

	t.Run("Admin Mengakses Endpoint Admin (Harus 200 OK)", func(t *testing.T) {
		adminUser := &model.User{ID: 1, Email: "admin@example.com", Role: model.RoleAdmin}
		adminToken, _ := auth.GenerateToken(adminUser, secret, time.Hour)

		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		rec := httptest.NewRecorder()

		chain.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status code = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("Requester Mengakses Endpoint Admin (Harus 403 Forbidden)", func(t *testing.T) {
		reqUser := &model.User{ID: 2, Email: "requester@example.com", Role: model.RoleRequester}
		reqToken, _ := auth.GenerateToken(reqUser, secret, time.Hour)

		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		req.Header.Set("Authorization", "Bearer "+reqToken)
		rec := httptest.NewRecorder()

		chain.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status code = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("Agent Mengakses Endpoint Admin (Harus 403 Forbidden)", func(t *testing.T) {
		agentUser := &model.User{ID: 3, Email: "agent@example.com", Role: model.RoleAgent}
		agentToken, _ := auth.GenerateToken(agentUser, secret, time.Hour)

		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		req.Header.Set("Authorization", "Bearer "+agentToken)
		rec := httptest.NewRecorder()

		chain.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status code = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})
}
