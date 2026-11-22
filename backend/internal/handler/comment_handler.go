package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"helpdesk-backend/internal/middleware"
	"helpdesk-backend/internal/model"
	"helpdesk-backend/internal/repository"
)

type CommentHandler struct {
	commentRepo *repository.CommentRepository
	ticketRepo  *repository.TicketRepository
}

func NewCommentHandler(commentRepo *repository.CommentRepository, ticketRepo *repository.TicketRepository) *CommentHandler {
	return &CommentHandler{
		commentRepo: commentRepo,
		ticketRepo:  ticketRepo,
	}
}

type CreateCommentRequest struct {
	Body       string `json:"body"`
	IsInternal bool   `json:"is_internal"`
}

// CreateComment menambahkan komentar publik atau catatan internal pada suatu tiket.
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
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

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid JSON request payload", http.StatusBadRequest)
		return
	}

	req.Body = strings.TrimSpace(req.Body)
	if req.Body == "" {
		sendJSONError(w, "Comment body cannot be empty", http.StatusBadRequest)
		return
	}

	// Periksa keberadaan tiket
	ticket, err := h.ticketRepo.GetByID(r.Context(), ticketID)
	if err != nil {
		if errors.Is(err, repository.ErrTicketNotFound) {
			sendJSONError(w, "Ticket not found", http.StatusNotFound)
			return
		}
		sendJSONError(w, "Failed to retrieve ticket", http.StatusInternalServerError)
		return
	}

	// Aturan Hak Akses Komentar:
	// 1. Requester hanya boleh berkomentar pada tiket miliknya sendiri
	if claims.Role == model.RoleRequester && ticket.RequesterID != claims.UserID {
		sendJSONError(w, "Forbidden: you can only comment on your own tickets", http.StatusForbidden)
		return
	}

	// 2. Requester dilarang membuat catatan internal (hanya boleh dibuat oleh agent/admin)
	if claims.Role == model.RoleRequester && req.IsInternal {
		sendJSONError(w, "Forbidden: requesters cannot create internal notes", http.StatusForbidden)
		return
	}

	comment := &model.Comment{
		TicketID:   ticketID,
		AuthorID:   claims.UserID,
		Body:       req.Body,
		IsInternal: req.IsInternal,
	}

	if err := h.commentRepo.Create(r.Context(), comment); err != nil {
		sendJSONError(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	// Kembalikan detail komentar lengkap dengan info penulis
	commentDetail := &model.CommentDetail{
		Comment:     *comment,
		AuthorName:  claims.Email, // default fallback
		AuthorEmail: claims.Email,
		AuthorRole:  claims.Role,
	}

	sendJSON(w, commentDetail, http.StatusCreated)
}
