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
	themeCmd.AddCommand(themeNewCmd)
}

var themeNewCmd = &cobra.Command{
	Use:   "new THEME_NAME",
	Short: "Create a new theme",
	Long:  `Create a new empty theme on the server and download it locally.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runThemeNew,
}

func runThemeNew(cmd *cobra.Command, args []string) error {
	themeName := args[0]
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

	// Create theme on server
	spinner := ui.NewSpinner(fmt.Sprintf("Creating theme '%s'", themeName))
	spinner.Start()

	themeData, err := themesService.Create(themeName)
	if err != nil {
		spinner.Error(fmt.Sprintf("Failed to create theme: %v", err))
		return err
	}

	spinner.UpdateMessage("Downloading theme")

	// Serialize theme to local filesystem
	serializer := theme.NewSerializer(".")
	if err := serializer.Serialize(themeData); err != nil {
		spinner.Error(fmt.Sprintf("Failed to write theme: %v", err))
		return err
	}

	// Update sync state
	syncState, _ := config.NewSyncState(fmt.Sprintf("themes/%s", themeName))
	if syncState != nil {
		syncState.UpdatePull(themeData.SCID, themeData.SFID)
	}

	spinner.Success(fmt.Sprintf("Created theme '%s'", themeName))
	formatter.Newline()
	formatter.Dim(fmt.Sprintf("Theme files created in themes/%s", themeName))

	return nil
}
