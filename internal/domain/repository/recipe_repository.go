package repository

import (
	"context"

	"github.com/tyler/recipebox/internal/domain/entity"
)

// RecipeRepository defines the persistence interface for recipes.
type RecipeRepository interface {
	Save(ctx context.Context, recipe *entity.ValidatedRecipe) error
	FindByID(ctx context.Context, id string) (*entity.Recipe, error)
	FindBySourceURL(ctx context.Context, sourceURL string) (*entity.Recipe, error)
	FindAll(ctx context.Context, offset, limit int) ([]*entity.Recipe, int, error)
	Search(ctx context.Context, query string, offset, limit int) ([]*entity.Recipe, int, error)
	Delete(ctx context.Context, id string) error
}
