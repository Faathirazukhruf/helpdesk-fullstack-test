package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"helpdesk-backend/internal/auth"
	"helpdesk-backend/internal/middleware"
	"helpdesk-backend/internal/model"
	"helpdesk-backend/internal/repository"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type AuthHandler struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

func NewAuthHandler(userRepo *repository.UserRepository, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

// Register menangani registrasi akun baru (peran otomatis diset ke 'requester').
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid JSON request payload", http.StatusBadRequest)
		return
	}

	// Validasi input
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Name == "" {
		sendJSONError(w, "Name is required", http.StatusBadRequest)
		return
	}
	if len(req.Name) > 255 {
		sendJSONError(w, "Name cannot exceed 255 characters", http.StatusBadRequest)
		return
	}
	if req.Email == "" || !emailRegex.MatchString(req.Email) {
		sendJSONError(w, "A valid email address is required", http.StatusBadRequest)
		return
	}
	if len(req.Email) > 255 {
		sendJSONError(w, "Email cannot exceed 255 characters", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 6 {
		sendJSONError(w, "Password must be at least 6 characters long", http.StatusBadRequest)
		return
	}
	if len(req.Password) > 72 {
		sendJSONError(w, "Password cannot exceed 72 characters", http.StatusBadRequest)
		return
	}

	// Periksa apakah email sudah terdaftar
	existing, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err == nil && existing != nil {
		sendJSONError(w, "An account with this email already exists", http.StatusConflict)
		return
	}

	// Hash password menggunakan bcrypt
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		sendJSONError(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	newUser := &model.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         model.RoleRequester, // Selalu default requester untuk registrasi publik
	}

	if err := h.userRepo.Create(r.Context(), newUser); err != nil {
		sendJSONError(w, "Failed to create user account", http.StatusInternalServerError)
		return
	}

	// Buat token JWT untuk user yang baru terdaftar
	token, err := auth.GenerateToken(newUser, h.jwtSecret, 24*time.Hour)
	if err != nil {
		sendJSONError(w, "Account created, but failed to generate token", http.StatusInternalServerError)
		return
	}

	sendJSON(w, AuthResponse{
		Token: token,
		User:  newUser,
	}, http.StatusCreated)
}

// Login memvalidasi kredensial pengguna dan mengembalikan token JWT.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid JSON request payload", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		sendJSONError(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			sendJSONError(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}
		sendJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Verifikasi hash bcrypt password
	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		sendJSONError(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Generate JWT token berlaku 24 jam
	token, err := auth.GenerateToken(user, h.jwtSecret, 24*time.Hour)
	if err != nil {
		sendJSONError(w, "Failed to generate authentication token", http.StatusInternalServerError)
		return
	}

	sendJSON(w, AuthResponse{
		Token: token,
		User:  user,
	}, http.StatusOK)
}

// Me mengembalikan profil pengguna yang sedang login.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), claims.UserID)
	if err != nil {
		sendJSONError(w, "User not found", http.StatusNotFound)
		return
	}

	sendJSON(w, user, http.StatusOK)
}

func sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	sendJSON(w, map[string]string{"error": message}, statusCode)
}
