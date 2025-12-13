package handle

import "go.uber.org/fx"

type locationsHandler struct{}

func provideLocationsHandler() *locationsHandler {
	return &locationsHandler{}
}

var locationsModule = fx.Module("tool.locations",
	fx.Provide(provideLocationsHandler),
	fx.Invoke(registerListLocationsTool),
	fx.Invoke(registerGetLocationTool),
	fx.Invoke(registerLocationResource),
)
