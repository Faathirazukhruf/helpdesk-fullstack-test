package model

import "time"

// Comment merepresentasikan baris pada tabel comments di database.
type Comment struct {
	ID         int64     `json:"id"`
	TicketID   int64     `json:"ticket_id"`
	AuthorID   int64     `json:"author_id"`
	Body       string    `json:"body"`
	IsInternal bool      `json:"is_internal"`
	CreatedAt  time.Time `json:"created_at"`
}

// CommentDetail menyertakan informasi identitas penulis komentar (nama, email, role).
type CommentDetail struct {
	Comment
	AuthorName  string   `json:"author_name"`
	AuthorEmail string   `json:"author_email"`
	AuthorRole  UserRole `json:"author_role"`
}
