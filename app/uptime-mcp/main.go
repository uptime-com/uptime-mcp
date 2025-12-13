package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/fx"

	"github.com/uptime-com/uptime-mcp/internal/app"
	"github.com/uptime-com/uptime-mcp/internal/handle"
	"github.com/uptime-com/uptime-mcp/internal/log"
	"github.com/uptime-com/uptime-mcp/internal/server"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger := log.Logger(os.Args[1:])

	fx.New(
		fx.Supply(ctx, logger),
		fx.WithLogger(log.FxeventLogger),
		app.Module,
		handle.Module,
		server.Module,
	).Run()
}
