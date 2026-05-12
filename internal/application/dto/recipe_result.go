package dto

import "time"

// RecipeResult is the DTO returned from application queries.
type RecipeResult struct {
	ID           string
	Title        string
	Description  string
	Ingredients  []IngredientResult
	Instructions []InstructionResult
	PrepTime     string
	CookTime     string
	TotalTime    string
	Servings     string
	Cuisine      string
	Course       string
	ImageURL     string
	ImagePath    string
	SourceURL    string
	Author       string
	Nutrition    string
	Notes        string
	ShareToken   string
	UserID       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// DisplayImageURL returns the best available image URL.
// Uploaded images (ImagePath) take priority over external URLs (ImageURL).
func (r RecipeResult) DisplayImageURL() string {
	if r.ImagePath != "" {
		return "/uploads/" + r.ImagePath
	}
	return r.ImageURL
}

// IngredientResult is the DTO for ingredients.
type IngredientResult struct {
	Amount string
	Unit   string
	Name   string
	Notes  string
	Raw    string
}

// InstructionResult is the DTO for instructions.
type InstructionResult struct {
	StepNumber int
	Text       string
}

// RecipeListResult holds a paginated list of recipes.
type RecipeListResult struct {
	Recipes    []RecipeResult
	Total      int
	Offset     int
	Limit      int
}
