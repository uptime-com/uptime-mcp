//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestE2E_ListOutages(t *testing.T) {
	session := makeClientSession(t)
	ctx := context.Background()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_outages",
		Arguments: map[string]any{
			"page_size": 10,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	text := result.Content[0].(*mcp.TextContent).Text
	require.Contains(t, text, "Found")
	t.Logf("Response:\n%s", text)
}

func TestE2E_ListOutages_WithTypeFilter(t *testing.T) {
	session := makeClientSession(t)
	ctx := context.Background()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_outages",
		Arguments: map[string]any{
			"type":      "HTTP",
			"page_size": 5,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	text := result.Content[0].(*mcp.TextContent).Text
	require.Contains(t, text, "Found")
	t.Logf("Response (HTTP only):\n%s", text)
}

func TestE2E_ReadOutageResource(t *testing.T) {
	session := makeClientSession(t)
	ctx := context.Background()

	// First, list outages to get a valid ID
	listResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_outages",
		Arguments: map[string]any{
			"page_size": 1,
		},
	})
	require.NoError(t, err)
	require.Len(t, listResult.Content, 1)

	text := listResult.Content[0].(*mcp.TextContent).Text
	// Extract outage ID from text like "- [123] CheckName (Type) - status"
	outageID := extractOutageID(t, text)
	if outageID == "" {
		t.Skip("no outages found in account")
	}

	// Now read the resource
	uri := fmt.Sprintf("uptime://outages/%s", outageID)
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

func TestE2E_ReadOutageResource_InvalidID(t *testing.T) {
	session := makeClientSession(t)
	ctx := context.Background()

	result, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
		URI: "uptime://outages/999999999",
	})

	require.Error(t, err)
	require.Nil(t, result)
}

// extractOutageID extracts the first outage ID from list output like "- [123] CheckName".
func extractOutageID(t *testing.T, text string) string {
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
