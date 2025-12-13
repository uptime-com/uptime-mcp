//go:build acc

package handle

import (
	"context"
	"fmt"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestAcc_HandleOutageResource(t *testing.T) {
	client := newAcceptanceClient(t)
	ctx := context.Background()

	// First, get an outage ID from the list
	outageList, _, err := client.Outages().List(ctx, nil)
	require.NoError(t, err)
	if len(outageList) == 0 {
		t.Skip("no outages found in account")
	}

	outageID := outageList[0].PK
	uri := fmt.Sprintf("%s%d", outageURIPrefix, outageID)

	h := &outages{service: client.Outages()}
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: uri,
		},
	}

	result, err := h.handleOutageResource(ctx, req)
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

func TestAcc_HandleOutageResource_InvalidID(t *testing.T) {
	client := newAcceptanceClient(t)
	ctx := context.Background()

	h := &outages{service: client.Outages()}
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: outageURIPrefix + "999999999",
		},
	}

	_, err := h.handleOutageResource(ctx, req)
	require.Error(t, err)
}
