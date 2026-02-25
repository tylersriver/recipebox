package repository

import (
	"context"

	"github.com/tyler/recipebox/internal/domain/entity"
)

// RecipeRepository defines the persistence interface for recipes.
type RecipeRepository interface {
	Save(ctx context.Context, recipe *entity.ValidatedRecipe) error
	FindByID(ctx context.Context, userID, id string) (*entity.Recipe, error)
	FindBySourceURL(ctx context.Context, userID, sourceURL string) (*entity.Recipe, error)
	FindAll(ctx context.Context, userID string, offset, limit int) ([]*entity.Recipe, int, error)
	Search(ctx context.Context, userID, query string, offset, limit int) ([]*entity.Recipe, int, error)
	Delete(ctx context.Context, userID, id string) error
	UpdateNotes(ctx context.Context, userID, id string, notes string) error
	UpdateImagePath(ctx context.Context, userID, id string, imagePath string) error
	FindByShareToken(ctx context.Context, token string) (*entity.Recipe, error)
	SetShareToken(ctx context.Context, userID, id, token string) error
}
