package main

import (
	"fmt"
	"os"

	"github.com/heimdall-cli/heimdall/internal/commands"
)

// Version information set via ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
