package commands

import (
	"fmt"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"github.com/GetStoreConnect/storeconnect-cli/internal/ui"
	"github.com/spf13/cobra"
)

var articleCmd = &cobra.Command{
	Use:   "article",
	Short: "Article management commands",
	Long:  `Commands for managing StoreConnect articles and blog posts.`,
}

func init() {
	rootCmd.AddCommand(articleCmd)
	articleCmd.AddCommand(articleListCmd)
}

var articleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all articles",
	Long:  `List all articles from the connected StoreConnect server.`,
	RunE:  runArticleList,
}

func runArticleList(cmd *cobra.Command, args []string) error {
	formatter := ui.NewFormatter()

	client, serverAlias, err := getAPIClient(cmd)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to get API client: %v", err))
		return err
	}

	articlesService := api.NewArticles(client)

	spinner := ui.NewSpinner("Fetching articles")
	spinner.Start()

	articles, err := articlesService.List()
	if err != nil {
		spinner.Error(fmt.Sprintf("Failed to fetch articles: %v", err))
		return err
	}

	spinner.Stop()

	if len(articles) == 0 {
		formatter.Warning("No articles found")
		return nil
	}

	formatter.Info(fmt.Sprintf("Articles on %s:", serverAlias))
	fmt.Println()

	for _, article := range articles {
		fmt.Printf("  • %s", article.Title)
		if article.Slug != "" {
			fmt.Printf(" (/%s)", article.Slug)
		}
		fmt.Println()
	}

	fmt.Println()
	formatter.Dim(fmt.Sprintf("Total: %d articles", len(articles)))

	return nil
}
