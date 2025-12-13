package handle

import (
	"go.uber.org/fx"

	"github.com/uptime-com/uptime-mcp/internal/uptime"
)

type checksHandler struct {
	service uptime.ChecksService
}

func provideChecksHandler(c uptime.Client) *checksHandler {
	return &checksHandler{service: c.Checks()}
}

var checksModule = fx.Module("tool.checks",
	fx.Provide(provideChecksHandler),
	fx.Invoke(registerListChecksTool),
	fx.Invoke(registerCheckResource),
	fx.Invoke(registerDeleteCheckTool),
	fx.Invoke(registerGetCheckStatsTool),
	fx.Invoke(registerCreateHTTPCheckTool),
	fx.Invoke(registerCreateDNSCheckTool),
	fx.Invoke(registerCreateSSLCheckTool),
	fx.Invoke(registerCreateTCPCheckTool),
	fx.Invoke(registerCreateICMPCheckTool),
	fx.Invoke(registerCreateIMAPCheckTool),
	fx.Invoke(registerCreatePOPCheckTool),
	fx.Invoke(registerCreateSMTPCheckTool),
)
