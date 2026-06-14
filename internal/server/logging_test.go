package server

import (
	"bytes"
	"context"
	"net/http"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggingMiddleware(t *testing.T) {
	t.Run("logs method and duration", func(t *testing.T) {
		var buf bytes.Buffer

		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			return nil, nil
		}

		middleware := loggingMiddleware(&buf)
		handler := middleware(next)

		_, err := handler(context.Background(), "tools/call", nil)
		require.NoError(t, err)

		output := buf.String()

		assert.Contains(t, output, "method=tools/call")
		assert.Contains(t, output, "latency=")
		assert.Contains(t, output, "latency_ms=")
	})

	t.Run("logs tool name for tool calls", func(t *testing.T) {
		var buf bytes.Buffer

		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			return nil, nil
		}

		middleware := loggingMiddleware(&buf)
		handler := middleware(next)

		req := &mcp.CallToolRequest{
			Params: &mcp.CallToolParamsRaw{
				Name: "list_checks",
			},
		}

		_, err := handler(context.Background(), "tools/call", req)
		require.NoError(t, err)

		output := buf.String()

		assert.Contains(t, output, "tool=list_checks")
	})

	t.Run("logs user agent", func(t *testing.T) {
		var buf bytes.Buffer

		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			return nil, nil
		}

		middleware := loggingMiddleware(&buf)
		handler := middleware(next)

		header := make(http.Header)
		header.Set("User-Agent", "Claude/1.0")
		req := &mcp.ServerRequest[mcp.Params]{
			Extra: &mcp.RequestExtra{
				Header: header,
			},
		}

		_, err := handler(context.Background(), "initialize", req)
		require.NoError(t, err)

		output := buf.String()

		assert.Contains(t, output, "user_agent=Claude/1.0")
	})

	t.Run("logs X-Forwarded-For as client IP", func(t *testing.T) {
		var buf bytes.Buffer

		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			return nil, nil
		}

		middleware := loggingMiddleware(&buf)
		handler := middleware(next)

		header := make(http.Header)
		header.Set("X-Forwarded-For", "192.168.1.100")
		req := &mcp.ServerRequest[mcp.Params]{
			Extra: &mcp.RequestExtra{
				Header: header,
			},
		}

		_, err := handler(context.Background(), "initialize", req)
		require.NoError(t, err)

		output := buf.String()

		assert.Contains(t, output, "client_ip=192.168.1.100")
	})

	t.Run("logs X-Real-IP as client IP fallback", func(t *testing.T) {
		var buf bytes.Buffer

		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			return nil, nil
		}

		middleware := loggingMiddleware(&buf)
		handler := middleware(next)

		header := make(http.Header)
		header.Set("X-Real-IP", "10.0.0.1")
		req := &mcp.ServerRequest[mcp.Params]{
			Extra: &mcp.RequestExtra{
				Header: header,
			},
		}

		_, err := handler(context.Background(), "initialize", req)
		require.NoError(t, err)

		output := buf.String()

		assert.Contains(t, output, "client_ip=10.0.0.1")
	})

	t.Run("X-Forwarded-For takes precedence over X-Real-IP", func(t *testing.T) {
		var buf bytes.Buffer

		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			return nil, nil
		}

		middleware := loggingMiddleware(&buf)
		handler := middleware(next)

		header := make(http.Header)
		header.Set("X-Forwarded-For", "203.0.113.50")
		header.Set("X-Real-IP", "10.0.0.1")
		req := &mcp.ServerRequest[mcp.Params]{
			Extra: &mcp.RequestExtra{
				Header: header,
			},
		}

		_, err := handler(context.Background(), "initialize", req)
		require.NoError(t, err)

		output := buf.String()

		assert.Contains(t, output, "client_ip=203.0.113.50")
		assert.NotContains(t, output, "10.0.0.1")
	})

	t.Run("logs error", func(t *testing.T) {
		var buf bytes.Buffer

		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			return nil, context.Canceled
		}

		middleware := loggingMiddleware(&buf)
		handler := middleware(next)

		_, err := handler(context.Background(), "resources/read", nil)
		require.ErrorIs(t, err, context.Canceled)

		output := buf.String()

		assert.Contains(t, output, "error=")
	})
}
