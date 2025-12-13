package handle

import "go.uber.org/fx"

type outagesHandler struct{}

func provideOutagesHandler() *outagesHandler {
	return &outagesHandler{}
}

var outagesModule = fx.Module("tool.outages",
	fx.Provide(provideOutagesHandler),
	fx.Invoke(registerListOutagesTool),
	fx.Invoke(registerGetOutageTool),
	fx.Invoke(registerOutageResource),
)
