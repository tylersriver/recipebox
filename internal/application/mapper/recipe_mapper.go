package mapper

import (
	"github.com/tyler/recipebox/internal/application/dto"
	"github.com/tyler/recipebox/internal/domain/entity"
)

// ToDTO converts a domain Recipe entity to a RecipeResult DTO.
func ToDTO(r *entity.Recipe) dto.RecipeResult {
	result := dto.RecipeResult{
		ID:          r.ID,
		Title:       r.Title,
		Description: r.Description,
		PrepTime:    r.PrepTime,
		CookTime:    r.CookTime,
		TotalTime:   r.TotalTime,
		Servings:    r.Servings,
		Cuisine:     r.Cuisine,
		Course:      r.Course,
		ImageURL:    r.ImageURL,
		SourceURL:   r.SourceURL,
		Author:      r.Author,
		Nutrition:   r.Nutrition,
		Notes:       r.Notes,
		UserID:      r.UserID,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}

	for _, ing := range r.Ingredients {
		result.Ingredients = append(result.Ingredients, dto.IngredientResult{
			Amount: ing.Amount,
			Unit:   ing.Unit,
			Name:   ing.Name,
			Notes:  ing.Notes,
			Raw:    ing.Raw,
		})
	}

	for _, inst := range r.Instructions {
		result.Instructions = append(result.Instructions, dto.InstructionResult{
			StepNumber: inst.StepNumber,
			Text:       inst.Text,
		})
	}

	return result
}

// ToDTOList converts a slice of domain Recipes to a RecipeListResult.
func ToDTOList(recipes []*entity.Recipe, total, offset, limit int) dto.RecipeListResult {
	results := make([]dto.RecipeResult, 0, len(recipes))
	for _, r := range recipes {
		results = append(results, ToDTO(r))
	}
	return dto.RecipeListResult{
		Recipes: results,
		Total:   total,
		Offset:  offset,
		Limit:   limit,
	}
}
