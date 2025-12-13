package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerDeleteTagTool(srv *mcp.Server, h *tagsHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "delete_tag",
		Description: "Delete a check tag by ID",
	}, h.HandleDeleteTag)
}

type deleteTagInput struct {
	ID int64 `json:"id"`
}

func (t *tagsHandler) HandleDeleteTag(ctx context.Context, _ *mcp.CallToolRequest, in deleteTagInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	err = client.Tags().Delete(ctx, upapi.PrimaryKey(in.ID))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to delete tag: %w", err)
	}

	return textResult(fmt.Sprintf("Successfully deleted tag #%d", in.ID)), nil, nil
}
