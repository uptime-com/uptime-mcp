package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdateHeartbeatCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_heartbeat_check",
		Description: "Update an existing heartbeat monitoring check by ID. Only provided fields are changed.",
	}, h.HandleUpdateHeartbeatCheck)
}

type updateHeartbeatCheckInput struct {
	ID            int64    `json:"id" jsonschema:"check ID"`
	Name          string   `json:"name,omitempty" jsonschema:"display name for the check"`
	Interval      int64    `json:"interval,omitempty" jsonschema:"expected ping interval in minutes"`
	ContactGroups []string `json:"contact_groups,omitempty" jsonschema:"contact group names to notify on alerts"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
}

func (c *checksHandler) HandleUpdateHeartbeatCheck(ctx context.Context, _ *mcp.CallToolRequest, in updateHeartbeatCheckInput) (*mcp.CallToolResult, any, error) {
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

	check := upapi.CheckHeartbeat{
		Name:          in.Name,
		Interval:      in.Interval,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Notes:         in.Notes,
	}

	updated, err := client.Checks().UpdateHeartbeat(ctx, upapi.PrimaryKey(in.ID), check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update heartbeat check: %w", err)
	}

	return textResult(fmt.Sprintf("Updated heartbeat check #%d: %s", updated.PK, updated.Name)), nil, nil
}
