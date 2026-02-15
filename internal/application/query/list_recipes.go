package query

// ListRecipesQuery retrieves a paginated list of recipes.
type ListRecipesQuery struct {
	Offset int
	Limit  int
}
