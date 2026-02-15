package scraper

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tyler/recipebox/internal/domain/entity"
	"github.com/tyler/recipebox/internal/domain/service"
)

type recipeScraper struct {
	fetcher   *HTTPFetcher
	jsonLD    *JSONLDExtractor
	wprm      *WPRMExtractor
}

// NewRecipeScraper creates a new scraper orchestrator that tries JSON-LD first, then WPRM.
func NewRecipeScraper() service.RecipeScraper {
	return &recipeScraper{
		fetcher: NewHTTPFetcher(),
		jsonLD:  NewJSONLDExtractor(),
		wprm:    NewWPRMExtractor(),
	}
}

func (s *recipeScraper) ScrapeRecipe(ctx context.Context, url string) (*entity.Recipe, error) {
	doc, err := s.fetcher.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetching page: %w", err)
	}

	// Try JSON-LD first
	recipe, err := s.jsonLD.Extract(doc)
	if err == nil && recipe.Title != "" {
		return s.finalize(recipe, url)
	}

	// Fall back to WPRM
	recipe, err = s.wprm.Extract(doc)
	if err == nil && recipe.Title != "" {
		return s.finalize(recipe, url)
	}

	return nil, fmt.Errorf("could not extract recipe from %s: no JSON-LD or WPRM data found", url)
}

func (s *recipeScraper) finalize(recipe *entity.Recipe, sourceURL string) (*entity.Recipe, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	recipe.ID = id.String()
	recipe.SourceURL = sourceURL
	recipe.CreatedAt = now
	recipe.UpdatedAt = now
	return recipe, nil
}
