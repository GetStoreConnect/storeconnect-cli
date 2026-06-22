package commands

import (
	"fmt"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"github.com/GetStoreConnect/storeconnect-cli/internal/ui"
	"github.com/spf13/cobra"
)

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "Content block management commands",
	Long:  `Commands for managing StoreConnect content blocks.`,
}

func init() {
	rootCmd.AddCommand(blockCmd)
	blockCmd.AddCommand(blockListCmd)
}

var blockListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all content blocks",
	Long:  `List all content blocks from the connected StoreConnect server.`,
	RunE:  runBlockList,
}

func runBlockList(cmd *cobra.Command, args []string) error {
	formatter := ui.NewFormatter()

	client, serverAlias, err := getAPIClient(cmd)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to get API client: %v", err))
		return err
	}

	blocksService := api.NewContentBlocks(client)

	spinner := ui.NewSpinner("Fetching content blocks")
	spinner.Start()

	blocks, err := blocksService.List()
	if err != nil {
		spinner.Error(fmt.Sprintf("Failed to fetch content blocks: %v", err))
		return err
	}

	spinner.Stop()

	if len(blocks) == 0 {
		formatter.Warning("No content blocks found")
		return nil
	}

	formatter.Info(fmt.Sprintf("Content blocks on %s:", serverAlias))
	fmt.Println()

	for _, block := range blocks {
		fmt.Printf("  • %s", block.Name)
		if block.Template != "" {
			fmt.Printf(" (%s)", block.Template)
		}
		fmt.Println()
	}

	fmt.Println()
	formatter.Dim(fmt.Sprintf("Total: %d content blocks", len(blocks)))

	return nil
}
