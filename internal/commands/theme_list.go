package commands

import (
	"fmt"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"github.com/GetStoreConnect/storeconnect-cli/internal/config"
	"github.com/GetStoreConnect/storeconnect-cli/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	themeCmd.AddCommand(themeListCmd)
}

var themeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all themes",
	Long:  `List all custom themes for the connected StoreConnect server.`,
	RunE:  runThemeList,
}

func runThemeList(cmd *cobra.Command, args []string) error {
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
		return outputError(fmt.Errorf("no server configured"))
	}

	// Load credentials and project
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
		return outputError(fmt.Errorf("server not found"))
	}

	// Create API client
	client := api.NewClient(cred.URL, cred.StoreSFID, cred.APIKey, api.WithOrgID(cred.OrgID))
	themesService := api.NewThemes(client)

	// Fetch themes (only show spinner if not JSON mode)
	var spinner *ui.Spinner
	if !jsonOutput {
		spinner = ui.NewSpinner("Fetching themes")
		spinner.Start()
	}

	themes, err := themesService.List()
	if err != nil {
		if spinner != nil {
			spinner.Error(fmt.Sprintf("Failed to fetch themes: %v", err))
		}
		return outputError(err)
	}

	if spinner != nil {
		spinner.Stop()
	}

	// JSON output
	if jsonOutput {
		themeInfos := make([]ThemeInfo, len(themes))
		for i, t := range themes {
			themeInfos[i] = ThemeInfo{
				Name: t.Name,
				SCID: t.SCID,
				SFID: t.SFID,
			}
		}
		return outputResponse(ThemeListResponse{
			Themes: themeInfos,
			Total:  len(themes),
		}, nil)
	}

	// Human-friendly output
	if len(themes) == 0 {
		formatter.Warning("No themes found")
		return nil
	}

	formatter.Info(fmt.Sprintf("Themes on %s:", serverAlias))
	fmt.Println()

	for _, theme := range themes {
		fmt.Printf("  • %s\n", theme.Name)
		if theme.SCID != "" {
			formatter.Dim(fmt.Sprintf("    SC ID: %s", theme.SCID))
		}
	}

	fmt.Println()
	formatter.Dim(fmt.Sprintf("Total: %d themes", len(themes)))

	return nil
}
