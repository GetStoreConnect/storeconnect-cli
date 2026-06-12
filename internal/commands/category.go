package commands

import (
	"fmt"
	"strings"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"github.com/GetStoreConnect/storeconnect-cli/internal/ui"
	"github.com/spf13/cobra"
)

var categoryCmd = &cobra.Command{
	Use:   "category",
	Short: "Category management commands",
	Long:  `Commands for managing StoreConnect product categories.`,
}

func init() {
	rootCmd.AddCommand(categoryCmd)
	categoryCmd.AddCommand(categoryListCmd)
	categoryCmd.AddCommand(categoryTreeCmd)
}

var categoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all categories",
	Long:  `List all categories from the connected StoreConnect server.`,
	RunE:  runCategoryList,
}

var categoryTreeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Show category hierarchy",
	Long:  `Show category hierarchy as a tree.`,
	RunE:  runCategoryTree,
}

func runCategoryList(cmd *cobra.Command, args []string) error {
	formatter := ui.NewFormatter()

	client, serverAlias, err := getAPIClient(cmd)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to get API client: %v", err))
		return err
	}

	categoriesService := api.NewCategories(client)

	spinner := ui.NewSpinner("Fetching categories")
	spinner.Start()

	categories, err := categoriesService.List()
	if err != nil {
		spinner.Error(fmt.Sprintf("Failed to fetch categories: %v", err))
		return err
	}

	spinner.Stop()

	if len(categories) == 0 {
		formatter.Warning("No categories found")
		return nil
	}

	formatter.Info(fmt.Sprintf("Categories on %s:", serverAlias))
	fmt.Println()

	for _, category := range categories {
		fmt.Printf("  • %s", category.Name)
		if category.Path != "" {
			fmt.Printf(" (%s)", category.Path)
		}
		fmt.Println()
	}

	fmt.Println()
	formatter.Dim(fmt.Sprintf("Total: %d categories", len(categories)))

	return nil
}

func runCategoryTree(cmd *cobra.Command, args []string) error {
	formatter := ui.NewFormatter()

	client, serverAlias, err := getAPIClient(cmd)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to get API client: %v", err))
		return err
	}

	categoriesService := api.NewCategories(client)

	spinner := ui.NewSpinner("Fetching category tree")
	spinner.Start()

	categories, err := categoriesService.Tree()
	if err != nil {
		spinner.Error(fmt.Sprintf("Failed to fetch categories: %v", err))
		return err
	}

	spinner.Stop()

	if len(categories) == 0 {
		formatter.Warning("No categories found")
		return nil
	}

	formatter.Info(fmt.Sprintf("Category tree on %s:", serverAlias))
	fmt.Println()

	printCategoryTree(categories, 0)

	return nil
}

func printCategoryTree(categories []api.Category, depth int) {
	indent := strings.Repeat("  ", depth)
	for _, category := range categories {
		fmt.Printf("%s├─ %s\n", indent, category.Name)
		if len(category.Children) > 0 {
			printCategoryTree(category.Children, depth+1)
		}
	}
}
