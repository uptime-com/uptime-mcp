package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerDeleteStatusPageIncidentTool(srv *mcp.Server, h *statusPagesHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "delete_status_page_incident",
		Description: "Delete a status page incident by ID",
	}, h.HandleDeleteStatusPageIncident)
}

type deleteStatusPageIncidentInput struct {
	StatusPageID int64 `json:"status_page_id" jsonschema:"status page ID"`
	ID           int64 `json:"id" jsonschema:"incident ID"`
}

func (h *statusPagesHandler) HandleDeleteStatusPageIncident(ctx context.Context, _ *mcp.CallToolRequest, in deleteStatusPageIncidentInput) (*mcp.CallToolResult, any, error) {
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

	err = client.StatusPages().Incidents(upapi.PrimaryKey(in.StatusPageID)).Delete(ctx, upapi.PrimaryKey(in.ID))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to delete status page incident: %w", err)
	}

	return textResult(fmt.Sprintf("Successfully deleted incident #%d", in.ID)), nil, nil
}
