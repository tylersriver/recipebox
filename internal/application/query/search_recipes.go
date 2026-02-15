package query

// SearchRecipesQuery searches recipes by a text query.
type SearchRecipesQuery struct {
	Query  string
	Offset int
	Limit  int
}
