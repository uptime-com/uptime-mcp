package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerGetStatusPageIncidentTool(srv *mcp.Server, h *statusPagesHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_status_page_incident",
		Description: "Get detailed information about a status page incident",
	}, h.HandleGetStatusPageIncident)
}

type getStatusPageIncidentInput struct {
	StatusPageID int64 `json:"status_page_id" jsonschema:"status page ID"`
	ID           int64 `json:"id" jsonschema:"incident ID"`
}

func (h *statusPagesHandler) HandleGetStatusPageIncident(ctx context.Context, _ *mcp.CallToolRequest, in getStatusPageIncidentInput) (*mcp.CallToolResult, any, error) {
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

	incident, err := client.StatusPages().Incidents(upapi.PrimaryKey(in.StatusPageID)).Get(ctx, upapi.PrimaryKey(in.ID))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get status page incident: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Incident #%d\n", incident.PK)
	fmt.Fprintf(&sb, "Name: %s\n", incident.Name)
	fmt.Fprintf(&sb, "IncidentType: %s\n", incident.IncidentType)
	fmt.Fprintf(&sb, "StartsAt: %s\n", incident.StartsAt)
	fmt.Fprintf(&sb, "EndsAt: %s\n", incident.EndsAt)
	fmt.Fprintf(&sb, "Status: %s\n", incident.Status)

	if len(incident.Updates) > 0 {
		sb.WriteString("\nUpdates:\n")
		for _, u := range incident.Updates {
			fmt.Fprintf(&sb, "- [%s] %s: %s\n", u.UpdatedAt, u.IncidentState, u.Description)
		}
	}

	if len(incident.AffectedComponents) > 0 {
		sb.WriteString("\nAffected Components:\n")
		for _, ac := range incident.AffectedComponents {
			fmt.Fprintf(&sb, "- Component #%d (status: %s)\n", ac.Component.PK, ac.Status)
		}
	}

	return textResult(sb.String()), nil, nil
}
