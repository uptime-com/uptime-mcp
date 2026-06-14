package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdateStatusPageIncidentTool(srv *mcp.Server, h *statusPagesHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_status_page_incident",
		Description: "Update an existing status page incident",
	}, h.HandleUpdateStatusPageIncident)
}

type updateStatusPageIncidentInput struct {
	StatusPageID int64  `json:"status_page_id" jsonschema:"status page ID"`
	ID           int64  `json:"id" jsonschema:"incident ID"`
	Name         string `json:"name,omitempty" jsonschema:"incident name"`
	StartsAt     string `json:"starts_at,omitempty" jsonschema:"start time"`
	EndsAt       string `json:"ends_at,omitempty" jsonschema:"end time"`
}

func (h *statusPagesHandler) HandleUpdateStatusPageIncident(ctx context.Context, _ *mcp.CallToolRequest, in updateStatusPageIncidentInput) (*mcp.CallToolResult, any, error) {
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

	incident := upapi.StatusPageIncident{
		Name:     in.Name,
		StartsAt: in.StartsAt,
		EndsAt:   in.EndsAt,
	}

	updated, err := client.StatusPages().Incidents(upapi.PrimaryKey(in.StatusPageID)).Update(ctx, upapi.PrimaryKey(in.ID), incident)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update status page incident: %w", err)
	}

	return textResult(fmt.Sprintf("Updated incident #%d: %s", updated.PK, updated.Name)), nil, nil
}
