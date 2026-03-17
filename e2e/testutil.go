//go:build e2e

package e2e

import (
	"context"
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/fx/fxtest"

	"github.com/uptime-com/uptime-mcp/internal/app"
	"github.com/uptime-com/uptime-mcp/internal/handle"
)

// makeClientSession creates an MCP client session connected to the server via in-memory transport.
// Uses full fx DI stack to register all tools and resources.
func makeClientSession(t *testing.T) *mcp.ClientSession {
	t.Helper()

	bearerToken := os.Getenv("UPTIME_BEARER_TOKEN")
	if bearerToken == "" {
		t.Fatal("UPTIME_BEARER_TOKEN environment variable is required for e2e tests")
	}

	baseURL := os.Getenv("UPTIME_API_URL")
	if baseURL == "" {
		baseURL = "https://uptime.com/api/v1/"
	}

	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	// Create Uptime API client
	opts := []upapi.Option{
		upapi.WithBearerToken(bearerToken),
	}
	if baseURL != "" {
		opts = append(opts, upapi.WithBaseURL(baseURL))
	}
	uptimeClient, err := upapi.New(opts...)
	require.NoError(t, err)

	// Create MCP server
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "uptime-mcp",
		Version: "e2e-test",
	}, nil)

	// Add middleware that injects session into context
	srv.AddReceivingMiddleware(clientMiddleware(uptimeClient))

	// Use fx to wire up all tools
	fxApp := fxtest.New(t,
		fx.WithLogger(func() fxevent.Logger {
			return fxevent.NopLogger
		}),
		fx.Supply(srv),
		handle.Module,
	)
	fxApp.RequireStart()
	t.Cleanup(func() { fxApp.RequireStop() })

	// Connect server first (MCP protocol requirement)
	go func() {
		_, err := srv.Connect(ctx, serverTransport, nil)
		if err != nil {
			t.Logf("server connect error: %v", err)
		}
	}()

	// Create and connect MCP client
	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "e2e-test"}, nil)
	session, err := mcpClient.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = session.Close()
	})

	return session
}

// clientMiddleware creates a middleware that injects the API client into context.
func clientMiddleware(client upapi.API) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			session := &app.Session{Client: client}
			ctx = app.ContextWithSession(ctx, session)
			return next(ctx, method, req)
		}
	}
}
