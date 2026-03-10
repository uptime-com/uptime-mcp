package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerDeleteStatusPageComponentTool(srv *mcp.Server, h *statusPagesHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "delete_status_page_component",
		Description: "Delete a status page component by ID",
	}, h.HandleDeleteStatusPageComponent)
}

type deleteStatusPageComponentInput struct {
	StatusPageID int64 `json:"status_page_id" jsonschema:"status page ID"`
	ID           int64 `json:"id" jsonschema:"component ID"`
}

func (h *statusPagesHandler) HandleDeleteStatusPageComponent(ctx context.Context, _ *mcp.CallToolRequest, in deleteStatusPageComponentInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.StatusPageID == 0 {
		return nil, nil, fmt.Errorf("status_page_id is required")
	}
	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	err = client.StatusPages().Components(upapi.PrimaryKey(in.StatusPageID)).Delete(ctx, upapi.PrimaryKey(in.ID))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to delete status page component: %w", err)
	}

	return textResult(fmt.Sprintf("Successfully deleted component #%d", in.ID)), nil, nil
}
