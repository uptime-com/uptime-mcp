package handle

import "go.uber.org/fx"

type contactsHandler struct{}

func provideContactsHandler() *contactsHandler {
	return &contactsHandler{}
}

var contactsModule = fx.Module("tool.contacts",
	fx.Provide(provideContactsHandler),
	fx.Invoke(registerListContactsTool),
	fx.Invoke(registerContactResource),
	fx.Invoke(registerCreateContactTool),
	fx.Invoke(registerDeleteContactTool),
)
