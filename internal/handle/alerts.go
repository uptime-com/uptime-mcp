package handle

import "go.uber.org/fx"

type alertsHandler struct{}

func provideAlertsHandler() *alertsHandler {
	return &alertsHandler{}
}

var alertsModule = fx.Module("tool.alerts",
	fx.Provide(provideAlertsHandler),
	fx.Invoke(registerListAlertsTool),
	fx.Invoke(registerGetAlertTool),
	fx.Invoke(registerIgnoreAlertTool),
	fx.Invoke(registerAlertResource),
)
