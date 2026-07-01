package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/uptime-com/uptime-mcp/internal/app"
)

// newTestMux builds an HTTP handler equivalent to production: a real MCP server
// with one tool, the same receiving middleware as runHTTP, routed through
// newHTTPMux.
func newTestMux(t *testing.T) http.Handler {
	t.Helper()

	srv := mcp.NewServer(&mcp.Implementation{Name: "uptime-mcp", Version: "test"}, nil)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "ping",
		Description: "test tool",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "pong"}},
		}, nil, nil
	})
	srv.AddReceivingMiddleware(httpTokenMiddleware())

	return newHTTPMux(srv, app.Config{})
}

// initializeBody is a minimal JSON-RPC "initialize" request.
const initializeBody = `{"jsonrpc":"2.0","id":1,"method":"initialize",` +
	`"params":{"protocolVersion":"2025-06-18","capabilities":{},` +
	`"clientInfo":{"name":"test","version":"0"}}}`

// TestStatelessSessionHandling asserts the streamable handler runs in stateless
// mode: a request carrying an unknown Mcp-Session-Id must be served, not
// rejected with "session not found" (404). Stateful mode pins each session to
// the pod that created it, which breaks hosted deployments behind a load
// balancer with multiple replicas.
func TestStatelessSessionHandling(t *testing.T) {
	server := httptest.NewServer(newTestMux(t))
	defer server.Close()

	t.Run("unknown session id is served, not 404", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, server.URL+"/", strings.NewReader(initializeBody))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		req.Header.Set("Mcp-Session-Id", "bogus-nonexistent")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"unknown session id must not be rejected in stateless mode; body: %s", body)
		assert.NotContains(t, string(body), "session not found")
		assert.Contains(t, string(body), "protocolVersion")
	})

	t.Run("missing bearer token returns 401", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, server.URL+"/", strings.NewReader(initializeBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("WWW-Authenticate"), "Bearer")
	})
}
