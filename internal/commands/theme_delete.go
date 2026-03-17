package commands

import (
	"fmt"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"github.com/GetStoreConnect/storeconnect-cli/internal/config"
	"github.com/GetStoreConnect/storeconnect-cli/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	themeCmd.AddCommand(themeDeleteCmd)
}

var themeDeleteCmd = &cobra.Command{
	Use:   "delete THEME_NAME",
	Short: "Delete theme from server",
	Long:  `Delete a theme from the StoreConnect server. This does not delete local files.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runThemeDelete,
}

func runThemeDelete(cmd *cobra.Command, args []string) error {
	themeIdentifier := args[0]
	formatter := ui.NewFormatter()

	// Get server alias
	serverAlias, _ := cmd.Flags().GetString("server")
	if serverAlias == "" {
		project, err := config.NewProject()
		if err != nil {
			formatter.Error(fmt.Sprintf("Failed to load project: %v", err))
			return err
		}
		serverAlias = project.GetDefaultServer()
	}

	if serverAlias == "" {
		formatter.Error("No server specified and no default server set")
		return fmt.Errorf("no server configured")
	}

	// Load credentials
	credentials, err := config.NewCredentials()
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to load credentials: %v", err))
		return err
	}

	cred, ok := credentials.GetServer(serverAlias)
	if !ok {
		formatter.Error(fmt.Sprintf("Server '%s' not found in credentials", serverAlias))
		return fmt.Errorf("server not found")
	}

	// Create API client
	client := api.NewClient(cred.URL, cred.StoreSFID, cred.APIKey, api.WithOrgID(cred.OrgID))
	themesService := api.NewThemes(client)

	// Lookup theme by name if not a UUID
	themeID := themeIdentifier
	themeName := themeIdentifier
	if !isUUID(themeIdentifier) {
		themes, err := themesService.List()
		if err != nil {
			formatter.Error(fmt.Sprintf("Failed to list themes: %v", err))
			return err
		}

		found := false
		for _, theme := range themes {
			if theme.Name == themeIdentifier {
				themeID = theme.SCID
				themeName = theme.Name
				found = true
				break
			}
		}

		if !found {
			formatter.Error(fmt.Sprintf("Theme '%s' not found", themeIdentifier))
			return fmt.Errorf("theme not found")
		}
	}

	// Delete theme
	spinner := ui.NewSpinner(fmt.Sprintf("Deleting theme '%s'", themeName))
	spinner.Start()

	if err := themesService.Delete(themeID); err != nil {
		spinner.Error(fmt.Sprintf("Failed to delete theme: %v", err))
		return err
	}

	spinner.Success(fmt.Sprintf("Deleted theme '%s' from server", themeName))
	formatter.Newline()
	formatter.Dim("Local files in themes/" + themeName + " were not deleted")

	return nil
}
