package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerDeleteStatusPageTool(srv *mcp.Server, h *statusPagesHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "delete_status_page",
		Description: "Delete a status page by ID",
	}, h.HandleDeleteStatusPage)
}

type deleteStatusPageInput struct {
	ID int64 `json:"id"`
}

func (h *statusPagesHandler) HandleDeleteStatusPage(ctx context.Context, _ *mcp.CallToolRequest, in deleteStatusPageInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	err = client.StatusPages().Delete(ctx, upapi.PrimaryKey(in.ID))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to delete status page: %w", err)
	}

	return textResult(fmt.Sprintf("Successfully deleted status page #%d", in.ID)), nil, nil
}
