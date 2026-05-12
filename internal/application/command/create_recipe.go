package command

// CreateRecipeCommand holds the data for manually creating a recipe.
type CreateRecipeCommand struct {
	Title        string
	Description  string
	PrepTime     string
	CookTime     string
	TotalTime    string
	Servings     string
	Cuisine      string
	Course       string
	ImageURL     string
	ImagePath    string
	Ingredients  []IngredientInput
	Instructions []string
}

// IngredientInput holds a single ingredient's fields from user input.
type IngredientInput struct {
	Raw string
}
