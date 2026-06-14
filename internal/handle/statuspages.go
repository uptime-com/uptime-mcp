package handle

import "go.uber.org/fx"

type statusPagesHandler struct{}

func provideStatusPagesHandler() *statusPagesHandler {
	return &statusPagesHandler{}
}

var statusPagesModule = fx.Module("tool.statuspages",
	fx.Provide(provideStatusPagesHandler),
	// Pages
	fx.Invoke(registerListStatusPagesTool),
	fx.Invoke(registerGetStatusPageTool),
	fx.Invoke(registerStatusPageResource),
	fx.Invoke(registerCreateStatusPageTool),
	fx.Invoke(registerUpdateStatusPageTool),
	fx.Invoke(registerDeleteStatusPageTool),
	// Components
	fx.Invoke(registerListStatusPageComponentsTool),
	fx.Invoke(registerGetStatusPageComponentTool),
	fx.Invoke(registerCreateStatusPageComponentTool),
	fx.Invoke(registerUpdateStatusPageComponentTool),
	fx.Invoke(registerDeleteStatusPageComponentTool),
	// Incidents
	fx.Invoke(registerListStatusPageIncidentsTool),
	fx.Invoke(registerGetStatusPageIncidentTool),
	fx.Invoke(registerCreateStatusPageIncidentTool),
	fx.Invoke(registerUpdateStatusPageIncidentTool),
	fx.Invoke(registerDeleteStatusPageIncidentTool),
)
