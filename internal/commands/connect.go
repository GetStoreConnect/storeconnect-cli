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
  # Interactive mode (prompts for credentials)
  sc connect https://dev.mystore.com --alias dev

  # Non-interactive mode with flags
  sc connect https://dev.mystore.com --alias dev \
    --org-id 00D000000000062 \
    --store-id a0A7Z00000AbCdEFGH \
    --api-key your-api-key \
    --non-interactive

  # Using environment variables
  export SC_ORG_ID=00D000000000062
  export SC_STORE_ID=a0A7Z00000AbCdEFGH
  export SC_API_KEY=your-api-key
  sc connect https://dev.mystore.com --alias dev --non-interactive

SECURITY NOTE:
  • API keys are stored securely in ~/.storeconnect/credentials.yml (0600 permissions)
  • Project config contains NO secrets and is safe to commit to git
  • Each developer should have their own API key
  • Use environment variables or flags for CI/CD automation`,
	Args: cobra.ExactArgs(1),
	RunE: runConnect,
}

var (
	connectOrgID   string
	connectStoreID string
	connectAPIKey  string
	connectAlias   string
)

func init() {
	rootCmd.AddCommand(connectCmd)

	connectCmd.Flags().StringVar(&connectOrgID, "org-id", "", "Salesforce Organization ID (or set SC_ORG_ID env var)")
	connectCmd.Flags().StringVar(&connectStoreID, "store-id", "", "Store Salesforce ID (or set SC_STORE_ID env var)")
	connectCmd.Flags().StringVar(&connectAPIKey, "api-key", "", "API key (or set SC_API_KEY env var)")
	connectCmd.Flags().StringVar(&connectAlias, "alias", "", "Alias name for this server (e.g., dev, staging, prod)")
	connectCmd.MarkFlagRequired("alias")
}

func runConnect(cmd *cobra.Command, args []string) error {
	url := normalizeURL(args[0])
	formatter := ui.NewFormatter()

	// Get org ID from flag, env, or prompt
	var err error
	connectOrgID, err = getCredentialInput(
		connectOrgID,
		"SC_ORG_ID",
		"Organization ID (15 or 18 chars, starts with 00D)",
		"org-id",
	)
	if err != nil {
		if !jsonOutput {
			formatter.Error(err.Error())
		}
		return outputError(err)
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

	// Get store ID from flag, env, or prompt
	connectStoreID, err = getCredentialInput(
		connectStoreID,
		"SC_STORE_ID",
		"Store Salesforce ID (15 or 18 alphanumeric characters)",
		"store-id",
	)
	if err != nil {
		if !jsonOutput {
			formatter.Error(err.Error())
		}
		return outputError(err)
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

	// Get API key from flag, env, or secure prompt
	apiKey, err := getSecretInput(
		connectAPIKey,
		"SC_API_KEY",
		"API key (hidden)",
		"api-key",
	)
	if err != nil {
		if !jsonOutput {
			formatter.Error(err.Error())
		}
		return outputError(err)
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

