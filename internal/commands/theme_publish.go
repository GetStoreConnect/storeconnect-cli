package commands

import (
	"fmt"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"github.com/GetStoreConnect/storeconnect-cli/internal/config"
	"github.com/GetStoreConnect/storeconnect-cli/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	themeCmd.AddCommand(themePublishCmd)
}

var themePublishCmd = &cobra.Command{
	Use:   "publish THEME_NAME",
	Short: "Publish draft theme to live site",
	Long: `Publish a draft theme to the live site.

You must push the theme first using 'sc theme push THEME_NAME'.
On development/staging environments, this auto-publishes.
On production environments, this may require manual approval in Salesforce.`,
	Args: cobra.ExactArgs(1),
	RunE: runThemePublish,
}

func runThemePublish(cmd *cobra.Command, args []string) error {
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

	// Publish content change (only show spinner if not JSON mode)
	var spinner *ui.Spinner
	if !jsonOutput {
		spinner = ui.NewSpinner("Publishing theme")
		spinner.Start()
	}

	publishResp, err := contentChangesService.Publish(syncState.ContentChangeSCID)
	if err != nil {
		if spinner != nil {
			spinner.Error(fmt.Sprintf("Failed to publish theme: %v", err))
		}
		return outputError(err)
	}

	contentChangeID := syncState.ContentChangeSCID

	// Check if this was submitted for approval or actually published
	if publishResp.Status == "review" {
		// Submitted for approval - keep content change in sync state
		if spinner != nil {
			spinner.Stop()
			formatter.Info(publishResp.Message)
			if publishResp.Note != "" {
				formatter.Dim(publishResp.Note)
			}
			formatter.Newline()
			formatter.Dim(fmt.Sprintf("Content change ID: %s", contentChangeID))
			formatter.Dim("The theme will be published after approval in Salesforce")
		}
	} else {
		// Actually published - clear content change from sync state
		syncState.ClearContentChange()

		if spinner != nil {
			spinner.Success(publishResp.Message)
		}
	}

	// JSON output
	if jsonOutput {
		return outputResponse(ThemePublishResponse{
			ThemeName:       themeName,
			ContentChangeID: contentChangeID,
			Status:          publishResp.Status,
			Message:         publishResp.Message,
		}, nil)
	}

	return nil
}
