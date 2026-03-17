package commands

import (
	"fmt"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"github.com/GetStoreConnect/storeconnect-cli/internal/config"
	"github.com/GetStoreConnect/storeconnect-cli/internal/theme"
	"github.com/GetStoreConnect/storeconnect-cli/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	themeCmd.AddCommand(themePullCmd)
}

var themePullCmd = &cobra.Command{
	Use:   "pull THEME_NAME",
	Short: "Download theme from server",
	Long:  `Download a theme from the StoreConnect server to your local themes directory.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runThemePull,
}

func runThemePull(cmd *cobra.Command, args []string) error {
	themeIdentifier := args[0]
	formatter := ui.NewFormatter()

	// Get server alias
	serverAlias, _ := cmd.Flags().GetString("server")
	if serverAlias == "" {
		project, err := config.NewProject()
		if err != nil {
			if !jsonOutput {
				formatter.Error(fmt.Sprintf("Failed to load project: %v", err))
			}
			return outputError(err)
		}
		serverAlias = project.GetDefaultServer()
	}

	if serverAlias == "" {
		if !jsonOutput {
			formatter.Error("No server specified and no default server set")
		}
		return fmt.Errorf("no server configured")
	}

	// Load credentials
	credentials, err := config.NewCredentials()
	if err != nil {
		if !jsonOutput {
			formatter.Error(fmt.Sprintf("Failed to load credentials: %v", err))
		}
		return outputError(err)
	}

	cred, ok := credentials.GetServer(serverAlias)
	if !ok {
		if !jsonOutput {
			formatter.Error(fmt.Sprintf("Server '%s' not found in credentials", serverAlias))
		}
		return fmt.Errorf("server not found")
	}

	// Create API client
	client := api.NewClient(cred.URL, cred.StoreSFID, cred.APIKey, api.WithOrgID(cred.OrgID))
	themesService := api.NewThemes(client)

	// Lookup theme by name if not a UUID (only show spinner if not JSON mode)
	var spinner *ui.Spinner
	if !jsonOutput {
		spinner = ui.NewSpinner(fmt.Sprintf("Downloading theme '%s'", themeIdentifier))
		spinner.Start()
	}

	themeID := themeIdentifier
	// If it doesn't look like a UUID, try to find by name
	if !isUUID(themeIdentifier) {
		themes, err := themesService.List()
		if err != nil {
			if spinner != nil {
				spinner.Error(fmt.Sprintf("Failed to list themes: %v", err))
			}
			return outputError(err)
		}

		found := false
		for _, theme := range themes {
			if theme.Name == themeIdentifier {
				themeID = theme.SCID
				found = true
				break
			}
		}

		if !found {
			if spinner != nil {
				spinner.Error(fmt.Sprintf("Theme '%s' not found", themeIdentifier))
			}
			return fmt.Errorf("theme not found")
		}
	}

	// Fetch theme
	themeData, err := themesService.Get(themeID)
	if err != nil {
		if spinner != nil {
			spinner.Error(fmt.Sprintf("Failed to fetch theme: %v", err))
		}
		return outputError(err)
	}

	if spinner != nil {
		spinner.UpdateMessage("Writing theme files")
	}

	// Serialize theme to local filesystem
	serializer := theme.NewSerializer(".")
	if err := serializer.Serialize(themeData); err != nil {
		if spinner != nil {
			spinner.Error(fmt.Sprintf("Failed to write theme: %v", err))
		}
		return outputError(err)
	}

	// Update sync state (use actual theme name from API response)
	syncState, _ := config.NewSyncState(fmt.Sprintf("themes/%s", themeData.Name))
	if syncState != nil {
		syncState.UpdatePull(themeData.SCID, themeData.SFID)
	}

	// Count files
	filesCount := len(themeData.Templates) + len(themeData.Assets)

	if spinner != nil {
		spinner.Success(fmt.Sprintf("Downloaded theme '%s' to themes/%s", themeData.Name, themeData.Name))
	}

	// JSON output
	if jsonOutput {
		return outputResponse(ThemePullResponse{
			ThemeName:  themeData.Name,
			ThemeID:    themeData.SCID,
			FilesCount: filesCount,
		}, nil)
	}

	return nil
}
