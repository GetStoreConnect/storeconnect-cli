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
	themeCmd.AddCommand(themePushCmd)
}

var themePushCmd = &cobra.Command{
	Use:   "push THEME_NAME",
	Short: "Upload theme to server (creates draft)",
	Long: `Upload a theme to the StoreConnect server as a draft.

This creates a ContentChange (draft) that can be previewed before publishing.
Use 'sc theme preview' to get the preview URL.
Use 'sc theme publish' to publish the draft to the live site.`,
	Args: cobra.ExactArgs(1),
	RunE: runThemePush,
}

func runThemePush(cmd *cobra.Command, args []string) error {
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

	// Deserialize theme from local filesystem (only show spinner if not JSON mode)
	var spinner *ui.Spinner
	if !jsonOutput {
		spinner = ui.NewSpinner(fmt.Sprintf("Reading theme '%s'", themeName))
		spinner.Start()
	}

	deserializer := theme.NewDeserializer(".")
	themeData, err := deserializer.Deserialize(themeName)
	if err != nil {
		if spinner != nil {
			spinner.Error(fmt.Sprintf("Failed to read theme: %v", err))
		}
		return outputError(err)
	}

	// Load sync state
	syncState, err := config.NewSyncState(fmt.Sprintf("themes/%s", themeName))
	if err != nil {
		if spinner != nil {
			spinner.Error(fmt.Sprintf("Failed to load sync state: %v", err))
		}
		return outputError(err)
	}

	// Ensure theme exists on server (use sc_id from sync state or theme.yml)
	themeID := themeData.SCID
	if themeID == "" {
		themeID = syncState.ThemeSCID
	}

	if themeID == "" {
		if spinner != nil {
			spinner.Error("Theme has no SC ID. Pull the theme first or create it on the server.")
		}
		return fmt.Errorf("theme has no sc_id")
	}

	if spinner != nil {
		spinner.UpdateMessage("Creating draft")
	}

	// Create API client
	client := api.NewClient(cred.URL, cred.StoreSFID, cred.APIKey, api.WithOrgID(cred.OrgID))
	contentChangesService := api.NewContentChanges(client)
	themesService := api.NewThemes(client)
	mediaService := api.NewMedia(client)

	// Determine which local assets need uploading by hash-diffing against the
	// server's copy of the theme. Unchanged binaries are skipped entirely.
	localAssets, err := theme.ReadLocalAssets(fmt.Sprintf("themes/%s", themeName))
	if err != nil {
		if spinner != nil {
			spinner.Error(fmt.Sprintf("Failed to read assets: %v", err))
		}
		return outputError(err)
	}

	var changedAssets []theme.LocalAsset
	if len(localAssets) > 0 {
		serverTheme, err := themesService.Get(themeID)
		if err != nil {
			if spinner != nil {
				spinner.Error(fmt.Sprintf("Failed to fetch server theme: %v", err))
			}
			return outputError(err)
		}
		changedAssets = theme.SelectChangedAssets(localAssets, serverTheme.Assets)
	}

	// Create content change (draft)
	contentChange, err := contentChangesService.Create(themeID)
	if err != nil {
		if spinner != nil {
			spinner.Error(fmt.Sprintf("Failed to create draft: %v", err))
		}
		return outputError(err)
	}

	// Upload each changed asset binary through the media flow and collect the
	// hosted URLs so they can be registered on the draft.
	assets := make([]api.ContentChangeAsset, 0, len(changedAssets))
	for _, asset := range changedAssets {
		if spinner != nil {
			spinner.UpdateMessage(fmt.Sprintf("Uploading asset %s", asset.Key))
		}

		fileType := theme.InferFileType(asset.Key)
		upload, err := mediaService.UploadURL(fileType, asset.Key, asset.ContentType)
		if err != nil {
			if spinner != nil {
				spinner.Error(fmt.Sprintf("Failed to get upload URL for %s: %v", asset.Key, err))
			}
			return outputError(err)
		}

		hostedURL, err := mediaService.UploadFile(upload, asset.Key, asset.Content)
		if err != nil {
			if spinner != nil {
				spinner.Error(fmt.Sprintf("Failed to upload asset %s: %v", asset.Key, err))
			}
			return outputError(err)
		}

		assets = append(assets, api.ContentChangeAsset{
			Key:         asset.Key,
			URL:         hostedURL,
			ContentType: asset.ContentType,
			ContentHash: asset.ContentHash,
		})
	}

	if spinner != nil {
		spinner.UpdateMessage("Uploading templates")
	}

	// Convert templates to ContentChangeTemplate format
	templates := make([]api.ContentChangeTemplate, len(themeData.Templates))
	for i, t := range themeData.Templates {
		templates[i] = api.ContentChangeTemplate{
			Key:     t.Key,
			Content: t.Content,
			Action:  "update", // Default to update
		}
	}

	// Update content change with templates and any uploaded assets
	if len(templates) > 0 || len(assets) > 0 {
		if err := contentChangesService.Update(contentChange.SCID, themeID, templates, assets); err != nil {
			if spinner != nil {
				spinner.Error(fmt.Sprintf("Failed to upload templates: %v", err))
			}
			return outputError(err)
		}
	}

	// Update sync state
	syncState.UpdatePush(contentChange.SCID)

	if spinner != nil {
		spinner.Success(fmt.Sprintf("Pushed theme '%s' as draft", themeName))
		formatter.Newline()
		formatter.Dim(fmt.Sprintf("Content Change ID: %s", contentChange.SCID))
		formatter.Dim("Run 'sc theme preview " + themeName + "' to get preview URL")
		formatter.Dim("Run 'sc theme publish " + themeName + "' to publish to live site")
	}

	// JSON output
	if jsonOutput {
		return outputResponse(ContentChangeResponse{
			ContentChangeID: contentChange.SCID,
			ThemeName:       themeName,
			Status:          "draft",
		}, nil)
	}

	return nil
}
