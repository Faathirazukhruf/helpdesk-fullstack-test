package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"helpdesk-backend/internal/auth"
	"helpdesk-backend/internal/model"
)

type contextKey string

const (
	UserClaimsKey contextKey = "user_claims"
)

// RoleGetterFunc adalah tipe fungsi untuk mengambil peran pengguna saat ini dari database.
type RoleGetterFunc func(ctx context.Context, userID int64) (model.UserRole, error)

// AuthMiddleware memvalidasi token JWT dari header Authorization (Bearer <token>)
// dan menyuntikkan data klaim user ke dalam context request.
func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return AuthMiddlewareWithRoleGetter(secret, nil)
}

// AuthMiddlewareWithRoleGetter memvalidasi token JWT dan menyinkronkan peran aktual pengguna
// langsung dari database sehingga demotion atau penonaktifan akun berlaku seketika pada semua endpoint.
func AuthMiddlewareWithRoleGetter(secret string, roleGetter RoleGetterFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeError(w, "Invalid authorization header format (expected 'Bearer <token>')", http.StatusUnauthorized)
				return
			}

			claims, err := auth.ValidateToken(parts[1], secret)
			if err != nil {
				writeError(w, "Invalid or expired authorization token", http.StatusUnauthorized)
				return
			}

			// Sinkronkan peran pengguna langsung dari database jika roleGetter disediakan
			if roleGetter != nil {
				actualRole, err := roleGetter(r.Context(), claims.UserID)
				if err != nil {
					writeError(w, "User account not found or deactivated", http.StatusUnauthorized)
					return
				}
				claims.Role = actualRole
			}

			// Simpan claims ke dalam request context
			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole membatasi akses endpoint hanya untuk peran pengguna tertentu (RBAC).
func RequireRole(allowedRoles ...model.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetUserClaims(r.Context())
			if !ok {
				writeError(w, "Unauthorized: authentication required", http.StatusUnauthorized)
				return
			}

			allowed := false
			for _, role := range allowedRoles {
				if claims.Role == role {
					allowed = true
					break
				}
			}

			if !allowed {
				writeError(w, "Forbidden: you do not have permission to access this resource", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserClaims mengambil data user yang disuntikkan oleh AuthMiddleware dari context.
func GetUserClaims(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(*auth.Claims)
	return claims, ok
}

func writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
