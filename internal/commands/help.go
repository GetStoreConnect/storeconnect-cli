package commands

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommandHelp represents structured help information for a command
type CommandHelp struct {
	Name        string            `json:"name"`
	Usage       string            `json:"usage"`
	Short       string            `json:"short"`
	Long        string            `json:"long"`
	Flags       []FlagHelp        `json:"flags,omitempty"`
	Subcommands []string          `json:"subcommands,omitempty"`
	Examples    []string          `json:"examples,omitempty"`
	ExitCodes   map[string]string `json:"exit_codes,omitempty"`
}

// FlagHelp represents help information for a single flag
type FlagHelp struct {
	Name      string `json:"name"`
	Shorthand string `json:"shorthand,omitempty"`
	Type      string `json:"type"`
	Default   string `json:"default,omitempty"`
	Usage     string `json:"usage"`
}

// GetCommandHelp extracts help information from a Cobra command
func GetCommandHelp(cmd *cobra.Command) CommandHelp {
	help := CommandHelp{
		Name:   cmd.Name(),
		Usage:  cmd.UseLine(),
		Short:  cmd.Short,
		Long:   cmd.Long,
		Flags:  []FlagHelp{},
	}

	// Extract flags
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		help.Flags = append(help.Flags, FlagHelp{
			Name:      flag.Name,
			Shorthand: flag.Shorthand,
			Type:      flag.Value.Type(),
			Default:   flag.DefValue,
			Usage:     flag.Usage,
		})
	})

	// Extract persistent flags from parent
	if cmd.HasParent() {
		cmd.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
			help.Flags = append(help.Flags, FlagHelp{
				Name:      flag.Name,
				Shorthand: flag.Shorthand,
				Type:      flag.Value.Type(),
				Default:   flag.DefValue,
				Usage:     flag.Usage + " (global)",
			})
		})
	}

	// Extract subcommands
	for _, subCmd := range cmd.Commands() {
		if !subCmd.Hidden {
			help.Subcommands = append(help.Subcommands, subCmd.Name())
		}
	}

	// Add exit codes for relevant commands
	help.ExitCodes = map[string]string{
		"0": "Success",
		"1": "Generic error",
		"2": "Authentication failed",
		"3": "Resource not found",
		"4": "Validation error",
		"5": "Network error",
		"6": "Resource conflict",
		"7": "User cancelled",
		"8": "Configuration error",
	}

	return help
}

// OutputJSONHelp outputs help information in JSON format
func OutputJSONHelp(cmd *cobra.Command) error {
	help := GetCommandHelp(cmd)
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(help)
}

// helpCommand is a custom help command that supports JSON output
var helpCommand = &cobra.Command{
	Use:   "help [command]",
	Short: "Help about any command",
	Long: `Help provides help for any command in the application.
Simply type sc help [path to command] for full details.

Use --json flag for machine-readable help output.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find the command
		targetCmd, _, err := rootCmd.Find(args)
		if err != nil {
			return err
		}

		// Output JSON help if requested
		if jsonOutput {
			return OutputJSONHelp(targetCmd)
		}

		// Default help output
		return targetCmd.Help()
	},
}

func init() {
	rootCmd.SetHelpCommand(helpCommand)
}

// Example of adding structured examples to a command:
// cmd.Example = `  # Interactive mode
//   sc connect https://dev.mystore.com --alias dev
//
//   # Non-interactive with environment variables
//   export SC_ORG_ID=00D...
//   export SC_API_KEY=...
//   sc connect https://dev.mystore.com --alias dev --non-interactive`
