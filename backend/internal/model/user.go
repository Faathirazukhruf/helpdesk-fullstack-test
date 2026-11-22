package model

import "time"

type UserRole string

const (
	RoleRequester UserRole = "requester"
	RoleAgent     UserRole = "agent"
	RoleAdmin     UserRole = "admin"
)

// User merepresentasikan entitas pengguna dalam sistem helpdesk.
type User struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Tanda "-" memastikan password tidak pernah bocor ke output JSON
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

// IsValid memeriksa apakah peran pengguna valid sesuai spesifikasi.
func (r UserRole) IsValid() bool {
	return r == RoleRequester || r == RoleAgent || r == RoleAdmin
}
