package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tyler/recipebox/internal/application/query"
	appservice "github.com/tyler/recipebox/internal/application/service"
	"github.com/tyler/recipebox/internal/infrastructure/database"
	infrarepo "github.com/tyler/recipebox/internal/infrastructure/repository"
	"github.com/tyler/recipebox/internal/infrastructure/scraper"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search recipes by keyword",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.Open(viper.GetString("database.path"))
		if err != nil {
			return err
		}
		defer db.Close()

		repo := infrarepo.NewSQLiteRecipeRepository(db)
		scr := scraper.NewRecipeScraper()
		svc := appservice.NewRecipeService(repo, scr)

		q := strings.Join(args, " ")
		result, err := svc.SearchRecipes(context.Background(), "", query.SearchRecipesQuery{Query: q, Limit: 100})
		if err != nil {
			return err
		}

		if len(result.Recipes) == 0 {
			fmt.Printf("No recipes found for '%s'.\n", q)
			return nil
		}

		fmt.Printf("Found %d recipe(s) matching '%s':\n\n", result.Total, q)
		for _, r := range result.Recipes {
			fmt.Printf("  %s  %s\n", r.ID[:8], r.Title)
		}
		return nil
	},
}
