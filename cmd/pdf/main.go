package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/lgbarn/pdf-cli/internal/cleanup"
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
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer func() { _ = cleanup.Run() }()

	cli.SetVersion(version, commit, date)
	if err := cli.ExecuteContext(ctx); err != nil {
		return 1
	}
	return 0
}
