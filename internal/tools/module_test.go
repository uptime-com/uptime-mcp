package tools

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// TestToolRegistration validates that all tools can be registered without panic.
// This catches issues like invalid jsonschema tags before deployment.
func TestToolRegistration(t *testing.T) {
	app := fxtest.New(t,
		fx.Provide(func() *mcp.Server {
			return mcp.NewServer(&mcp.Implementation{
				Name:    "uptime-mcp-test",
				Version: "test",
			}, nil)
		}),
		Module,
	)

	// If tool registration panics due to invalid schemas, fxtest.New will fail
	app.RequireStart()
	app.RequireStop()
}
