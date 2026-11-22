package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"helpdesk-backend/internal/model"
)

var (
	ErrTicketNotFound = errors.New("ticket not found")
)

type TicketRepository struct {
	db *sql.DB
}

func NewTicketRepository(db *sql.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

// Create menyimpan tiket baru ke dalam database.
func (r *TicketRepository) Create(ctx context.Context, ticket *model.Ticket) error {
	query := `
		INSERT INTO tickets (title, description, category, priority, status, requester_id, assignee_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at;`

	err := r.db.QueryRowContext(ctx, query,
		ticket.Title,
		ticket.Description,
		ticket.Category,
		ticket.Priority,
		ticket.Status,
		ticket.RequesterID,
		ticket.AssigneeID,
	).Scan(&ticket.ID, &ticket.CreatedAt, &ticket.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert ticket: %w", err)
	}

	return nil
}

// GetByID mengambil detail tiket beserta nama/email pembuat dan assignee.
func (r *TicketRepository) GetByID(ctx context.Context, id int64) (*model.TicketDetail, error) {
	query := `
		SELECT 
			t.id, t.title, t.description, t.category, t.priority, t.status,
			t.requester_id, t.assignee_id, t.created_at, t.updated_at,
			req.name AS requester_name, req.email AS requester_email,
			asg.name AS assignee_name, asg.email AS assignee_email
		FROM tickets t
		JOIN users req ON t.requester_id = req.id
		LEFT JOIN users asg ON t.assignee_id = asg.id
		WHERE t.id = $1;`

	var d model.TicketDetail
	var assigneeName, assigneeEmail sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&d.ID, &d.Title, &d.Description, &d.Category, &d.Priority, &d.Status,
		&d.RequesterID, &d.AssigneeID, &d.CreatedAt, &d.UpdatedAt,
		&d.RequesterName, &d.RequesterEmail,
		&assigneeName, &assigneeEmail,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTicketNotFound
		}
		return nil, fmt.Errorf("failed to query ticket by id: %w", err)
	}

	if assigneeName.Valid {
		d.AssigneeName = &assigneeName.String
	}
	if assigneeEmail.Valid {
		d.AssigneeEmail = &assigneeEmail.String
	}

	return &d, nil
}

// List mengambil daftar tiket berdasarkan filter, pencarian judul, dan pagination.
func (r *TicketRepository) List(ctx context.Context, f model.TicketFilter) ([]*model.TicketDetail, int64, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if f.RequesterID != nil {
		conditions = append(conditions, fmt.Sprintf("t.requester_id = $%d", argIdx))
		args = append(args, *f.RequesterID)
		argIdx++
	}

	if f.Status != "" {
		conditions = append(conditions, fmt.Sprintf("t.status = $%d", argIdx))
		args = append(args, f.Status)
		argIdx++
	}

	if f.Priority != "" {
		conditions = append(conditions, fmt.Sprintf("t.priority = $%d", argIdx))
		args = append(args, f.Priority)
		argIdx++
	}

	if f.AssigneeID != nil {
		conditions = append(conditions, fmt.Sprintf("t.assignee_id = $%d", argIdx))
		args = append(args, *f.AssigneeID)
		argIdx++
	}

	if f.Search != "" {
		conditions = append(conditions, fmt.Sprintf("t.title ILIKE $%d", argIdx))
		args = append(args, "%"+f.Search+"%")
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 1. Hitung total baris untuk pagination
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tickets t %s;", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count tickets: %w", err)
	}

	// 2. Query data dengan limit dan offset
	orderClause := "ORDER BY t.created_at DESC"
	switch f.SortBy {
	case "created_at_asc":
		orderClause = "ORDER BY t.created_at ASC"
	case "created_at_desc":
		orderClause = "ORDER BY t.created_at DESC"
	case "priority_desc":
		orderClause = "ORDER BY CASE t.priority WHEN 'urgent' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 WHEN 'low' THEN 4 ELSE 5 END ASC"
	}

	if f.Limit <= 0 {
		f.Limit = 10
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	dataQuery := fmt.Sprintf(`
		SELECT 
			t.id, t.title, t.description, t.category, t.priority, t.status,
			t.requester_id, t.assignee_id, t.created_at, t.updated_at,
			req.name AS requester_name, req.email AS requester_email,
			asg.name AS assignee_name, asg.email AS assignee_email
		FROM tickets t
		JOIN users req ON t.requester_id = req.id
		LEFT JOIN users asg ON t.assignee_id = asg.id
		%s
		%s
		LIMIT $%d OFFSET $%d;`,
		whereClause, orderClause, argIdx, argIdx+1,
	)

	dataArgs := append(args, f.Limit, offset)
	rows, err := r.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query ticket list: %w", err)
	}
	defer rows.Close()

	var tickets []*model.TicketDetail
	for rows.Next() {
		var d model.TicketDetail
		var assigneeName, assigneeEmail sql.NullString

		err := rows.Scan(
			&d.ID, &d.Title, &d.Description, &d.Category, &d.Priority, &d.Status,
			&d.RequesterID, &d.AssigneeID, &d.CreatedAt, &d.UpdatedAt,
			&d.RequesterName, &d.RequesterEmail,
			&assigneeName, &assigneeEmail,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan ticket row: %w", err)
		}

		if assigneeName.Valid {
			d.AssigneeName = &assigneeName.String
		}
		if assigneeEmail.Valid {
			d.AssigneeEmail = &assigneeEmail.String
		}

		tickets = append(tickets, &d)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return tickets, total, nil
}

// Update memperbarui status, priority, atau assignee tiket.
func (r *TicketRepository) Update(ctx context.Context, id int64, update model.TicketUpdate) (*model.TicketDetail, error) {
	var sets []string
	var args []interface{}
	argIdx := 1

	if update.Status != nil {
		sets = append(sets, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *update.Status)
		argIdx++
	}

	if update.Priority != nil {
		sets = append(sets, fmt.Sprintf("priority = $%d", argIdx))
		args = append(args, *update.Priority)
		argIdx++
	}

	if update.AssigneeID != nil {
		sets = append(sets, fmt.Sprintf("assignee_id = $%d", argIdx))
		args = append(args, *update.AssigneeID)
		argIdx++
	}

	if len(sets) == 0 {
		return r.GetByID(ctx, id)
	}

	sets = append(sets, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now())
	argIdx++

	args = append(args, id)
	query := fmt.Sprintf("UPDATE tickets SET %s WHERE id = $%d;", strings.Join(sets, ", "), argIdx)

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update ticket: %w", err)
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return nil, ErrTicketNotFound
	}

	return r.GetByID(ctx, id)
}
