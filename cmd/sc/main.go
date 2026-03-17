package main

import (
	"github.com/GetStoreConnect/storeconnect-cli/internal/commands"
)

func main() {
	// Execute will handle its own exit codes via outputError
	_ = commands.Execute()
}
