//go:build acc

package handle

import (
	"context"
	"fmt"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestAcc_HandleCheckResource(t *testing.T) {
	client := newAcceptanceClient(t)
	ctx := context.Background()

	// First, get a check ID from the list
	checks, _, err := client.Checks().List(ctx, nil)
	require.NoError(t, err)
	require.NotEmpty(t, checks, "no checks found in account")

	checkID := checks[0].PK
	uri := fmt.Sprintf("%s%d", checkURIPrefix, checkID)

	h := &checksHandler{service: client.Checks()}
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: uri,
		},
	}

	result, err := h.handleCheckResource(ctx, req)
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

func TestAcc_HandleCheckResource_InvalidID(t *testing.T) {
	client := newAcceptanceClient(t)
	ctx := context.Background()

	h := &checksHandler{service: client.Checks()}
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: checkURIPrefix + "999999999",
		},
	}

	_, err := h.handleCheckResource(ctx, req)
	require.Error(t, err)
}
