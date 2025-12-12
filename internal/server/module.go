package server

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/fx"
)

var Module = fx.Module("server",
	fx.Provide(New),
	fx.Invoke(Run),
)

func New() *mcp.Server {
	return mcp.NewServer(&mcp.Implementation{
		Name:    "uptime-mcp",
		Version: "0.1.0",
	}, nil)
}
