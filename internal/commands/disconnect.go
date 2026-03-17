package commands

import (
	"fmt"

	"github.com/GetStoreConnect/storeconnect-cli/internal/config"
	"github.com/GetStoreConnect/storeconnect-cli/internal/ui"
	"github.com/spf13/cobra"
)

var disconnectCmd = &cobra.Command{
	Use:   "disconnect ALIAS",
	Short: "Remove server connection from project",
	Long:  `Remove a server connection from both project config and global credentials.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDisconnect,
}

func init() {
	rootCmd.AddCommand(disconnectCmd)
}

func runDisconnect(cmd *cobra.Command, args []string) error {
	alias := args[0]
	formatter := ui.NewFormatter()

	// Load credentials
	credentials, err := config.NewCredentials()
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to load credentials: %v", err))
		return err
	}

	// Load project config
	project, err := config.NewProject()
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to load project config: %v", err))
		return err
	}

	// Remove from project if it exists
	if project.Exists() {
		if _, ok := project.GetServer(alias); ok {
			if err := project.RemoveServer(alias); err != nil {
				formatter.Error(fmt.Sprintf("Failed to remove from project: %v", err))
				return err
			}
			formatter.Success(fmt.Sprintf("Removed '%s' from project config", alias))
		}
	}

	// Remove from credentials
	if _, ok := credentials.GetServer(alias); ok {
		if err := credentials.RemoveServer(alias); err != nil {
			formatter.Error(fmt.Sprintf("Failed to remove credentials: %v", err))
			return err
		}
		formatter.Success(fmt.Sprintf("Removed '%s' credentials", alias))
	} else {
		formatter.Warning(fmt.Sprintf("Server '%s' not found in credentials", alias))
	}

	return nil
}
