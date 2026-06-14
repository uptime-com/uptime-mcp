package handle

import (
	"context"
	"errors"
	"fmt"

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

// optString returns a pointer to s, or nil when s is empty, so optional API
// fields with omit/set semantics are omitted rather than sent as empty.
func optString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// optBool returns a pointer to b, or nil when b is false, so optional flags are
// omitted rather than sent as false.
func optBool(b bool) *bool {
	if !b {
		return nil
	}
	return &b
}

// formatPaginationHeader formats pagination info for tool output.
func formatPaginationHeader(totalCount int64, page, pageSize int64, resultCount int) string {
	if totalCount <= pageSize {
		return fmt.Sprintf("Found %d results.\n\n", totalCount)
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if page >= totalPages {
		return fmt.Sprintf("Showing %d of %d total (page %d of %d, final page).\n\n",
			resultCount, totalCount, page, totalPages)
	}

	return fmt.Sprintf("Showing %d of %d total (page %d of %d).\nUse page=%d to see next page.\n\n",
		resultCount, totalCount, page, totalPages, page+1)
}

// clientFromContext retrieves the cached API client from session context.
func clientFromContext(ctx context.Context) (upapi.API, error) {
	session := app.SessionFromContext(ctx)
	if session == nil || session.Client == nil {
		return nil, errors.New("no API client in context")
	}
	return session.Client, nil
}
