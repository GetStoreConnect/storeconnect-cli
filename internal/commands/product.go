package commands

import (
	"fmt"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"github.com/GetStoreConnect/storeconnect-cli/internal/config"
	"github.com/GetStoreConnect/storeconnect-cli/internal/ui"
	"github.com/spf13/cobra"
)

var productCmd = &cobra.Command{
	Use:   "product",
	Short: "Product management commands",
	Long:  `Commands for managing StoreConnect products.`,
}

func init() {
	rootCmd.AddCommand(productCmd)
	productCmd.AddCommand(productListCmd)
}

var productListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all products",
	Long:  `List all products from the connected StoreConnect server.`,
	RunE:  runProductList,
}

func runProductList(cmd *cobra.Command, args []string) error {
	formatter := ui.NewFormatter()

	// Get server and credentials
	client, serverAlias, err := getAPIClient(cmd)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to get API client: %v", err))
		return err
	}

	productsService := api.NewProducts(client)

	// Fetch products
	spinner := ui.NewSpinner("Fetching products")
	spinner.Start()

	products, err := productsService.List()
	if err != nil {
		spinner.Error(fmt.Sprintf("Failed to fetch products: %v", err))
		return err
	}

	spinner.Stop()

	if len(products) == 0 {
		formatter.Warning("No products found")
		return nil
	}

	// Display products
	formatter.Info(fmt.Sprintf("Products on %s:", serverAlias))
	fmt.Println()

	for _, product := range products {
		fmt.Printf("  • %s", product.Name)
		if product.SKU != "" {
			fmt.Printf(" (SKU: %s)", product.SKU)
		}
		fmt.Println()
		if product.Price > 0 {
			formatter.Dim(fmt.Sprintf("    Price: $%.2f", product.Price))
		}
	}

	fmt.Println()
	formatter.Dim(fmt.Sprintf("Total: %d products", len(products)))

	return nil
}

// Helper function to get API client with server selection
func getAPIClient(cmd *cobra.Command) (*api.Client, string, error) {
	// Get server alias
	serverAlias, _ := cmd.Flags().GetString("server")
	if serverAlias == "" {
		project, err := config.NewProject()
		if err != nil {
			return nil, "", err
		}
		serverAlias = project.GetDefaultServer()
	}

	if serverAlias == "" {
		return nil, "", fmt.Errorf("no server configured")
	}

	// Load credentials
	credentials, err := config.NewCredentials()
	if err != nil {
		return nil, "", err
	}

	cred, ok := credentials.GetServer(serverAlias)
	if !ok {
		return nil, "", fmt.Errorf("server '%s' not found in credentials", serverAlias)
	}

	// Create API client
	client := api.NewClient(cred.URL, cred.StoreSFID, cred.APIKey, api.WithOrgID(cred.OrgID))

	return client, serverAlias, nil
}
