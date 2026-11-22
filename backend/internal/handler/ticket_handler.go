package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"helpdesk-backend/internal/middleware"
	"helpdesk-backend/internal/model"
	"helpdesk-backend/internal/repository"
)

type TicketHandler struct {
	ticketRepo  *repository.TicketRepository
	userRepo    *repository.UserRepository
	commentRepo *repository.CommentRepository
}

func NewTicketHandler(ticketRepo *repository.TicketRepository, userRepo *repository.UserRepository, commentRepo *repository.CommentRepository) *TicketHandler {
	return &TicketHandler{
		ticketRepo:  ticketRepo,
		userRepo:    userRepo,
		commentRepo: commentRepo,
	}
}

type CreateTicketRequest struct {
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Category    string               `json:"category"`
	Priority    model.TicketPriority `json:"priority"`
}

type UpdateTicketRequest struct {
	Status     *model.TicketStatus   `json:"status"`
	Priority   *model.TicketPriority `json:"priority"`
	AssigneeID *int64                `json:"assignee_id"`
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type TicketListResponse struct {
	Data       []*model.TicketDetail `json:"data"`
	Pagination PaginationMeta        `json:"pagination"`
}

// CreateTicket menangani pembuatan tiket baru oleh requester atau staf.
func (h *TicketHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid JSON request payload", http.StatusBadRequest)
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	req.Category = strings.TrimSpace(req.Category)

	if req.Title == "" {
		sendJSONError(w, "Title is required", http.StatusBadRequest)
		return
	}
	if len(req.Title) > 255 {
		sendJSONError(w, "Title cannot exceed 255 characters", http.StatusBadRequest)
		return
	}
	if req.Description == "" {
		sendJSONError(w, "Description is required", http.StatusBadRequest)
		return
	}
	if req.Category == "" {
		sendJSONError(w, "Category is required", http.StatusBadRequest)
		return
	}

	if req.Priority == "" {
		req.Priority = model.PriorityMedium
	} else if !req.Priority.IsValid() {
		sendJSONError(w, "Invalid priority, must be: low, medium, high, urgent", http.StatusBadRequest)
		return
	}

	ticket := &model.Ticket{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Priority:    req.Priority,
		Status:      model.StatusOpen, // Tiket baru selalu dimulai dari status Open
		RequesterID: claims.UserID,
		AssigneeID:  nil,
	}

	if err := h.ticketRepo.Create(r.Context(), ticket); err != nil {
		sendJSONError(w, "Failed to create ticket", http.StatusInternalServerError)
		return
	}

	detail, err := h.ticketRepo.GetByID(r.Context(), ticket.ID)
	if err != nil {
		sendJSON(w, ticket, http.StatusCreated)
		return
	}

	sendJSON(w, detail, http.StatusCreated)
}

// ListTickets mengambil daftar tiket dengan filter, pencarian judul, dan pagination.
// Aturan RBAC: Requester hanya diizinkan melihat tiket miliknya sendiri.
func (h *TicketHandler) ListTickets(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()
	page := 1
	if pStr := q.Get("page"); pStr != "" {
		p, err := strconv.Atoi(pStr)
		if err != nil || p <= 0 || p > 1000000 {
			sendJSONError(w, "Invalid page parameter, must be between 1 and 1,000,000", http.StatusBadRequest)
			return
		}
		page = p
	}

	limit := 10
	if lStr := q.Get("limit"); lStr != "" {
		l, err := strconv.Atoi(lStr)
		if err != nil || l <= 0 || l > 100 {
			sendJSONError(w, "Invalid limit parameter, must be between 1 and 100", http.StatusBadRequest)
			return
		}
		limit = l
	}

	filter := model.TicketFilter{
		Status:   q.Get("status"),
		Priority: q.Get("priority"),
		Search:   q.Get("search"),
		SortBy:   q.Get("sort"),
		Page:     page,
		Limit:    limit,
	}

	if asgStr := q.Get("assignee_id"); asgStr != "" {
		if asgID, err := strconv.ParseInt(asgStr, 10, 64); err == nil {
			filter.AssigneeID = &asgID
		}
	}

	// Pembatasan Akses: Requester dipaksa hanya melihat tiket miliknya sendiri
	if claims.Role == model.RoleRequester {
		filter.RequesterID = &claims.UserID
	}

	tickets, total, err := h.ticketRepo.List(r.Context(), filter)
	if err != nil {
		sendJSONError(w, "Failed to retrieve tickets", http.StatusInternalServerError)
		return
	}

	if tickets == nil {
		tickets = []*model.TicketDetail{}
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages == 0 {
		totalPages = 1
	}

	sendJSON(w, TicketListResponse{
		Data: tickets,
		Pagination: PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, http.StatusOK)
}

// GetTicket mengambil detail satu tiket berdasarkan ID di URL path.
// Aturan RBAC: Requester dilarang melihat tiket milik requester lain.
func (h *TicketHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Menggunakan fitur PathValue bawaan router Go 1.22
	idStr := r.PathValue("id")
	ticketID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		sendJSONError(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}

	ticket, err := h.ticketRepo.GetByID(r.Context(), ticketID)
	if err != nil {
		if errors.Is(err, repository.ErrTicketNotFound) {
			sendJSONError(w, "Ticket not found", http.StatusNotFound)
			return
		}
		sendJSONError(w, "Failed to get ticket", http.StatusInternalServerError)
		return
	}

	// Pembatasan Hak Akses: Requester dilarang melihat tiket milik orang lain
	if claims.Role == model.RoleRequester && ticket.RequesterID != claims.UserID {
		sendJSONError(w, "Forbidden: you do not have permission to view this ticket", http.StatusForbidden)
		return
	}

	// Ambil komentar untuk tiket ini.
	// Jika user adalah agent/admin (isStaff = true), sertakan catatan internal.
	// Jika requester (isStaff = false), catatan internal disembunyikan secara otomatis.
	isStaff := claims.Role == model.RoleAgent || claims.Role == model.RoleAdmin
	comments, err := h.commentRepo.ListByTicketID(r.Context(), ticketID, isStaff)
	if err == nil {
		ticket.Comments = comments
	}

	sendJSON(w, ticket, http.StatusOK)
}

// UpdateTicket menangani perubahan status, prioritas, atau penugasan (assignee) tiket.
// Memvalidasi kepemilikan dan aturan transisi status tiket secara ketat.
func (h *TicketHandler) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserClaims(r.Context())
	if !ok {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	ticketID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		sendJSONError(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}

	var req UpdateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid JSON request payload", http.StatusBadRequest)
		return
	}

	// Ambil data tiket saat ini dari database
	currentTicket, err := h.ticketRepo.GetByID(r.Context(), ticketID)
	if err != nil {
		if errors.Is(err, repository.ErrTicketNotFound) {
			sendJSONError(w, "Ticket not found", http.StatusNotFound)
			return
		}
		sendJSONError(w, "Failed to retrieve ticket", http.StatusInternalServerError)
		return
	}

	// Validasi Hak Akses Khusus Requester
	if claims.Role == model.RoleRequester {
		// Requester hanya boleh mengupdate tiket miliknya sendiri
		if currentTicket.RequesterID != claims.UserID {
			sendJSONError(w, "Forbidden: you can only update your own tickets", http.StatusForbidden)
			return
		}

		// Requester tidak diizinkan mengubah prioritas tiket
		if req.Priority != nil {
			sendJSONError(w, "Forbidden: requesters are not allowed to change ticket priority", http.StatusForbidden)
			return
		}

		// Requester tidak diizinkan menetapkan assignee tiket
		if req.AssigneeID != nil {
			sendJSONError(w, "Forbidden: requesters are not allowed to assign tickets", http.StatusForbidden)
			return
		}
	}

	// Validasi Transisi Status (Jika ada permintaan perubahan status)
	if req.Status != nil {
		if err := model.ValidateStatusTransition(currentTicket.Status, *req.Status, claims.Role); err != nil {
			sendJSONError(w, fmt.Sprintf("Invalid status transition: %v", err), http.StatusUnprocessableEntity)
			return
		}
	}

	// Validasi Prioritas (Jika diubah oleh Agent/Admin)
	if req.Priority != nil {
		if !req.Priority.IsValid() {
			sendJSONError(w, "Invalid priority, must be: low, medium, high, urgent", http.StatusBadRequest)
			return
		}
	}

	// Validasi Assignee (Jika diubah oleh Agent/Admin)
	if req.AssigneeID != nil {
		assignee, err := h.userRepo.GetByID(r.Context(), *req.AssigneeID)
		if err != nil {
			sendJSONError(w, "Assignee user not found", http.StatusBadRequest)
			return
		}
		if assignee.Role != model.RoleAgent && assignee.Role != model.RoleAdmin {
			sendJSONError(w, "Ticket can only be assigned to an agent or admin", http.StatusBadRequest)
			return
		}
	}

	updateData := model.TicketUpdate{
		Status:     req.Status,
		Priority:   req.Priority,
		AssigneeID: req.AssigneeID,
	}

	updatedTicket, err := h.ticketRepo.Update(r.Context(), ticketID, updateData)
	if err != nil {
		sendJSONError(w, "Failed to update ticket", http.StatusInternalServerError)
		return
	}

	sendJSON(w, updatedTicket, http.StatusOK)
}
