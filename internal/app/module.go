package app

import (
	"go.uber.org/fx"
)

var Module = fx.Module("app",
	fx.Provide(provideConfig),
	fx.Provide(provideLogger),
	fx.Provide(provideContext),
	fx.Invoke(attachSignalHandler),
)
