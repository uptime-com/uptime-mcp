package handle

import (
	"go.uber.org/fx"

	"github.com/uptime-com/uptime-mcp/internal/uptime"
)

type outages struct {
	service uptime.OutagesService
}

func provideOutages(c uptime.Client) *outages {
	return &outages{service: c.Outages()}
}

var outagesModule = fx.Module("tool.outages",
	fx.Provide(provideOutages),
	fx.Invoke(registerListOutagesTool),
	fx.Invoke(registerOutageResource),
)
