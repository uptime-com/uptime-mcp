package tools

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/fx/fxtest"

	"github.com/uptime-com/uptime-mcp/internal/uptime"
)

// mockClient provides a minimal uptime.Client for fx wiring tests.
type mockClient struct{}

func (m *mockClient) Checks() uptime.ChecksService   { return nil }
func (m *mockClient) Tags() uptime.TagsService       { return nil }
func (m *mockClient) Outages() uptime.OutagesService { return nil }

// TestToolRegistration validates that all tools can be registered without panic.
// This catches issues like invalid jsonschema tags before deployment.
func TestToolRegistration(t *testing.T) {
	app := fxtest.New(t,
		fx.WithLogger(func() fxevent.Logger {
			return fxevent.NopLogger
		}),
		fx.Provide(func() *mcp.Server {
			return mcp.NewServer(&mcp.Implementation{
				Name:    "uptime-mcp-test",
				Version: "test",
			}, nil)
		}),
		fx.Provide(func() uptime.Client {
			return &mockClient{}
		}),
		Module,
	)

	// If tool registration panics due to invalid schemas, fxtest.New will fail
	app.RequireStart()
	app.RequireStop()
}
