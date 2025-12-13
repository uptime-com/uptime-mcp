package tools

import "go.uber.org/fx"

// Module aggregates all tool modules.
var Module = fx.Module("tools",
	// Providers for handler structs
	fx.Provide(provideChecksHandler),
	fx.Provide(provideTags),
	fx.Provide(provideOutages),

	// Checks - read operations
	ListChecksToolModule,
	GetCheckToolModule,
	DeleteCheckToolModule,
	GetCheckStatsToolModule,

	// Checks - create by type
	CreateHTTPCheckToolModule,
	CreateDNSCheckToolModule,
	CreateSSLCheckToolModule,
	CreateTCPCheckToolModule,
	CreateICMPCheckToolModule,
	CreateSMTPCheckToolModule,
	CreateIMAPCheckToolModule,
	CreatePOPCheckToolModule,

	// Outages
	ListOutagesToolModule,
	GetOutageToolModule,

	// Tags
	ListTagsToolModule,
	GetTagToolModule,
	CreateTagToolModule,
	UpdateTagToolModule,
	DeleteTagToolModule,
)
