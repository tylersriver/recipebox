package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tyler/recipebox/cmd"
)

var rootCmd = &cobra.Command{
	Use:   "recipebox",
	Short: "RecipeBox - import, store, and browse recipes",
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("db", "", "database file path (default: recipebox.db)")
	viper.BindPFlag("database.path", rootCmd.PersistentFlags().Lookup("db"))

	cmd.Register(rootCmd)
}

func initConfig() {
	viper.SetConfigName(".recipebox")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")

	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("database.path", "recipebox.db")
	viper.SetDefault("session.secret", "recipebox-default-secret-change-me")

	viper.SetEnvPrefix("RECIPEBOX")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.ReadInConfig() // ignore error if config not found

	// Honor $PORT env var (injected by Railway and other PaaS providers)
	// when RECIPEBOX_SERVER_PORT is not explicitly set.
	if os.Getenv("RECIPEBOX_SERVER_PORT") == "" {
		if p, err := strconv.Atoi(os.Getenv("PORT")); err == nil && p > 0 {
			viper.Set("server.port", p)
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
