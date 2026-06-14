package server

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/fx"
)

// Info holds server metadata injected at build time.
type Info struct {
	Version string
	Commit  string
}

var Module = fx.Module("server",
	fx.Provide(New),
	fx.Invoke(Run),
)

func New(info Info) *mcp.Server {
	return mcp.NewServer(&mcp.Implementation{
		Name:    "uptime-mcp",
		Version: info.Version,
	}, &mcp.ServerOptions{
		Instructions: instructions,
		// Disable list_changed notifications — our tool/resource sets are
		// static, and the go-sdk fires these notifications before the client
		// sends "initialize" in stdio mode, which breaks the protocol.
		Capabilities: &mcp.ServerCapabilities{
			Tools:     &mcp.ToolCapabilities{},
			Resources: &mcp.ResourceCapabilities{},
		},
	})
}
