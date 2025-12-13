//go:build acc

package handle

import (
	"context"
	"fmt"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestAcc_HandleTagResource(t *testing.T) {
	client := newAcceptanceClient(t)
	ctx := context.Background()

	// First, get a tag ID from the list
	tagList, _, err := client.Tags().List(ctx, nil)
	require.NoError(t, err)
	require.NotEmpty(t, tagList, "no tags found in account")

	tagID := tagList[0].PK
	uri := fmt.Sprintf("%s%d", tagURIPrefix, tagID)

	h := &tags{service: client.Tags()}
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: uri,
		},
	}

	result, err := h.handleTagResource(ctx, req)
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

func TestAcc_HandleTagResource_InvalidID(t *testing.T) {
	client := newAcceptanceClient(t)
	ctx := context.Background()

	h := &tags{service: client.Tags()}
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: tagURIPrefix + "999999999",
		},
	}

	_, err := h.handleTagResource(ctx, req)
	require.Error(t, err)
}
