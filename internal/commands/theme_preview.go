package commands

import (
	"fmt"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"github.com/GetStoreConnect/storeconnect-cli/internal/config"
	"github.com/GetStoreConnect/storeconnect-cli/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	themeCmd.AddCommand(themePreviewCmd)
}

var themePreviewCmd = &cobra.Command{
	Use:   "preview THEME_NAME",
	Short: "Get preview URL for draft theme",
	Long: `Get the preview URL for a draft theme.

You must push the theme first using 'sc theme push THEME_NAME'.
The preview URL allows you to view changes before publishing to the live site.`,
	Args: cobra.ExactArgs(1),
	RunE: runThemePreview,
}

func runThemePreview(cmd *cobra.Command, args []string) error {
	themeName := args[0]
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

	// Load sync state
	syncState, err := config.NewSyncState(fmt.Sprintf("themes/%s", themeName))
	if err != nil {
		if !jsonOutput {
			formatter.Error(fmt.Sprintf("Failed to load sync state: %v", err))
		}
		return outputError(err)
	}

	if syncState.ContentChangeSCID == "" {
		if !jsonOutput {
			formatter.Error(fmt.Sprintf("No draft found for theme '%s'", themeName))
			formatter.Dim("Run 'sc theme push " + themeName + "' first")
		}
		return fmt.Errorf("no draft found")
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
	contentChangesService := api.NewContentChanges(client)

	// Get preview URL (only show spinner if not JSON mode)
	var spinner *ui.Spinner
	if !jsonOutput {
		spinner = ui.NewSpinner("Generating preview URL")
		spinner.Start()
	}

	previewURL, err := contentChangesService.GetPreviewURL(syncState.ContentChangeSCID)
	if err != nil {
		if spinner != nil {
			spinner.Error(fmt.Sprintf("Failed to get preview URL: %v", err))
		}
		return outputError(err)
	}

	if spinner != nil {
		spinner.Success("Preview URL generated")
	}

	// JSON output
	if jsonOutput {
		return outputResponse(PreviewURLResponse{
			PreviewURL:      previewURL,
			ContentChangeID: syncState.ContentChangeSCID,
			ThemeID:         syncState.ThemeSCID,
		}, nil)
	}

	// Human-friendly output
	formatter.Newline()
	formatter.Info("Preview URL:")
	fmt.Println(previewURL)
	formatter.Newline()
	formatter.Dim("Open this URL in your browser to preview the theme")
	formatter.Dim("Run 'sc theme publish " + themeName + "' when ready to publish")

	return nil
}
