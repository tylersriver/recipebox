package command

// UpdateRecipeCommand holds the data for editing an existing recipe.
type UpdateRecipeCommand struct {
	ID           string
	Title        string
	Description  string
	PrepTime     string
	CookTime     string
	TotalTime    string
	Servings     string
	Cuisine      string
	Course       string
	ImageURL     string
	Ingredients  []IngredientInput
	Instructions []string
}
