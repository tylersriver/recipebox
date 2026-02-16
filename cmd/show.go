package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tyler/recipebox/internal/application/dto"
	"github.com/tyler/recipebox/internal/application/query"
	appservice "github.com/tyler/recipebox/internal/application/service"
	"github.com/tyler/recipebox/internal/infrastructure/database"
	infrarepo "github.com/tyler/recipebox/internal/infrastructure/repository"
	"github.com/tyler/recipebox/internal/infrastructure/scraper"
)

var showCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show full details of a recipe",
	Long:  "Display a recipe's ingredients, instructions, and metadata. Pass a full or partial recipe ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.Open(viper.GetString("database.path"))
		if err != nil {
			return err
		}
		defer db.Close()

		repo := infrarepo.NewSQLiteRecipeRepository(db)
		scr := scraper.NewRecipeScraper()
		svc := appservice.NewRecipeService(repo, scr)

		id := args[0]

		// Try exact match first
		result, err := svc.GetRecipe(context.Background(), "", query.GetRecipeByIDQuery{ID: id})
		if err != nil {
			// Try prefix match by listing all and filtering
			result, err = findByPrefix(svc, id)
			if err != nil {
				return fmt.Errorf("recipe not found: %s", id)
			}
		}

		printRecipe(result)
		return nil
	},
}

func findByPrefix(svc *appservice.RecipeService, prefix string) (*dto.RecipeResult, error) {
	list, err := svc.ListRecipes(context.Background(), "", query.ListRecipesQuery{Limit: 1000})
	if err != nil {
		return nil, err
	}
	for _, r := range list.Recipes {
		if strings.HasPrefix(r.ID, prefix) {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("no match")
}

func printRecipe(r *dto.RecipeResult) {
	fmt.Printf("%-12s %s\n", "Title:", r.Title)
	if r.Author != "" {
		fmt.Printf("%-12s %s\n", "Author:", r.Author)
	}
	if r.SourceURL != "" {
		fmt.Printf("%-12s %s\n", "Source:", r.SourceURL)
	}
	fmt.Printf("%-12s %s\n", "ID:", r.ID)

	if r.Servings != "" || r.PrepTime != "" || r.CookTime != "" || r.TotalTime != "" {
		fmt.Println()
		if r.Servings != "" {
			fmt.Printf("%-12s %s\n", "Servings:", r.Servings)
		}
		if r.PrepTime != "" {
			fmt.Printf("%-12s %s\n", "Prep Time:", r.PrepTime)
		}
		if r.CookTime != "" {
			fmt.Printf("%-12s %s\n", "Cook Time:", r.CookTime)
		}
		if r.TotalTime != "" {
			fmt.Printf("%-12s %s\n", "Total Time:", r.TotalTime)
		}
	}

	if r.Cuisine != "" || r.Course != "" {
		fmt.Println()
		if r.Cuisine != "" {
			fmt.Printf("%-12s %s\n", "Cuisine:", r.Cuisine)
		}
		if r.Course != "" {
			fmt.Printf("%-12s %s\n", "Course:", r.Course)
		}
	}

	if r.Description != "" {
		fmt.Printf("\n%s\n", r.Description)
	}

	if len(r.Ingredients) > 0 {
		fmt.Printf("\nIngredients\n%s\n", strings.Repeat("-", 40))
		for _, ing := range r.Ingredients {
			if ing.Raw != "" {
				fmt.Printf("  - %s\n", ing.Raw)
			} else {
				line := ""
				if ing.Amount != "" {
					line += ing.Amount + " "
				}
				if ing.Unit != "" {
					line += ing.Unit + " "
				}
				line += ing.Name
				if ing.Notes != "" {
					line += ", " + ing.Notes
				}
				fmt.Printf("  - %s\n", line)
			}
		}
	}

	if len(r.Instructions) > 0 {
		fmt.Printf("\nInstructions\n%s\n", strings.Repeat("-", 40))
		for _, inst := range r.Instructions {
			fmt.Printf("  %d. %s\n\n", inst.StepNumber, inst.Text)
		}
	}
}
