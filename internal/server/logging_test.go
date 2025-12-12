package server

import (
	"bytes"
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := buf.String()

		// Expected format: "tools/call 0\n" (method followed by duration in ms)
		pattern := regexp.MustCompile(`^tools/call \d+\n$`)
		if !pattern.MatchString(output) {
			t.Errorf("unexpected log output: %q", output)
		}
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
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := buf.String()

		// Should log tool name instead of method
		pattern := regexp.MustCompile(`^list_checks \d+\n$`)
		if !pattern.MatchString(output) {
			t.Errorf("unexpected log output: %q, expected tool name 'list_checks'", output)
		}
	})

	t.Run("logs on error", func(t *testing.T) {
		var buf bytes.Buffer

		next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			return nil, context.Canceled
		}

		middleware := loggingMiddleware(&buf)
		handler := middleware(next)

		_, err := handler(context.Background(), "resources/read", nil)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got: %v", err)
		}

		output := buf.String()

		// Should still log even when handler returns error
		pattern := regexp.MustCompile(`^resources/read \d+\n$`)
		if !pattern.MatchString(output) {
			t.Errorf("unexpected log output: %q", output)
		}
	})
}
