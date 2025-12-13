package handle

import (
	"go.uber.org/fx"

	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

type tags struct {
	service upapi.TagsEndpoint
}

func provideTags(c upapi.API) *tags {
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
