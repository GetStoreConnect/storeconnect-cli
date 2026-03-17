package commands

import (
	"fmt"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"github.com/GetStoreConnect/storeconnect-cli/internal/ui"
	"github.com/spf13/cobra"
)

var mediaCmd = &cobra.Command{
	Use:   "media",
	Short: "Media management commands",
	Long:  `Commands for managing StoreConnect media assets (images, videos, files).`,
}

func init() {
	rootCmd.AddCommand(mediaCmd)
	mediaCmd.AddCommand(mediaListCmd)
}

var mediaListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all media assets",
	Long:  `List all media assets from the connected StoreConnect server.`,
	RunE:  runMediaList,
}

func runMediaList(cmd *cobra.Command, args []string) error {
	formatter := ui.NewFormatter()

	client, serverAlias, err := getAPIClient(cmd)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to get API client: %v", err))
		return err
	}

	mediaService := api.NewMedia(client)

	spinner := ui.NewSpinner("Fetching media assets")
	spinner.Start()

	media, err := mediaService.List()
	if err != nil {
		spinner.Error(fmt.Sprintf("Failed to fetch media: %v", err))
		return err
	}

	spinner.Stop()

	if len(media) == 0 {
		formatter.Warning("No media assets found")
		return nil
	}

	formatter.Info(fmt.Sprintf("Media assets on %s:", serverAlias))
	fmt.Println()

	for _, asset := range media {
		fmt.Printf("  • %s (%s)", asset.Filename, asset.ContentType)
		if asset.Size > 0 {
			fmt.Printf(" - %s", formatBytes(asset.Size))
		}
		fmt.Println()
		formatter.Dim(fmt.Sprintf("    %s", asset.URL))
	}

	fmt.Println()
	formatter.Dim(fmt.Sprintf("Total: %d media assets", len(media)))

	return nil
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
