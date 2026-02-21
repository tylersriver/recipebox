package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Recipe is the core domain entity.
type Recipe struct {
	ID           string
	Title        string
	Description  string
	Ingredients  []Ingredient
	Instructions []Instruction
	PrepTime     string
	CookTime     string
	TotalTime    string
	Servings     string
	Cuisine      string
	Course       string
	ImageURL     string
	SourceURL    string
	Author       string
	Nutrition    string
	Notes        string
	ShareToken   string
	UserID       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewRecipe creates a new Recipe with a generated UUID v7 and timestamps.
func NewRecipe(title, sourceURL string) (*Recipe, error) {
	if title == "" {
		return nil, errors.New("recipe title is required")
	}
	if sourceURL == "" {
		return nil, errors.New("recipe source URL is required")
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Recipe{
		ID:        id.String(),
		Title:     title,
		SourceURL: sourceURL,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// NewManualRecipe creates a new Recipe for manually-entered data.
// It generates a unique source URL in the form "manual://<id>".
func NewManualRecipe(title string) (*Recipe, error) {
	if title == "" {
		return nil, errors.New("recipe title is required")
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	idStr := id.String()
	now := time.Now().UTC()
	return &Recipe{
		ID:        idStr,
		Title:     title,
		SourceURL: "manual://" + idStr,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
