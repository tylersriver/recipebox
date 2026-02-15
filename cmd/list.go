package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tyler/recipebox/internal/application/query"
	appservice "github.com/tyler/recipebox/internal/application/service"
	"github.com/tyler/recipebox/internal/infrastructure/database"
	infrarepo "github.com/tyler/recipebox/internal/infrastructure/repository"
	"github.com/tyler/recipebox/internal/infrastructure/scraper"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stored recipes",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.Open(viper.GetString("database.path"))
		if err != nil {
			return err
		}
		defer db.Close()

		repo := infrarepo.NewSQLiteRecipeRepository(db)
		scr := scraper.NewRecipeScraper()
		svc := appservice.NewRecipeService(repo, scr)

		result, err := svc.ListRecipes(context.Background(), query.ListRecipesQuery{Limit: 100})
		if err != nil {
			return err
		}

		if len(result.Recipes) == 0 {
			fmt.Println("No recipes found. Import one from the web UI.")
			return nil
		}

		fmt.Printf("Found %d recipe(s):\n\n", result.Total)
		for _, r := range result.Recipes {
			fmt.Printf("  %s  %s\n", r.ID[:8], r.Title)
		}
		return nil
	},
}
