package handle

import "go.uber.org/fx"

// Module aggregates all tool modules.
var Module = fx.Module("tools",
	alertsModule,
	checksModule,
	cloudStatusModule,
	contactsModule,
	dashboardsModule,
	locationsModule,
	outagesModule,
	statusPagesModule,
	tagsModule,
)
