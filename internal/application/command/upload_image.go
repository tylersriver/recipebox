package command

// UploadImageCommand holds the data for uploading a recipe photo.
type UploadImageCommand struct {
	RecipeID string
	Filename string
}
