package handle

import (
	"go.uber.org/fx"

	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

type outages struct {
	service upapi.OutagesEndpoint
}

func provideOutages(c upapi.API) *outages {
	return &outages{service: c.Outages()}
}

var outagesModule = fx.Module("tool.outages",
	fx.Provide(provideOutages),
	fx.Invoke(registerListOutagesTool),
	fx.Invoke(registerOutageResource),
)
