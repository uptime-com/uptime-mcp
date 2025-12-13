package handle

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/fx/fxtest"
)

// TestToolRegistration validates that all tools can be registered without panic.
// This catches issues like invalid jsonschema tags before deployment.
func TestToolRegistration(t *testing.T) {
	// Set up client mock with service accessor expectations
	client := newClientMock(t)
	client.EXPECT().Checks().Return(newChecksServiceMock(t)).Maybe()
	client.EXPECT().Contacts().Return(newContactsServiceMock(t)).Maybe()
	client.EXPECT().Outages().Return(newOutagesServiceMock(t)).Maybe()
	client.EXPECT().Tags().Return(newTagsServiceMock(t)).Maybe()

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
		fx.Provide(func() upapi.API {
			return client
		}),
		Module,
	)

	// If tool registration panics due to invalid schemas, fxtest.New will fail
	app.RequireStart()
	app.RequireStop()
}
