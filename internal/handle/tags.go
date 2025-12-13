package handle

import "go.uber.org/fx"

type tagsHandler struct{}

func provideTagsHandler() *tagsHandler {
	return &tagsHandler{}
}

var tagsModule = fx.Module("tool.tags",
	fx.Provide(provideTagsHandler),
	fx.Invoke(registerListTagsTool),
	fx.Invoke(registerTagResource),
	fx.Invoke(registerCreateTagTool),
	fx.Invoke(registerUpdateTagTool),
	fx.Invoke(registerDeleteTagTool),
)
