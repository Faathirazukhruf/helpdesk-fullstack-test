package model

import (
	"errors"
	"fmt"
	"time"
)

type TicketPriority string

const (
	PriorityLow    TicketPriority = "low"
	PriorityMedium TicketPriority = "medium"
	PriorityHigh   TicketPriority = "high"
	PriorityUrgent TicketPriority = "urgent"
)

type TicketStatus string

const (
	StatusOpen       TicketStatus = "Open"
	StatusInProgress TicketStatus = "In Progress"
	StatusResolved   TicketStatus = "Resolved"
	StatusClosed     TicketStatus = "Closed"
)

var (
	ErrInvalidPriority   = errors.New("invalid priority, must be: low, medium, high, urgent")
	ErrInvalidStatus     = errors.New("invalid status, must be: Open, In Progress, Resolved, Closed")
	ErrSameStatus        = errors.New("ticket is already in this status")
	ErrIllegalTransition = errors.New("illegal status transition")
)

// Ticket merepresentasikan tabel tickets di database.
type Ticket struct {
	ID          int64          `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Category    string         `json:"category"`
	Priority    TicketPriority `json:"priority"`
	Status      TicketStatus   `json:"status"`
	RequesterID int64          `json:"requester_id"`
	AssigneeID  *int64         `json:"assignee_id"` // Pointer memperbolehkan nilai null jika belum di-assign
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// TicketDetail menyertakan informasi relasi pengguna pembuat, agent yang menangani, serta riwayat komentar.
type TicketDetail struct {
	Ticket
	RequesterName  string           `json:"requester_name"`
	RequesterEmail string           `json:"requester_email"`
	AssigneeName   *string          `json:"assignee_name,omitempty"`
	AssigneeEmail  *string          `json:"assignee_email,omitempty"`
	Comments       []*CommentDetail `json:"comments,omitempty"`
}

// TicketFilter menampung parameter pencarian, filter, dan pagination daftar tiket.
type TicketFilter struct {
	Status      string
	Priority    string
	AssigneeID  *int64
	RequesterID *int64 // Digunakan untuk membatasi requester agar hanya melihat tiketnya sendiri
	Search      string
	SortBy      string // created_at_asc, created_at_desc, priority_desc
	Page        int
	Limit       int
}

// TicketUpdate menampung field yang boleh diperbarui via PATCH.
type TicketUpdate struct {
	Status     *TicketStatus   `json:"status,omitempty"`
	Priority   *TicketPriority `json:"priority,omitempty"`
	AssigneeID *int64          `json:"assignee_id,omitempty"`
}

// IsValidPriority memvalidasi apakah nilai prioritas tiket diizinkan.
func (p TicketPriority) IsValid() bool {
	return p == PriorityLow || p == PriorityMedium || p == PriorityHigh || p == PriorityUrgent
}

// IsValidStatus memvalidasi apakah nilai status tiket diizinkan.
func (s TicketStatus) IsValid() bool {
	return s == StatusOpen || s == StatusInProgress || s == StatusResolved || s == StatusClosed
}

// ValidateStatusTransition memvalidasi apakah perpindahan status diizinkan berdasarkan
// aturan baku spesifikasi:
// - Open -> In Progress (Agent, Admin)
// - In Progress -> Resolved (Agent, Admin)
// - Resolved -> Closed (Requester, Agent, Admin)
// - Resolved -> Open (Requester)
// - Open -> Closed (Requester, Admin)
func ValidateStatusTransition(from, to TicketStatus, role UserRole) error {
	if !to.IsValid() {
		return ErrInvalidStatus
	}
	if from == to {
		return ErrSameStatus
	}

	switch {
	// Open -> In Progress (Oleh: Agent, Admin)
	case from == StatusOpen && to == StatusInProgress:
		if role != RoleAgent && role != RoleAdmin {
			return fmt.Errorf("%w: role %s cannot move ticket from Open to In Progress", ErrIllegalTransition, role)
		}
		return nil

	// In Progress -> Resolved (Oleh: Agent, Admin)
	case from == StatusInProgress && to == StatusResolved:
		if role != RoleAgent && role != RoleAdmin {
			return fmt.Errorf("%w: role %s cannot move ticket from In Progress to Resolved", ErrIllegalTransition, role)
		}
		return nil

	// Resolved -> Closed (Oleh: Requester, Agent, Admin)
	case from == StatusResolved && to == StatusClosed:
		// Semua peran boleh menutup tiket yang sudah resolved
		return nil

	// Resolved -> Open (Oleh: Requester saja)
	case from == StatusResolved && to == StatusOpen:
		if role != RoleRequester {
			return fmt.Errorf("%w: only requester can reopen a resolved ticket", ErrIllegalTransition)
		}
		return nil

	// Open -> Closed (Oleh: Requester [membatalkan], Admin)
	case from == StatusOpen && to == StatusClosed:
		if role != RoleRequester && role != RoleAdmin {
			return fmt.Errorf("%w: role %s cannot cancel/close an open ticket", ErrIllegalTransition, role)
		}
		return nil

	default:
		return fmt.Errorf("%w: transition from '%s' to '%s' is not allowed", ErrIllegalTransition, from, to)
	}
}
