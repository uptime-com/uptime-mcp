package uptime

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api "github.com/uptime-com/uptime-client-go"

	"github.com/uptime-com/uptime-mcp/internal/app"
)

// getClient retrieves the authenticated Uptime client from context.
func getClient(ctx context.Context) (*api.Client, error) {
	session := app.SessionFromContext(ctx)
	if session == nil || session.Client == nil {
		return nil, fmt.Errorf("not authenticated: missing Uptime.com API token")
	}
	return session.Client, nil
}

// textResult creates a successful text response.
func textResult(text string) *mcp.CallToolResultFor[any] {
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

// errorResult creates an error response.
func errorResult(err error) *mcp.CallToolResultFor[any] {
	return &mcp.CallToolResultFor[any]{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
	}
}
