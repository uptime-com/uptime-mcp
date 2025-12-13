//go:build acc

package handle

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestAcc_HandleListTags(t *testing.T) {
	client := newAcceptanceClient(t)

	h := &tags{service: client.Tags()}
	result, _, err := h.HandleListTags(context.Background(), nil, listTagsInput{
		PageSize: 10,
	})

	require.NoError(t, err)
	require.Len(t, result.Content, 1)

	text := result.Content[0].(*mcp.TextContent).Text
	t.Logf("Response:\n%s", text)
}

func TestAcc_HandleListTags_WithSearch(t *testing.T) {
	client := newAcceptanceClient(t)

	h := &tags{service: client.Tags()}
	result, _, err := h.HandleListTags(context.Background(), nil, listTagsInput{
		Search:   "prod",
		PageSize: 5,
	})

	require.NoError(t, err)
	require.Len(t, result.Content, 1)

	text := result.Content[0].(*mcp.TextContent).Text
	t.Logf("Response (search 'prod'):\n%s", text)
}
