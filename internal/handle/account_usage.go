package handle

import "go.uber.org/fx"

type accountUsageHandler struct{}

func provideAccountUsageHandler() *accountUsageHandler {
	return &accountUsageHandler{}
}

var accountUsageModule = fx.Module("tool.account_usage",
	fx.Provide(provideAccountUsageHandler),
	fx.Invoke(registerGetAccountUsageTool),
)
