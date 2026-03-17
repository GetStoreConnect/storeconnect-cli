package commands

import (
	"github.com/spf13/cobra"
)

var themeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Theme management commands",
	Long:  `Commands for managing StoreConnect themes - list, pull, push, preview, and publish.`,
}

func init() {
	rootCmd.AddCommand(themeCmd)

	// Theme subcommands are registered in their respective files:
	// - theme_list.go
	// - theme_pull.go
	// - theme_push.go
	// - theme_preview.go
	// - theme_publish.go
	// - theme_delete.go
	// - theme_new.go
}
