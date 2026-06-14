package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/fx"

	"github.com/uptime-com/uptime-mcp/internal/app"
	"github.com/uptime-com/uptime-mcp/internal/handle"
	"github.com/uptime-com/uptime-mcp/internal/log"
	"github.com/uptime-com/uptime-mcp/internal/server"
)

// Build-time variables set via ldflags.
var (
	Version = "dev"
	Commit  = "unknown"
)

func main() {
	// Handle -version flag
	for _, arg := range os.Args[1:] {
		if arg == "-version" || arg == "--version" {
			commit := Commit
			if len(commit) > 8 {
				commit = commit[:8]
			}
			fmt.Printf("%s (%s)\n", Version, commit)
			return
		}
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger := log.Logger(os.Args[1:])

	fx.New(
		fx.Supply(ctx, logger, server.Info{Version: Version, Commit: Commit}),
		fx.WithLogger(log.FxeventLogger),
		app.Module,
		handle.Module,
		server.Module,
	).Run()
}
