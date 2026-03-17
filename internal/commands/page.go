package commands

import (
	"fmt"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"github.com/GetStoreConnect/storeconnect-cli/internal/ui"
	"github.com/spf13/cobra"
)

var pageCmd = &cobra.Command{
	Use:   "page",
	Short: "Page management commands",
	Long:  `Commands for managing StoreConnect pages.`,
}

func init() {
	rootCmd.AddCommand(pageCmd)
	pageCmd.AddCommand(pageListCmd)
}

var pageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all pages",
	Long:  `List all pages from the connected StoreConnect server.`,
	RunE:  runPageList,
}

func runPageList(cmd *cobra.Command, args []string) error {
	formatter := ui.NewFormatter()

	client, serverAlias, err := getAPIClient(cmd)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to get API client: %v", err))
		return err
	}

	pagesService := api.NewPages(client)

	spinner := ui.NewSpinner("Fetching pages")
	spinner.Start()

	pages, err := pagesService.List()
	if err != nil {
		spinner.Error(fmt.Sprintf("Failed to fetch pages: %v", err))
		return err
	}

	spinner.Stop()

	if len(pages) == 0 {
		formatter.Warning("No pages found")
		return nil
	}

	formatter.Info(fmt.Sprintf("Pages on %s:", serverAlias))
	fmt.Println()

	for _, page := range pages {
		fmt.Printf("  • %s", page.Title)
		if page.Slug != "" {
			fmt.Printf(" (/%s)", page.Slug)
		}
		fmt.Println()
	}

	fmt.Println()
	formatter.Dim(fmt.Sprintf("Total: %d pages", len(pages)))

	return nil
}
