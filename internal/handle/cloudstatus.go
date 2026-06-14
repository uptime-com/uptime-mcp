package handle

import (
	"github.com/uptime-com/uptime-mcp/internal/cloudstatus"
	"go.uber.org/fx"
)

type cloudStatusHandler struct {
	index *cloudstatus.Index
}

func provideCloudStatusHandler() (*cloudStatusHandler, error) {
	idx, err := cloudstatus.NewIndex()
	if err != nil {
		return nil, err
	}
	return &cloudStatusHandler{index: idx}, nil
}

var cloudStatusModule = fx.Module("tool.cloudstatus",
	fx.Provide(provideCloudStatusHandler),
	fx.Invoke(registerListCloudStatusProvidersTool),
	fx.Invoke(registerSearchCloudStatusServicesTool),
)
