//go:build acc

package handle

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestAcc_HandleListChecks(t *testing.T) {
	client := newAcceptanceClient(t)

	h := &checksHandler{service: client.Checks()}
	result, _, err := h.HandleListChecks(context.Background(), nil, listChecksInput{
		PageSize: 5,
	})

	require.NoError(t, err)
	require.Len(t, result.Content, 1)

	text := result.Content[0].(*mcp.TextContent).Text
	t.Logf("Response:\n%s", text)
}

func TestAcc_HandleListChecks_WithFilter(t *testing.T) {
	client := newAcceptanceClient(t)

	h := &checksHandler{service: client.Checks()}
	result, _, err := h.HandleListChecks(context.Background(), nil, listChecksInput{
		Type:     "HTTP",
		PageSize: 3,
	})

	require.NoError(t, err)
	require.Len(t, result.Content, 1)

	text := result.Content[0].(*mcp.TextContent).Text
	t.Logf("Response (HTTP only):\n%s", text)
}
