package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"helpdesk-backend/internal/model"
	"helpdesk-backend/internal/repository"
)

type UserHandler struct {
	userRepo *repository.UserRepository
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

type UpdateRoleRequest struct {
	Role string `json:"role"`
}

// ListUsers mengambil daftar semua pengguna dalam sistem (Khusus Admin).
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.List(r.Context())
	if err != nil {
		sendJSONError(w, "Failed to retrieve user list", http.StatusInternalServerError)
		return
	}

	sendJSON(w, users, http.StatusOK)
}

// UpdateUserRole mengubah peran pengguna menjadi requester, agent, atau admin (Khusus Admin).
func (h *UserHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	targetUserID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		sendJSONError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid JSON request payload", http.StatusBadRequest)
		return
	}

	newRole := model.UserRole(strings.TrimSpace(strings.ToLower(req.Role)))
	if !newRole.IsValid() {
		sendJSONError(w, "Invalid role, must be: requester, agent, or admin", http.StatusBadRequest)
		return
	}

	updatedUser, err := h.userRepo.UpdateRole(r.Context(), targetUserID, newRole)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			sendJSONError(w, "User not found", http.StatusNotFound)
			return
		}
		sendJSONError(w, "Failed to update user role", http.StatusInternalServerError)
		return
	}

	sendJSON(w, updatedUser, http.StatusOK)
}
