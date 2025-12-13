package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerDeleteCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "delete_check",
		Description: "Delete a monitoring check by ID",
	}, h.HandleDeleteCheck)
}

type deleteCheckInput struct {
	ID int `json:"id"`
}

func (c *checksHandler) HandleDeleteCheck(ctx context.Context, _ *mcp.CallToolRequest, in deleteCheckInput) (*mcp.CallToolResult, any, error) {
	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	_, err := c.service.Delete(ctx, in.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to delete check: %w", err)
	}

	return textResult(fmt.Sprintf("Successfully deleted check #%d", in.ID)), nil, nil
}
