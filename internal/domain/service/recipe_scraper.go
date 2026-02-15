package service

import (
	"context"

	"github.com/tyler/recipebox/internal/domain/entity"
)

// RecipeScraper defines the interface for scraping recipes from URLs.
type RecipeScraper interface {
	ScrapeRecipe(ctx context.Context, url string) (*entity.Recipe, error)
}
