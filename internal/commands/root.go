package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "sc",
		Short: "StoreConnect CLI - Build and manage StoreConnect themes",
		Long: `StoreConnect CLI is a command-line interface for developing and managing
StoreConnect themes. It allows developers to:
  • Edit theme files locally with version control
  • Work with multiple servers (dev, staging, production)
  • Push/pull themes between local machine and StoreConnect servers
  • Preview and validate themes before publishing to Salesforce`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
)

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .storeconnect/config.yaml)")
	rootCmd.PersistentFlags().StringP("server", "s", "", "server alias to use (e.g., dev, staging, prod)")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format (machine-readable)")

	// Add version command
	rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Look for config in .storeconnect directory
		viper.AddConfigPath(".storeconnect")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	// Read config if it exists (don't error if it doesn't)
	_ = viper.ReadInConfig()
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Show CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("StoreConnect CLI version %s\n", Version)
	},
}
