package commands

import (
	"fmt"
	"strings"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"github.com/GetStoreConnect/storeconnect-cli/internal/ui"
	"github.com/spf13/cobra"
)

var menuCmd = &cobra.Command{
	Use:   "menu",
	Short: "Menu management commands",
	Long:  `Commands for managing StoreConnect navigation menus.`,
}

func init() {
	rootCmd.AddCommand(menuCmd)
	menuCmd.AddCommand(menuListCmd)
	menuCmd.AddCommand(menuShowCmd)
}

var menuListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all menus",
	Long:  `List all menus from the connected StoreConnect server.`,
	RunE:  runMenuList,
}

var menuShowCmd = &cobra.Command{
	Use:   "show MENU_ID",
	Short: "Show menu structure",
	Long:  `Show the structure of a menu including all items.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runMenuShow,
}

func runMenuList(cmd *cobra.Command, args []string) error {
	formatter := ui.NewFormatter()

	client, serverAlias, err := getAPIClient(cmd)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to get API client: %v", err))
		return err
	}

	menusService := api.NewMenus(client)

	spinner := ui.NewSpinner("Fetching menus")
	spinner.Start()

	menus, err := menusService.List()
	if err != nil {
		spinner.Error(fmt.Sprintf("Failed to fetch menus: %v", err))
		return err
	}

	spinner.Stop()

	if len(menus) == 0 {
		formatter.Warning("No menus found")
		return nil
	}

	formatter.Info(fmt.Sprintf("Menus on %s:", serverAlias))
	fmt.Println()

	for _, menu := range menus {
		fmt.Printf("  • %s (%d items)\n", menu.Name, len(menu.Items))
	}

	fmt.Println()
	formatter.Dim(fmt.Sprintf("Total: %d menus", len(menus)))

	return nil
}

func runMenuShow(cmd *cobra.Command, args []string) error {
	menuID := args[0]
	formatter := ui.NewFormatter()

	client, _, err := getAPIClient(cmd)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to get API client: %v", err))
		return err
	}

	menusService := api.NewMenus(client)

	spinner := ui.NewSpinner("Fetching menu")
	spinner.Start()

	menu, err := menusService.Get(menuID)
	if err != nil {
		spinner.Error(fmt.Sprintf("Failed to fetch menu: %v", err))
		return err
	}

	spinner.Stop()

	formatter.Info(fmt.Sprintf("Menu: %s", menu.Name))
	fmt.Println()

	printMenuItems(menu.Items, 0)

	return nil
}

func printMenuItems(items []api.MenuItem, depth int) {
	indent := strings.Repeat("  ", depth)
	for _, item := range items {
		fmt.Printf("%s├─ %s → %s\n", indent, item.Label, item.URL)
		if len(item.Children) > 0 {
			printMenuItems(item.Children, depth+1)
		}
	}
}
