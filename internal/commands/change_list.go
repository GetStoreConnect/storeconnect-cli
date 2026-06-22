package commands

import (
	"fmt"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"github.com/GetStoreConnect/storeconnect-cli/internal/ui"
	"github.com/spf13/cobra"
)

var changeCmd = &cobra.Command{
	Use:   "change",
	Short: "Content change management commands",
	Long:  `Commands for working with StoreConnect content changes (drafts) so they can be resumed from any machine.`,
}

var changeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List content changes",
	Long: `List the store's content changes (drafts).

Use --status to filter by status (e.g. draft, review, published). Drafts
created on one machine can be discovered and resumed from another.`,
	RunE: runChangeList,
}

func init() {
	rootCmd.AddCommand(changeCmd)
	changeCmd.AddCommand(changeListCmd)

	changeListCmd.Flags().String("status", "", "filter content changes by status (e.g. draft, review, published)")
}

func runChangeList(cmd *cobra.Command, args []string) error {
	formatter := ui.NewFormatter()
	status, _ := cmd.Flags().GetString("status")

	client, serverAlias, err := getAPIClient(cmd)
	if err != nil {
		if !jsonOutput {
			formatter.Error(fmt.Sprintf("Failed to get API client: %v", err))
		}
		return outputError(err)
	}

	changesService := api.NewContentChanges(client)

	var spinner *ui.Spinner
	if !jsonOutput {
		spinner = ui.NewSpinner("Fetching content changes")
		spinner.Start()
	}

	changes, err := changesService.List(status)
	if err != nil {
		if spinner != nil {
			spinner.Error(fmt.Sprintf("Failed to fetch content changes: %v", err))
		}
		return outputError(err)
	}

	if spinner != nil {
		spinner.Stop()
	}

	// JSON output
	if jsonOutput {
		return outputResponse(changes, nil)
	}

	// Human-friendly table
	if len(changes) == 0 {
		formatter.Warning("No content changes found")
		return nil
	}

	formatter.Info(fmt.Sprintf("Content changes on %s:", serverAlias))
	fmt.Println()

	fmt.Printf("%-22s %-12s %-8s %s\n", "SC ID", "STATUS", "RECORDS", "SUMMARY")
	for _, change := range changes {
		fmt.Printf("%-22s %-12s %-8d %s\n", change.SCID, change.Status, change.RecordsCount, change.Summary)
	}

	fmt.Println()
	formatter.Dim(fmt.Sprintf("Total: %d content changes", len(changes)))

	return nil
}
