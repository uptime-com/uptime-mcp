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
