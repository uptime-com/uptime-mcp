package tools

import "github.com/uptime-com/uptime-mcp/internal/uptime"

type checksHandler struct {
	service uptime.ChecksService
}

func provideChecksHandler(c uptime.Client) *checksHandler {
	return &checksHandler{service: c.Checks()}
}
