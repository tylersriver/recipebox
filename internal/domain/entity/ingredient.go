package entity

// Ingredient is a value object representing a single recipe ingredient.
type Ingredient struct {
	Amount string
	Unit   string
	Name   string
	Notes  string
	Raw    string // original text fallback
}
