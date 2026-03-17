package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GetStoreConnect/storeconnect-cli/internal/ui"
	"github.com/GetStoreConnect/storeconnect-cli/internal/validators"
	"github.com/spf13/cobra"
)

func init() {
	themeCmd.AddCommand(themeValidateCmd)
}

var themeValidateCmd = &cobra.Command{
	Use:   "validate THEME_NAME",
	Short: "Validate theme structure and templates",
	Long: `Validate theme structure and Liquid template syntax.

Checks for:
  • Valid theme directory structure
  • Liquid template syntax errors
  • Unclosed tags and braces
  • Template file naming conventions`,
	Args: cobra.ExactArgs(1),
	RunE: runThemeValidate,
}

func runThemeValidate(cmd *cobra.Command, args []string) error {
	themeName := args[0]
	formatter := ui.NewFormatter()

	themePath := filepath.Join("themes", themeName)

	// Check if theme exists
	if _, err := os.Stat(themePath); os.IsNotExist(err) {
		formatter.Error(fmt.Sprintf("Theme '%s' not found", themeName))
		return err
	}

	spinner := ui.NewSpinner(fmt.Sprintf("Validating theme '%s'", themeName))
	spinner.Start()

	validator := validators.NewLiquidValidator()
	hasErrors := false
	errorCount := 0

	// Validate all .liquid files
	templatesDir := filepath.Join(themePath, "templates")
	if _, err := os.Stat(templatesDir); err == nil {
		err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && strings.HasSuffix(path, ".liquid") {
				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				errors := validator.Validate(string(content))
				if len(errors) > 0 {
					hasErrors = true
					errorCount += len(errors)

					relPath, _ := filepath.Rel(themePath, path)

					if spinner != nil {
						spinner.Stop()
						spinner = nil
					}

					formatter.Error(relPath)
					for _, errMsg := range errors {
						fmt.Printf("  - %s\n", errMsg)
					}
					fmt.Println()
				}
			}

			return nil
		})

		if err != nil {
			if spinner != nil {
				spinner.Error(fmt.Sprintf("Failed to validate theme: %v", err))
			}
			return err
		}
	}

	if spinner != nil {
		if hasErrors {
			spinner.Error(fmt.Sprintf("Found %d validation errors", errorCount))
		} else {
			spinner.Success(fmt.Sprintf("Theme '%s' is valid", themeName))
		}
	} else {
		if !hasErrors {
			formatter.Success(fmt.Sprintf("Theme '%s' is valid", themeName))
		}
	}

	if hasErrors {
		return fmt.Errorf("validation failed")
	}

	return nil
}
