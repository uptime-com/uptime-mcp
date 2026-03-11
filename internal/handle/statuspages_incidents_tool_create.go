package handle

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateStatusPageIncidentTool(srv *mcp.Server, h *statusPagesHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_status_page_incident",
		Description: "Create a new incident on a status page",
	}, h.HandleCreateStatusPageIncident)
}

type createStatusPageIncidentInput struct {
	StatusPageID          int64  `json:"status_page_id" jsonschema:"status page ID"`
	Name                  string `json:"name" jsonschema:"incident name"`
	IncidentType          string `json:"incident_type,omitempty" jsonschema:"incident type, valid: INCIDENT SCHEDULED_MAINTENANCE"`
	StartsAt              string `json:"starts_at,omitempty" jsonschema:"start time in ISO 8601 format (YYYY-MM-DDThh:mm:ssZ), defaults to now"`
	EndsAt                string `json:"ends_at,omitempty" jsonschema:"end time in ISO 8601 format"`
	UpdateComponentStatus bool   `json:"update_component_status,omitempty" jsonschema:"whether to update component status"`
	NotifySubscribers     bool   `json:"notify_subscribers,omitempty" jsonschema:"whether to notify subscribers"`
	Description           string `json:"description,omitempty" jsonschema:"initial update description"`
}

func (h *statusPagesHandler) HandleCreateStatusPageIncident(ctx context.Context, _ *mcp.CallToolRequest, in createStatusPageIncidentInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.StatusPageID == 0 {
		return nil, nil, fmt.Errorf("status_page_id is required")
	}
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}

	incidentType := in.IncidentType
	if incidentType == "" {
		incidentType = "INCIDENT"
	}

	now := time.Now().UTC().Format(time.RFC3339)

	startsAt := in.StartsAt
	if startsAt == "" {
		startsAt = now
	}

	incident := upapi.StatusPageIncident{
		Name:                  in.Name,
		IncidentType:          incidentType,
		StartsAt:              startsAt,
		EndsAt:                in.EndsAt,
		UpdateComponentStatus: in.UpdateComponentStatus,
		NotifySubscribers:     in.NotifySubscribers,
	}

	if in.Description != "" {
		incident.Updates = []upapi.IncidentUpdate{
			{
				IncidentState: "investigating",
				Description:   in.Description,
				UpdatedAt:     now,
			},
		}
	}

	created, err := client.StatusPages().Incidents(upapi.PrimaryKey(in.StatusPageID)).Create(ctx, incident)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create status page incident: %w", err)
	}

	return textResult(fmt.Sprintf("Created incident #%d: %s", created.PK, created.Name)), nil, nil
}
