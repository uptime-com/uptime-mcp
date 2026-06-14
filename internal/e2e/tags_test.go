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

func TestE2E_ListTags(t *testing.T) {
	session := makeClientSession(t)
	ctx := context.Background()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_tags",
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

func TestE2E_ListTags_WithSearch(t *testing.T) {
	session := makeClientSession(t)
	ctx := context.Background()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_tags",
		Arguments: map[string]any{
			"search":    "prod",
			"page_size": 5,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	text := result.Content[0].(*mcp.TextContent).Text
	require.Contains(t, text, "Found")
	t.Logf("Response (search 'prod'):\n%s", text)
}

func TestE2E_ReadTagResource(t *testing.T) {
	session := makeClientSession(t)
	ctx := context.Background()

	// First, list tags to get a valid ID
	listResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_tags",
		Arguments: map[string]any{
			"page_size": 1,
		},
	})
	require.NoError(t, err)
	require.Len(t, listResult.Content, 1)

	text := listResult.Content[0].(*mcp.TextContent).Text
	// Extract tag ID from text like "- [123] TagName (5 checks)"
	tagID := extractTagID(t, text)
	if tagID == "" {
		t.Skip("no tags found in account")
	}

	// Now read the resource
	uri := fmt.Sprintf("uptime://tags/%s", tagID)
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

func TestE2E_ReadTagResource_InvalidID(t *testing.T) {
	session := makeClientSession(t)
	ctx := context.Background()

	result, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
		URI: "uptime://tags/999999999",
	})

	require.Error(t, err)
	require.Nil(t, result)
}

// extractTagID extracts the first tag ID from list output like "- [123] TagName".
func extractTagID(t *testing.T, text string) string {
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
