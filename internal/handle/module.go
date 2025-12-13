package handle

import "go.uber.org/fx"

// Module aggregates all tool modules.
var Module = fx.Module("tools",
	checksModule,
	contactsModule,
	locationsModule,
	outagesModule,
	tagsModule,
)
