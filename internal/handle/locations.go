package handle

import (
	"go.uber.org/fx"

	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

type locationsHandler struct {
	service upapi.ProbeServersEndpoint
}

func provideLocationsHandler(c upapi.API) *locationsHandler {
	return &locationsHandler{service: c.ProbeServers()}
}

var locationsModule = fx.Module("tool.locations",
	fx.Provide(provideLocationsHandler),
	fx.Invoke(registerListLocationsTool),
	fx.Invoke(registerLocationResource),
)
