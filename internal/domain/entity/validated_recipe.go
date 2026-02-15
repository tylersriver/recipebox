package entity

import "errors"

// ValidatedRecipe wraps a Recipe that has passed domain validation.
type ValidatedRecipe struct {
	recipe *Recipe
}

// Validate checks domain invariants and returns a ValidatedRecipe.
func Validate(r *Recipe) (*ValidatedRecipe, error) {
	if r.ID == "" {
		return nil, errors.New("recipe ID is required")
	}
	if r.Title == "" {
		return nil, errors.New("recipe title is required")
	}
	if r.SourceURL == "" {
		return nil, errors.New("recipe source URL is required")
	}
	return &ValidatedRecipe{recipe: r}, nil
}

// Recipe returns the underlying validated recipe.
func (v *ValidatedRecipe) Recipe() *Recipe {
	return v.recipe
}
