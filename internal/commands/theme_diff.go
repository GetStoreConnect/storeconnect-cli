package commands

import (
	"fmt"
	"os"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"github.com/GetStoreConnect/storeconnect-cli/internal/config"
	"github.com/GetStoreConnect/storeconnect-cli/internal/theme"
	"github.com/GetStoreConnect/storeconnect-cli/internal/ui"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	themeCmd.AddCommand(themeDiffCmd)
	themeDiffCmd.Flags().Bool("exit-code", false, "exit with code 9 when the local theme differs from the server (for CI)")
}

var themeDiffCmd = &cobra.Command{
	Use:   "diff THEME_NAME",
	Short: "Show how the local theme differs from the server",
	Long: `Compare a local theme against the server's copy without pushing.

Shows which templates and asset binaries have been added, modified, or only
exist on the server. This is the read-only sibling of 'sc theme push' — it
reports exactly what a push would send.

With --exit-code the command exits 9 when there are any differences, so CI can
gate on a theme being in sync with what's deployed:

    sc theme diff my-theme --exit-code || echo "theme has undeployed changes"`,
	Args: cobra.ExactArgs(1),
	RunE: runThemeDiff,
}

func runThemeDiff(cmd *cobra.Command, args []string) error {
	themeName := args[0]
	formatter := ui.NewFormatter()
	exitCode, _ := cmd.Flags().GetBool("exit-code")

	client, serverAlias, err := getAPIClient(cmd)
	if err != nil {
		if !jsonOutput {
			formatter.Error(fmt.Sprintf("Failed to get API client: %v", err))
		}
		return outputError(err)
	}

	var spinner *ui.Spinner
	if !jsonOutput {
		spinner = ui.NewSpinner(fmt.Sprintf("Comparing theme '%s'", themeName))
		spinner.Start()
	}

	// Read the local theme (templates) and its asset binaries.
	deserializer := theme.NewDeserializer(".")
	localTheme, err := deserializer.Deserialize(themeName)
	if err != nil {
		if spinner != nil {
			spinner.Error(fmt.Sprintf("Failed to read theme: %v", err))
		}
		return outputError(err)
	}

	themeDir := fmt.Sprintf("themes/%s", themeName)
	localAssets, err := theme.ReadLocalAssets(themeDir)
	if err != nil {
		if spinner != nil {
			spinner.Error(fmt.Sprintf("Failed to read assets: %v", err))
		}
		return outputError(err)
	}

	// Resolve the server-side theme id the same way push does.
	themeID := localTheme.SCID
	if themeID == "" {
		if syncState, sErr := config.NewSyncState(themeDir); sErr == nil {
			themeID = syncState.ThemeSCID
		}
	}
	if themeID == "" {
		if spinner != nil {
			spinner.Error("Theme has no SC ID. Pull the theme first or create it on the server.")
		}
		return fmt.Errorf("theme has no sc_id")
	}

	themesService := api.NewThemes(client)
	serverTheme, err := themesService.Get(themeID)
	if err != nil {
		if spinner != nil {
			spinner.Error(fmt.Sprintf("Failed to fetch server theme: %v", err))
		}
		return outputError(err)
	}

	if spinner != nil {
		spinner.Stop()
	}

	diff := theme.ComputeDiff(localTheme, localAssets, serverTheme)

	if jsonOutput {
		if err := outputResponse(ThemeDiffResponse{
			ThemeName:  themeName,
			Templates:  diff.Templates,
			Assets:     diff.Assets,
			HasChanges: diff.HasChanges(),
			Pushable:   diff.PushableCount(),
			ServerOnly: diff.ServerOnlyCount(),
		}, nil); err != nil {
			return err
		}
	} else {
		renderDiff(formatter, themeName, serverAlias, diff)
	}

	if exitCode && diff.HasChanges() {
		os.Exit(ExitDiffChanges)
	}

	return nil
}

func renderDiff(formatter *ui.Formatter, themeName, serverAlias string, diff theme.Diff) {
	if !diff.HasChanges() {
		formatter.Success(fmt.Sprintf("Theme '%s' is in sync with %s", themeName, serverAlias))
		return
	}

	formatter.Info(fmt.Sprintf("Diff for theme '%s' against %s:", themeName, serverAlias))

	renderCategory(formatter, "Templates", diff.Templates)
	renderCategory(formatter, "Assets", diff.Assets)

	formatter.Newline()
	formatter.Dim(fmt.Sprintf("%d to push, %d only on server", diff.PushableCount(), diff.ServerOnlyCount()))
}

func renderCategory(formatter *ui.Formatter, label string, c theme.CategoryDiff) {
	if len(c.Added)+len(c.Modified)+len(c.Removed) == 0 {
		return
	}

	formatter.Newline()
	formatter.Print(label + ":")

	added := color.New(color.FgGreen).SprintFunc()
	modified := color.New(color.FgYellow).SprintFunc()
	removed := color.New(color.FgRed).SprintFunc()

	for _, key := range c.Added {
		fmt.Printf("  %s %s\n", added("+"), key)
	}
	for _, key := range c.Modified {
		fmt.Printf("  %s %s\n", modified("~"), key)
	}
	for _, key := range c.Removed {
		fmt.Printf("  %s %s\n", removed("-"), key)
	}
}
