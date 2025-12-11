package main

import (
	"flag"
	"fmt"
	"os"

	"go.uber.org/fx"

	"github.com/uptime-com/uptime-mcp/internal/app"
	"github.com/uptime-com/uptime-mcp/internal/server"
	"github.com/uptime-com/uptime-mcp/internal/uptime"
)

func main() {
	var cfg app.Config

	flag.StringVar(&cfg.Transport, "transport", "stdio", "Transport mode: stdio or http")
	flag.StringVar(&cfg.ListenAddr, "listen", ":8080", "HTTP listen address (only used with -transport=http)")
	flag.StringVar(&cfg.APIBaseURL, "api-url", "", "Uptime.com API base URL (default: https://uptime.com/api/v1/)")
	flag.Parse()

	if cfg.Transport != "stdio" && cfg.Transport != "http" {
		fmt.Fprintf(os.Stderr, "invalid transport: %s (must be stdio or http)\n", cfg.Transport)
		os.Exit(1)
	}

	fx.New(
		fx.NopLogger,
		fx.Supply(cfg),
		fx.Provide(app.ProvideContext),
		fx.Invoke(app.AttachSignalHandler),
		server.Module,
		uptime.Module,
	).Run()
}
