package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdateStatusPageComponentTool(srv *mcp.Server, h *statusPagesHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_status_page_component",
		Description: "Update an existing status page component",
	}, h.HandleUpdateStatusPageComponent)
}

type updateStatusPageComponentInput struct {
	StatusPageID int64  `json:"status_page_id" jsonschema:"status page ID"`
	ID           int64  `json:"id" jsonschema:"component ID"`
	Name         string `json:"name,omitempty" jsonschema:"component name"`
	Description  string `json:"description,omitempty" jsonschema:"component description"`
	GroupID      *int64 `json:"group_id,omitempty" jsonschema:"parent group ID"`
	ServiceID    *int64 `json:"service_id,omitempty" jsonschema:"linked service (check) ID"`
}

func (h *statusPagesHandler) HandleUpdateStatusPageComponent(ctx context.Context, _ *mcp.CallToolRequest, in updateStatusPageComponentInput) (*mcp.CallToolResult, any, error) {
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

	component := upapi.StatusPageComponent{
		Name:        in.Name,
		Description: in.Description,
		GroupID:     in.GroupID,
		ServiceID:   in.ServiceID,
	}

	updated, err := client.StatusPages().Components(upapi.PrimaryKey(in.StatusPageID)).Update(ctx, upapi.PrimaryKey(in.ID), component)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update status page component: %w", err)
	}

	return textResult(fmt.Sprintf("Updated component #%d: %s", updated.PK, updated.Name)), nil, nil
}
