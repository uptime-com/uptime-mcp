package uptime

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/fx"
)

var Module = fx.Module("provider.uptime",
	fx.Provide(NewProvider),
	fx.Invoke(Register),
)

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

type RegisterParams struct {
	fx.In
	Server   *mcp.Server
	Provider *Provider
}

func Register(p RegisterParams) {
	// Checks - read operations
	mcp.AddTool(p.Server, listChecksTool, p.Provider.handleListChecks)
	mcp.AddTool(p.Server, getCheckTool, p.Provider.handleGetCheck)
	mcp.AddTool(p.Server, deleteCheckTool, p.Provider.handleDeleteCheck)
	mcp.AddTool(p.Server, getCheckStatsTool, p.Provider.handleGetCheckStats)

	// Checks - create by type
	mcp.AddTool(p.Server, createHTTPCheckTool, p.Provider.handleCreateHTTPCheck)
	mcp.AddTool(p.Server, createDNSCheckTool, p.Provider.handleCreateDNSCheck)
	mcp.AddTool(p.Server, createSSLCheckTool, p.Provider.handleCreateSSLCheck)
	mcp.AddTool(p.Server, createTCPCheckTool, p.Provider.handleCreateTCPCheck)
	mcp.AddTool(p.Server, createICMPCheckTool, p.Provider.handleCreateICMPCheck)
	mcp.AddTool(p.Server, createSMTPCheckTool, p.Provider.handleCreateSMTPCheck)
	mcp.AddTool(p.Server, createIMAPCheckTool, p.Provider.handleCreateIMAPCheck)
	mcp.AddTool(p.Server, createPOPCheckTool, p.Provider.handleCreatePOPCheck)

	// Outages
	mcp.AddTool(p.Server, listOutagesTool, p.Provider.handleListOutages)
	mcp.AddTool(p.Server, getOutageTool, p.Provider.handleGetOutage)

	// Tags
	mcp.AddTool(p.Server, listTagsTool, p.Provider.handleListTags)
	mcp.AddTool(p.Server, getTagTool, p.Provider.handleGetTag)
	mcp.AddTool(p.Server, createTagTool, p.Provider.handleCreateTag)
	mcp.AddTool(p.Server, updateTagTool, p.Provider.handleUpdateTag)
	mcp.AddTool(p.Server, deleteTagTool, p.Provider.handleDeleteTag)
}
