package tools

import (
	"go.uber.org/fx"

	"github.com/uptime-com/uptime-mcp/internal/uptime"
)

type checksHandler struct {
	service uptime.ChecksService
}

func provideChecksHandler(c uptime.Client) *checksHandler {
	return &checksHandler{service: c.Checks()}
}

var checksModule = fx.Module("tool.checks",
	fx.Provide(provideChecksHandler),
)
