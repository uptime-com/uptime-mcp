package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerDeleteTagTool(srv *mcp.Server, h *tags) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "delete_tag",
		Description: "Delete a check tag by ID",
	}, h.HandleDeleteTag)
}

type deleteTagInput struct {
	ID int `json:"id"`
}

func (t *tags) HandleDeleteTag(ctx context.Context, _ *mcp.CallToolRequest, in deleteTagInput) (*mcp.CallToolResult, any, error) {
	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	_, err := t.service.Delete(ctx, in.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to delete tag: %w", err)
	}

	return textResult(fmt.Sprintf("Successfully deleted tag #%d", in.ID)), nil, nil
}
