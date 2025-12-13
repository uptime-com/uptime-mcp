package handle

import (
	"go.uber.org/fx"

	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

type contactsHandler struct {
	service upapi.ContactsEndpoint
}

func provideContactsHandler(c upapi.API) *contactsHandler {
	return &contactsHandler{service: c.Contacts()}
}

var contactsModule = fx.Module("tool.contacts",
	fx.Provide(provideContactsHandler),
	fx.Invoke(registerListContactsTool),
	fx.Invoke(registerContactResource),
	fx.Invoke(registerCreateContactTool),
	fx.Invoke(registerDeleteContactTool),
)
