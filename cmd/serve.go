package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	appservice "github.com/tyler/recipebox/internal/application/service"
	"github.com/tyler/recipebox/internal/infrastructure/database"
	infrarepo "github.com/tyler/recipebox/internal/infrastructure/repository"
	"github.com/tyler/recipebox/internal/infrastructure/scraper"
	"github.com/tyler/recipebox/internal/interface/web"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web server",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.Open(viper.GetString("database.path"))
		if err != nil {
			return err
		}
		defer db.Close()

		repo := infrarepo.NewSQLiteRecipeRepository(db)
		scr := scraper.NewRecipeScraper()
		svc := appservice.NewRecipeService(repo, scr)

		host := viper.GetString("server.host")
		port := viper.GetInt("server.port")
		addr := fmt.Sprintf("%s:%d", host, port)
		sessionSecret := viper.GetString("session.secret")
		uploadsDir := viper.GetString("storage.uploads_dir")
		if err := os.MkdirAll(uploadsDir, 0755); err != nil {
			return fmt.Errorf("creating uploads directory: %w", err)
		}

		srv := web.NewServer(svc, db, addr, sessionSecret, uploadsDir)
		fmt.Printf("RecipeBox server starting on http://%s\n", addr)
		return srv.Start()
	},
}
