package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"helpdesk-backend/internal/model"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create menyimpan pengguna baru ke database dan mengembalikan ID yang dihasilkan.
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at;`

	err := r.db.QueryRowContext(ctx, query, user.Name, user.Email, user.PasswordHash, user.Role).
		Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

// GetByEmail mencari pengguna berdasarkan alamat email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, created_at
		FROM users
		WHERE email = $1;`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, email).
		Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}

	return &user, nil
}

// GetByID mencari pengguna berdasarkan primary key ID.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, created_at
		FROM users
		WHERE id = $1;`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by id: %w", err)
	}

	return &user, nil
}

// List mengambil seluruh data pengguna dari database (diurutkan berdasarkan ID ASC).
func (r *UserRepository) List(ctx context.Context) ([]*model.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, created_at
		FROM users
		ORDER BY id ASC;`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &u)
	}

	if users == nil {
		users = []*model.User{}
	}

	return users, nil
}

// UpdateRole mengubah peran pengguna di database.
func (r *UserRepository) UpdateRole(ctx context.Context, id int64, newRole model.UserRole) (*model.User, error) {
	query := `
		UPDATE users
		SET role = $1
		WHERE id = $2
		RETURNING id, name, email, password_hash, role, created_at;`

	var u model.User
	err := r.db.QueryRowContext(ctx, query, newRole, id).
		Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to update user role: %w", err)
	}

	return &u, nil
}
