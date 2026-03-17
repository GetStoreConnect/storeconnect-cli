package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"github.com/GetStoreConnect/storeconnect-cli/internal/config"
	"github.com/GetStoreConnect/storeconnect-cli/internal/ui"
	"github.com/GetStoreConnect/storeconnect-cli/internal/utils"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var connectCmd = &cobra.Command{
	Use:   "connect URL",
	Short: "Connect to StoreConnect server",
	Long: `Connect to a StoreConnect server for theme development.

You'll be prompted interactively for:
  • Organization ID (15-18 chars starting with "00D")
  • Store Salesforce ID (15-18 chars starting with "a0A")
  • API key (hidden input for security)

FINDING YOUR ORGANIZATION ID:
  1. Log into Salesforce
  2. Go to Setup → Company Information
  3. Copy the "Organization ID" field

FINDING YOUR STORE SFID:
  1. Navigate to your Store record in Salesforce
  2. Copy the 15-18 character ID from the URL
  3. Example: a0A7Z00000AbCdEFGH

GENERATING AN API KEY:
  1. Open your Store record in Salesforce
  2. Find "API Keys" related list
  3. Click "New API Key"
  4. Copy the generated key (you can't view it again)

EXAMPLES:
  # Provide store ID interactively (recommended)
  sc connect https://dev.mystore.com --alias dev

  # Or provide via flag
  sc connect https://dev.mystore.com --store-id a0A... --alias dev

  # With Organization ID pre-filled
  sc connect https://staging.mystore.com --store-id a0A... --org-id 00D000000000062 --alias staging

After connecting, run 'sc theme refresh' to download themes.

SECURITY NOTE:
  • API keys are stored securely in ~/.storeconnect/credentials.yml (0600 permissions)
  • Project config contains NO secrets and is safe to commit to git
  • Each developer should have their own API key`,
	Args: cobra.ExactArgs(1),
	RunE: runConnect,
}

var (
	connectOrgID   string
	connectStoreID string
	connectAlias   string
)

func init() {
	rootCmd.AddCommand(connectCmd)

	connectCmd.Flags().StringVar(&connectOrgID, "org-id", "", "Salesforce Organization ID (will prompt if not provided)")
	connectCmd.Flags().StringVar(&connectStoreID, "store-id", "", "Store Salesforce ID (will prompt if not provided)")
	connectCmd.Flags().StringVar(&connectAlias, "alias", "", "Alias name for this server (e.g., dev, staging, prod)")
	connectCmd.MarkFlagRequired("alias")
}

func runConnect(cmd *cobra.Command, args []string) error {
	url := normalizeURL(args[0])
	formatter := ui.NewFormatter()

	// Prompt for org ID if not provided
	if connectOrgID == "" {
		orgID, err := promptForOrgID()
		if err != nil {
			formatter.Error(fmt.Sprintf("Failed to read Organization ID: %v", err))
			return err
		}
		connectOrgID = orgID
	}

	// Validate org ID
	if !utils.ValidateOrgID(connectOrgID) {
		formatter.Error("Invalid Organization ID format. Must be 15 or 18 characters starting with '00D'")
		return fmt.Errorf("invalid org ID")
	}

	// Normalize org ID to 18 characters
	normalizedOrgID, err := utils.NormalizeSalesforceID(connectOrgID)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to normalize Organization ID: %v", err))
		return err
	}
	connectOrgID = normalizedOrgID

	// Prompt for store ID if not provided
	if connectStoreID == "" {
		storeID, err := promptForStoreID()
		if err != nil {
			formatter.Error(fmt.Sprintf("Failed to read Store ID: %v", err))
			return err
		}
		connectStoreID = storeID
	}

	// Validate store ID
	if !utils.ValidateSalesforceID(connectStoreID) {
		formatter.Error("Invalid Store ID format. Must be 15 or 18 alphanumeric characters")
		return fmt.Errorf("invalid store ID")
	}

	// Normalize store ID to 18 characters
	normalizedStoreID, err := utils.NormalizeSalesforceID(connectStoreID)
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to normalize Store ID: %v", err))
		return err
	}
	connectStoreID = normalizedStoreID

	// Always prompt for API key interactively (prevents shell history exposure)
	apiKey, err := promptForAPIKey()
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to read API key: %v", err))
		return err
	}

	// Load credentials
	credentials, err := config.NewCredentials()
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to load credentials: %v", err))
		return err
	}

	// Load project config
	project, err := config.NewProject()
	if err != nil {
		formatter.Error(fmt.Sprintf("Failed to load project config: %v", err))
		return err
	}

	// Test authentication
	spinner := ui.NewSpinner("Authenticating with server")
	spinner.Start()

	client := api.NewClient(url, connectStoreID, apiKey, api.WithOrgID(connectOrgID))
	authService := api.NewAuth(client)
	authInfo, err := authService.Info()

	if err != nil {
		spinner.Error(fmt.Sprintf("Authentication failed: %v", err))
		return err
	}

	spinner.Success("Authentication successful")

	// Save credentials globally
	if err := credentials.AddServer(connectAlias, url, connectStoreID, apiKey, connectOrgID); err != nil {
		formatter.Error(fmt.Sprintf("Failed to save credentials: %v", err))
		return err
	}
	formatter.Success(fmt.Sprintf("Saved credentials to %s", credentials.Path()))

	// Add server to project config if in a project
	if project.Exists() {
		setAsDefault := project.GetDefaultServer() == ""
		if err := project.AddServer(connectAlias, url, setAsDefault); err != nil {
			formatter.Error(fmt.Sprintf("Failed to update project config: %v", err))
			return err
		}

		// Update version information
		if authInfo.StoreconnectVersion != "" {
			project.UpdateServerInfo(connectAlias, authInfo.StoreconnectVersion, authInfo.BaseThemeVersion)
		}

		formatter.Success(fmt.Sprintf("Added '%s' to project config", connectAlias))
	}

	// Extract domain from URL
	domain := strings.TrimPrefix(url, "https://")
	domain = strings.TrimPrefix(domain, "http://")

	formatter.Success(fmt.Sprintf("Connected to %s", domain))

	if authInfo.StoreconnectVersion != "" {
		formatter.Success(fmt.Sprintf("StoreConnect version: %s", authInfo.StoreconnectVersion))
	}

	if project.Exists() && project.GetDefaultServer() == connectAlias {
		formatter.Success(fmt.Sprintf("Set '%s' as default server", connectAlias))
	}

	formatter.Newline()
	formatter.Dim("Run 'sc theme refresh' to download themes.")

	return nil
}

func normalizeURL(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	return strings.TrimSuffix(url, "/")
}

func promptForOrgID() (string, error) {
	fmt.Print("Enter Organization ID (15 or 18 chars, starts with 00D): ")
	reader := bufio.NewReader(os.Stdin)
	orgID, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	fmt.Println()
	return strings.TrimSpace(orgID), nil
}

func promptForStoreID() (string, error) {
	fmt.Print("Enter Store Salesforce ID (15 or 18 alphanumeric characters): ")
	reader := bufio.NewReader(os.Stdin)
	storeID, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	fmt.Println()
	return strings.TrimSpace(storeID), nil
}

func promptForAPIKey() (string, error) {
	fmt.Print("Enter API key (hidden): ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println()
	fmt.Println()
	return strings.TrimSpace(string(bytePassword)), nil
}
