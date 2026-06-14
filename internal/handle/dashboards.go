package handle

import "go.uber.org/fx"

type dashboardsHandler struct{}

func provideDashboardsHandler() *dashboardsHandler {
	return &dashboardsHandler{}
}

var dashboardsModule = fx.Module("tool.dashboards",
	fx.Provide(provideDashboardsHandler),
	fx.Invoke(registerListDashboardsTool),
	fx.Invoke(registerGetDashboardTool),
	fx.Invoke(registerDashboardResource),
	fx.Invoke(registerCreateDashboardTool),
	fx.Invoke(registerUpdateDashboardTool),
	fx.Invoke(registerDeleteDashboardTool),
)
