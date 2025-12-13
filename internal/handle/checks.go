package handle

import "go.uber.org/fx"

type checksHandler struct{}

func provideChecksHandler() *checksHandler {
	return &checksHandler{}
}

var checksModule = fx.Module("tool.checks",
	fx.Provide(provideChecksHandler),
	fx.Invoke(registerListChecksTool),
	fx.Invoke(registerGetCheckTool),
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
