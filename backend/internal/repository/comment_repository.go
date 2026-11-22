package repository

import (
	"context"
	"database/sql"
	"fmt"

	"helpdesk-backend/internal/model"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create menyimpan komentar baru ke dalam database.
func (r *CommentRepository) Create(ctx context.Context, comment *model.Comment) error {
	query := `
		INSERT INTO comments (ticket_id, author_id, body, is_internal)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at;`

	err := r.db.QueryRowContext(ctx, query,
		comment.TicketID,
		comment.AuthorID,
		comment.Body,
		comment.IsInternal,
	).Scan(&comment.ID, &comment.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert comment: %w", err)
	}

	return nil
}

// ListByTicketID mengambil daftar komentar untuk suatu tiket.
// Parameter includeInternal menentukan apakah catatan internal staf ditampilkan (true untuk agent/admin, false untuk requester).
func (r *CommentRepository) ListByTicketID(ctx context.Context, ticketID int64, includeInternal bool) ([]*model.CommentDetail, error) {
	query := `
		SELECT 
			c.id, c.ticket_id, c.author_id, c.body, c.is_internal, c.created_at,
			u.name AS author_name, u.email AS author_email, u.role AS author_role
		FROM comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.ticket_id = $1`

	if !includeInternal {
		query += " AND c.is_internal = false"
	}

	query += " ORDER BY c.created_at ASC;"

	rows, err := r.db.QueryContext(ctx, query, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	var comments []*model.CommentDetail
	for rows.Next() {
		var d model.CommentDetail
		err := rows.Scan(
			&d.ID, &d.TicketID, &d.AuthorID, &d.Body, &d.IsInternal, &d.CreatedAt,
			&d.AuthorName, &d.AuthorEmail, &d.AuthorRole,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment row: %w", err)
		}
		comments = append(comments, &d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("comment rows error: %w", err)
	}

	if comments == nil {
		comments = []*model.CommentDetail{}
	}

	return comments, nil
}
