package commands

import (
	"fmt"

	"github.com/GetStoreConnect/storeconnect-cli/internal/config"
	"github.com/GetStoreConnect/storeconnect-cli/internal/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init PROJECT_NAME",
	Short: "Create new StoreConnect project directory",
	Long: `Initialize a new StoreConnect project directory with the necessary
structure for theme development.

This command creates:
  • PROJECT_NAME/ - Root project directory
  • .storeconnect/ - Configuration directory
  • themes/ - Directory for theme files

After initialization, cd into the project directory and run 'sc connect'
to connect to your StoreConnect server.`,
	Args: cobra.ExactArgs(1),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	projectName := args[0]
	formatter := ui.NewFormatter()

	// Initialize project
	project := &config.Project{}
	if err := project.Init(projectName); err != nil {
		formatter.Error(fmt.Sprintf("Failed to initialize project: %v", err))
		return err
	}

	formatter.Success(fmt.Sprintf("Created project '%s'", projectName))
	formatter.Newline()
	formatter.Dim(fmt.Sprintf("Next steps:"))
	formatter.Dim(fmt.Sprintf("  cd %s", projectName))
	formatter.Dim("  sc connect <URL> --alias <name>")

	return nil
}
