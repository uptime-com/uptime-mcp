//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestE2E_ListChecks(t *testing.T) {
	session := makeClientSession(t)
	ctx := context.Background()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_checks",
		Arguments: map[string]any{
			"page_size": 5,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	text := result.Content[0].(*mcp.TextContent).Text
	requirePaginationHeader(t, text)
	t.Logf("Response:\n%s", text)
}

func TestE2E_ListChecks_WithFilter(t *testing.T) {
	session := makeClientSession(t)
	ctx := context.Background()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_checks",
		Arguments: map[string]any{
			"type":      "HTTP",
			"page_size": 3,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	text := result.Content[0].(*mcp.TextContent).Text
	requirePaginationHeader(t, text)
	t.Logf("Response (HTTP only):\n%s", text)
}

func TestE2E_ReadCheckResource(t *testing.T) {
	session := makeClientSession(t)
	ctx := context.Background()

	// First, list checks to get a valid ID
	listResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_checks",
		Arguments: map[string]any{
			"page_size": 1,
		},
	})
	require.NoError(t, err)
	require.Len(t, listResult.Content, 1)

	text := listResult.Content[0].(*mcp.TextContent).Text
	// Extract check ID from text like "- [123] Name (Type) - Address"
	checkID := extractCheckID(t, text)
	if checkID == "" {
		t.Skip("no checks found in account")
	}

	// Now read the resource
	uri := fmt.Sprintf("uptime://checks/%s", checkID)
	result, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
		URI: uri,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Contents, 1)

	content := result.Contents[0]
	require.Equal(t, uri, content.URI)
	require.Equal(t, "text/plain", content.MIMEType)
	require.NotEmpty(t, content.Text)

	t.Logf("Resource URI: %s", uri)
	t.Logf("Response:\n%s", content.Text)
}

func TestE2E_ReadCheckResource_InvalidID(t *testing.T) {
	session := makeClientSession(t)
	ctx := context.Background()

	result, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
		URI: "uptime://checks/999999999",
	})

	// Should return error for non-existent check
	require.Error(t, err)
	require.Nil(t, result)
}

// requirePaginationHeader asserts that text contains a valid pagination header.
// Either "Found X results" (single page) or "Showing X of Y total" (multiple pages).
func requirePaginationHeader(t *testing.T, text string) {
	t.Helper()
	hasPagination := strings.Contains(text, "Found") || strings.Contains(text, "Showing")
	require.True(t, hasPagination, "expected pagination header in output")
}

func TestE2E_GetCheckStats(t *testing.T) {
	session := makeClientSession(t)
	ctx := context.Background()

	// First, list checks to get a valid ID
	listResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_checks",
		Arguments: map[string]any{
			"page_size": 1,
		},
	})
	require.NoError(t, err)
	require.Len(t, listResult.Content, 1)

	text := listResult.Content[0].(*mcp.TextContent).Text
	checkIDStr := extractCheckID(t, text)
	if checkIDStr == "" {
		t.Skip("no checks found in account")
	}
	checkID, err := strconv.ParseInt(checkIDStr, 10, 64)
	require.NoError(t, err)

	// Get stats for the check
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "get_check_stats",
		Arguments: map[string]any{
			"id": checkID,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	statsText := result.Content[0].(*mcp.TextContent).Text
	require.Contains(t, statsText, "Statistics for check")
	require.Contains(t, statsText, "Totals:")

	t.Logf("Check ID: %d", checkID)
	t.Logf("Response:\n%s", statsText)
}

func TestE2E_GetCheckStats_WithDateRange(t *testing.T) {
	session := makeClientSession(t)
	ctx := context.Background()

	// First, list checks to get a valid ID
	listResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_checks",
		Arguments: map[string]any{
			"page_size": 1,
		},
	})
	require.NoError(t, err)
	require.Len(t, listResult.Content, 1)

	text := listResult.Content[0].(*mcp.TextContent).Text
	checkIDStr := extractCheckID(t, text)
	if checkIDStr == "" {
		t.Skip("no checks found in account")
	}
	checkID, err := strconv.ParseInt(checkIDStr, 10, 64)
	require.NoError(t, err)

	// Get stats with date range
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "get_check_stats",
		Arguments: map[string]any{
			"id":         checkID,
			"start_date": "2024-12-01",
			"end_date":   "2024-12-13",
		},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	statsText := result.Content[0].(*mcp.TextContent).Text
	require.Contains(t, statsText, "Statistics for check")
	require.Contains(t, statsText, "Period: 2024-12-01 to 2024-12-13")

	t.Logf("Check ID: %d", checkID)
	t.Logf("Response:\n%s", statsText)
}

// extractCheckID extracts the first check ID from list output like "- [123] Name".
func extractCheckID(t *testing.T, text string) string {
	t.Helper()
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "- [") {
			start := strings.Index(line, "[")
			end := strings.Index(line, "]")
			if start != -1 && end > start {
				return line[start+1 : end]
			}
		}
	}
	return ""
}
