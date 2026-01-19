package main

import (
	"os"

	"github.com/lgbarn/pdf-cli/internal/cli"
	_ "github.com/lgbarn/pdf-cli/internal/commands" // Register all commands
)

// Version information set by build flags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.SetVersion(version, commit, date)
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
