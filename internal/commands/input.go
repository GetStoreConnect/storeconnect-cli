package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// getCredentialInput gets input from flag, environment variable, or interactive prompt
// Priority: 1. Flag value, 2. Environment variable, 3. Prompt (if interactive)
func getCredentialInput(flagValue, envVar, promptText, flagName string) (string, error) {
	// 1. Check flag value
	if flagValue != "" {
		return flagValue, nil
	}

	// 2. Check environment variable
	if envVar != "" {
		if envValue := os.Getenv(envVar); envValue != "" {
			return envValue, nil
		}
	}

	// 3. Prompt if interactive mode
	if nonInteractive {
		errMsg := fmt.Sprintf("%s required in non-interactive mode", promptText)
		if flagName != "" {
			errMsg += fmt.Sprintf(" (use --%s flag", flagName)
			if envVar != "" {
				errMsg += fmt.Sprintf(" or %s environment variable", envVar)
			}
			errMsg += ")"
		}
		return "", fmt.Errorf("%s", errMsg)
	}

	// Interactive prompt
	return promptForInput(promptText), nil
}

// getSecretInput gets secret input (password/API key) from flag, env, or secure prompt
// Similar to getCredentialInput but uses hidden input for prompts
func getSecretInput(flagValue, envVar, promptText, flagName string) (string, error) {
	// 1. Check flag value
	if flagValue != "" {
		return flagValue, nil
	}

	// 2. Check environment variable
	if envVar != "" {
		if envValue := os.Getenv(envVar); envValue != "" {
			return envValue, nil
		}
	}

	// 3. Prompt if interactive mode
	if nonInteractive {
		errMsg := fmt.Sprintf("%s required in non-interactive mode", promptText)
		if flagName != "" {
			errMsg += fmt.Sprintf(" (use --%s flag", flagName)
			if envVar != "" {
				errMsg += fmt.Sprintf(" or %s environment variable", envVar)
			}
			errMsg += ")"
		}
		return "", fmt.Errorf("%s", errMsg)
	}

	// Interactive secure prompt (hidden input)
	return promptForSecret(promptText)
}

// promptForInput prompts the user for input (visible)
func promptForInput(promptText string) string {
	fmt.Printf("%s: ", promptText)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// promptForSecret prompts the user for secret input (hidden)
func promptForSecret(promptText string) (string, error) {
	fmt.Printf("%s: ", promptText)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // New line after hidden input
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	return strings.TrimSpace(string(bytePassword)), nil
}

// confirmAction asks for user confirmation (yes/no)
// In non-interactive mode with --yes flag, always returns true
// In non-interactive mode without --yes flag, returns error
func confirmAction(promptText string, yesFlag bool) (bool, error) {
	if yesFlag {
		return true, nil
	}

	if nonInteractive {
		return false, fmt.Errorf("confirmation required in non-interactive mode (use --yes flag)")
	}

	fmt.Printf("%s (y/n): ", promptText)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))

	return response == "y" || response == "yes", nil
}
