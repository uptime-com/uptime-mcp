package handle

import (
	"go.uber.org/fx"

	"github.com/uptime-com/uptime-mcp/internal/uptime"
)

type tags struct {
	service uptime.TagsService
}

func provideTags(c uptime.Client) *tags {
	return &tags{service: c.Tags()}
}

var tagsModule = fx.Module("tool.tags",
	fx.Provide(provideTags),
	fx.Invoke(registerListTagsTool),
	fx.Invoke(registerTagResource),
	fx.Invoke(registerCreateTagTool),
	fx.Invoke(registerUpdateTagTool),
	fx.Invoke(registerDeleteTagTool),
)
