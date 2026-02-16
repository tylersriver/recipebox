package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tyler/recipebox/internal/domain/entity"
	"github.com/tyler/recipebox/internal/domain/repository"
)

type sqliteUserRepository struct {
	db *sql.DB
}

// NewSQLiteUserRepository creates a new SQLite-backed user repository.
func NewSQLiteUserRepository(db *sql.DB) repository.UserRepository {
	return &sqliteUserRepository{db: db}
}

func (r *sqliteUserRepository) Save(ctx context.Context, user *entity.User) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`,
		user.ID, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("saving user: %w", err)
	}
	return nil
}

func (r *sqliteUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email = ?", email)
	user := &entity.User{}
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding user by email: %w", err)
	}
	return user, nil
}

func (r *sqliteUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, email, password_hash, created_at, updated_at FROM users WHERE id = ?", id)
	user := &entity.User{}
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding user by id: %w", err)
	}
	return user, nil
}
