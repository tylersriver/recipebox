package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/tyler/recipebox/internal/application/command"
	"github.com/tyler/recipebox/internal/application/dto"
	"github.com/tyler/recipebox/internal/application/mapper"
	"github.com/tyler/recipebox/internal/application/query"
	"github.com/tyler/recipebox/internal/domain/entity"
	"github.com/tyler/recipebox/internal/domain/repository"
	"github.com/tyler/recipebox/internal/domain/service"
)

// RecipeService orchestrates recipe commands and queries.
type RecipeService struct {
	repo    repository.RecipeRepository
	scraper service.RecipeScraper
}

// NewRecipeService creates a new recipe service.
func NewRecipeService(repo repository.RecipeRepository, scraper service.RecipeScraper) *RecipeService {
	return &RecipeService{repo: repo, scraper: scraper}
}

// ImportRecipe scrapes a recipe from a URL and stores it.
func (s *RecipeService) ImportRecipe(ctx context.Context, userID string, cmd command.ImportRecipeCommand) (*dto.RecipeResult, error) {
	recipe, err := s.scraper.ScrapeRecipe(ctx, cmd.URL)
	if err != nil {
		return nil, fmt.Errorf("scraping recipe: %w", err)
	}

	recipe.UserID = userID

	validated, err := entity.Validate(recipe)
	if err != nil {
		return nil, fmt.Errorf("validating recipe: %w", err)
	}

	if err := s.repo.Save(ctx, validated); err != nil {
		return nil, fmt.Errorf("saving recipe: %w", err)
	}

	result := mapper.ToDTO(recipe)
	return &result, nil
}

// DeleteRecipe removes a recipe by ID.
func (s *RecipeService) DeleteRecipe(ctx context.Context, userID string, cmd command.DeleteRecipeCommand) error {
	return s.repo.Delete(ctx, userID, cmd.ID)
}

// GetRecipe retrieves a single recipe by ID.
func (s *RecipeService) GetRecipe(ctx context.Context, userID string, q query.GetRecipeByIDQuery) (*dto.RecipeResult, error) {
	recipe, err := s.repo.FindByID(ctx, userID, q.ID)
	if err != nil {
		return nil, err
	}
	if recipe == nil {
		return nil, fmt.Errorf("recipe not found: %s", q.ID)
	}
	result := mapper.ToDTO(recipe)
	return &result, nil
}

// ListRecipes returns a paginated list of recipes.
func (s *RecipeService) ListRecipes(ctx context.Context, userID string, q query.ListRecipesQuery) (*dto.RecipeListResult, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}
	recipes, total, err := s.repo.FindAll(ctx, userID, q.Offset, limit)
	if err != nil {
		return nil, err
	}
	result := mapper.ToDTOList(recipes, total, q.Offset, limit)
	return &result, nil
}

// UpdateNotes updates the personal notes for a recipe.
func (s *RecipeService) UpdateNotes(ctx context.Context, userID string, cmd command.UpdateNotesCommand) error {
	return s.repo.UpdateNotes(ctx, userID, cmd.ID, cmd.Notes)
}

// GetSharedRecipe retrieves a recipe by its public share token.
func (s *RecipeService) GetSharedRecipe(ctx context.Context, token string) (*dto.RecipeResult, error) {
	recipe, err := s.repo.FindByShareToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if recipe == nil {
		return nil, fmt.Errorf("shared recipe not found")
	}
	result := mapper.ToDTO(recipe)
	return &result, nil
}

// ShareRecipe generates a share token for a recipe, or returns the existing one.
func (s *RecipeService) ShareRecipe(ctx context.Context, userID string, cmd command.ShareRecipeCommand) (*dto.RecipeResult, error) {
	recipe, err := s.repo.FindByID(ctx, userID, cmd.ID)
	if err != nil {
		return nil, err
	}
	if recipe == nil {
		return nil, fmt.Errorf("recipe not found: %s", cmd.ID)
	}

	if recipe.ShareToken == "" {
		token, err := generateShareToken()
		if err != nil {
			return nil, fmt.Errorf("generating share token: %w", err)
		}
		if err := s.repo.SetShareToken(ctx, userID, cmd.ID, token); err != nil {
			return nil, err
		}
		recipe.ShareToken = token
	}

	result := mapper.ToDTO(recipe)
	return &result, nil
}

func generateShareToken() (string, error) {
	b := make([]byte, 9) // 9 bytes → 12 base64url chars
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// SearchRecipes searches for recipes matching a query string.
func (s *RecipeService) SearchRecipes(ctx context.Context, userID string, q query.SearchRecipesQuery) (*dto.RecipeListResult, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}
	recipes, total, err := s.repo.Search(ctx, userID, q.Query, q.Offset, limit)
	if err != nil {
		return nil, err
	}
	result := mapper.ToDTOList(recipes, total, q.Offset, limit)
	return &result, nil
}
