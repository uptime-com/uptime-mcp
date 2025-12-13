package tools

import "github.com/uptime-com/uptime-mcp/internal/uptime"

type outages struct {
	service uptime.OutagesService
}

func provideOutages(c uptime.Client) *outages {
	return &outages{service: c.Outages()}
}
