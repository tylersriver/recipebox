package repository

import (
	"context"

	"github.com/tyler/recipebox/internal/domain/entity"
)

// UserRepository defines the persistence interface for users.
type UserRepository interface {
	Save(ctx context.Context, user *entity.User) error
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByID(ctx context.Context, id string) (*entity.User, error)
}
