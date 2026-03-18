package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdateCloudStatusCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_cloudstatus_check",
		Description: "Update an existing Cloud Status monitoring check by ID. Only provided fields are changed.",
	}, h.HandleUpdateCloudStatusCheck)
}

type updateCloudStatusCheckInput struct {
	ID            int64    `json:"id" jsonschema:"check ID"`
	Name          string   `json:"name,omitempty" jsonschema:"display name for the check"`
	ServiceName   string   `json:"service_name,omitempty" jsonschema:"cloud provider service to monitor"`
	Locations     []string `json:"locations,omitempty" jsonschema:"probe location identifiers"`
	ContactGroups []string `json:"contact_groups,omitempty" jsonschema:"contact group names to notify on alerts"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign"`
	IsPaused      *bool    `json:"is_paused,omitempty" jsonschema:"whether the check is paused"`
}

func (c *checksHandler) HandleUpdateCloudStatusCheck(ctx context.Context, _ *mcp.CallToolRequest, in updateCloudStatusCheckInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	var contactGroups *[]string
	if len(in.ContactGroups) > 0 {
		contactGroups = &in.ContactGroups
	}

	check := upapi.CheckCloudStatus{
		Name:          in.Name,
		Locations:     in.Locations,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		IsPaused:      in.IsPaused,
		CloudStatusConfig: upapi.CheckCloudStatusConfig{
			ServiceName: in.ServiceName,
		},
	}

	updated, err := client.Checks().UpdateCloudStatus(ctx, upapi.PrimaryKey(in.ID), check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update Cloud Status check: %w", err)
	}

	return textResult(fmt.Sprintf("Updated Cloud Status check #%d: %s", updated.PK, updated.Name)), nil, nil
}
