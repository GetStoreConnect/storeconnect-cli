package commands

import (
	"fmt"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"github.com/GetStoreConnect/storeconnect-cli/internal/config"
	"github.com/GetStoreConnect/storeconnect-cli/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	themeCmd.AddCommand(themeSubmitCmd)
}

var themeSubmitCmd = &cobra.Command{
	Use:   "submit THEME_NAME",
	Short: "Submit draft theme for review in Salesforce",
	Long: `Submit a draft theme for review without publishing it.

The draft becomes visible to the store's admins in Salesforce (the Change
Request flow) but is not live. Use 'sc theme publish THEME_NAME' when the
change should go live.

You must push the theme first using 'sc theme push THEME_NAME'.`,
	Args: cobra.ExactArgs(1),
	RunE: runThemeSubmit,
}

func runThemeSubmit(cmd *cobra.Command, args []string) error {
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

	var spinner *ui.Spinner
	if !jsonOutput {
		spinner = ui.NewSpinner("Submitting theme for review")
		spinner.Start()
	}

	contentChange, err := contentChangesService.Submit(syncState.ContentChangeSCID)
	if err != nil {
		if spinner != nil {
			spinner.Error(fmt.Sprintf("Failed to submit theme: %v", err))
		}
		return outputError(err)
	}

	if spinner != nil {
		spinner.Success("Theme submitted for review")
		formatter.Newline()
		formatter.Dim(fmt.Sprintf("Content change ID: %s (status: %s)", contentChange.SCID, contentChange.Status))
		formatter.Dim("Admins can review the change in Salesforce; publish it with 'sc theme publish " + themeName + "'")
	}

	if jsonOutput {
		return outputResponse(map[string]interface{}{
			"theme_name":        themeName,
			"content_change_id": contentChange.SCID,
			"status":            contentChange.Status,
		}, nil)
	}

	return nil
}
