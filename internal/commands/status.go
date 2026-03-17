package commands

import (
	"fmt"

	"github.com/GetStoreConnect/storeconnect-cli/internal/config"
	"github.com/GetStoreConnect/storeconnect-cli/internal/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show connection status and project info",
	Long:  `Display information about configured servers and project settings.`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	formatter := ui.NewFormatter()

	// Load credentials
	credentials, err := config.NewCredentials()
	if err != nil {
		if !jsonOutput {
			formatter.Error(fmt.Sprintf("Failed to load credentials: %v", err))
		}
		return outputError(err)
	}

	// Load project config
	project, err := config.NewProject()
	if err != nil {
		if !jsonOutput {
			formatter.Error(fmt.Sprintf("Failed to load project config: %v", err))
		}
		return outputError(err)
	}

	// Check if in a project
	if !project.Exists() {
		// JSON output for not in project
		if jsonOutput {
			return outputResponse(StatusResponse{
				Connected: false,
				Servers:   make(map[string]ServerStatus),
			}, nil)
		}

		formatter.Warning("Not in a StoreConnect project")
		formatter.Newline()
		formatter.Dim("Run 'sc init <project-name>' to create a new project")
		return nil
	}

	// Build status response
	servers := make(map[string]ServerStatus)
	defaultServer := project.GetDefaultServer()
	connected := len(project.Servers) > 0

	for alias, server := range project.Servers {
		// Check if credentials exist
		_, hasCredentials := credentials.GetServer(alias)

		servers[alias] = ServerStatus{
			URL:           server.URL,
			Version:       server.StoreconnectVersion,
			Authenticated: hasCredentials,
		}
	}

	// JSON output
	if jsonOutput {
		return outputResponse(StatusResponse{
			Connected:     connected,
			Servers:       servers,
			DefaultServer: defaultServer,
			ProjectDir:    ".storeconnect",
		}, nil)
	}

	// Human-friendly output
	formatter.Info("Project Configuration")
	fmt.Println()

	if len(project.Servers) == 0 {
		formatter.Warning("No servers configured")
		formatter.Newline()
		formatter.Dim("Run 'sc connect <url> --alias <name>' to add a server")
		return nil
	}

	// Display servers
	for alias, server := range project.Servers {
		marker := " "
		if alias == defaultServer {
			marker = "*"
		}

		fmt.Printf("%s %s\n", marker, alias)
		fmt.Printf("  URL: %s\n", server.URL)

		if server.StoreconnectVersion != "" {
			fmt.Printf("  Version: %s\n", server.StoreconnectVersion)
		}

		if server.BaseThemeVersion != "" {
			fmt.Printf("  Base Theme: %s\n", server.BaseThemeVersion)
		}

		// Check if credentials exist
		if _, ok := credentials.GetServer(alias); !ok {
			formatter.Warning(fmt.Sprintf("  Missing credentials in %s", credentials.Path()))
		}

		fmt.Println()
	}

	formatter.Dim(fmt.Sprintf("* = default server"))
	formatter.Newline()
	formatter.Dim("Credentials stored in: " + credentials.Path())

	return nil
}
