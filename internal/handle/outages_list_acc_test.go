//go:build acc

package handle

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestAcc_HandleListOutages(t *testing.T) {
	client := newAcceptanceClient(t)

	h := &outages{service: client.Outages()}
	result, _, err := h.HandleListOutages(context.Background(), nil, listOutagesInput{
		PageSize: 10,
	})

	require.NoError(t, err)
	require.Len(t, result.Content, 1)

	text := result.Content[0].(*mcp.TextContent).Text
	t.Logf("Response:\n%s", text)
}

func TestAcc_HandleListOutages_WithTypeFilter(t *testing.T) {
	client := newAcceptanceClient(t)

	h := &outages{service: client.Outages()}
	result, _, err := h.HandleListOutages(context.Background(), nil, listOutagesInput{
		Type:     "HTTP",
		PageSize: 5,
	})

	require.NoError(t, err)
	require.Len(t, result.Content, 1)

	text := result.Content[0].(*mcp.TextContent).Text
	t.Logf("Response (HTTP only):\n%s", text)
}
