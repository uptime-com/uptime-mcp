package handle

import (
	"context"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"

	"github.com/uptime-com/uptime-mcp/internal/app"
)

// textResult creates a successful text response.
func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

// clientFromContext retrieves the cached API client from session context.
func clientFromContext(ctx context.Context) (upapi.API, error) {
	session := app.SessionFromContext(ctx)
	if session == nil || session.Client == nil {
		return nil, errors.New("no API client in context")
	}
	return session.Client, nil
}
