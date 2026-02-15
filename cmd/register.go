package cmd

import "github.com/spf13/cobra"

// Register adds all subcommands to the root command.
func Register(root *cobra.Command) {
	root.AddCommand(serveCmd)
	root.AddCommand(listCmd)
	root.AddCommand(searchCmd)
	root.AddCommand(showCmd)
}
