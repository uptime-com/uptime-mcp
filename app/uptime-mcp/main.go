package main

import (
	"go.uber.org/fx"

	"github.com/uptime-com/uptime-mcp/internal/app"
	"github.com/uptime-com/uptime-mcp/internal/server"
	"github.com/uptime-com/uptime-mcp/internal/tools"
)

func main() {
	fx.New(
		fx.NopLogger,
		app.Module,
		tools.Module,
		server.Module,
	).Run()
}
