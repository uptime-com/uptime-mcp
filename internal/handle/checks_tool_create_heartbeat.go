package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateHeartbeatCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_heartbeat_check",
		Description: "Create a new heartbeat monitoring check. The server generates a unique URL that your service must ping at the configured interval. No locations needed. Use list_contacts for contact group names.",
	}, h.HandleCreateHeartbeatCheck)
}

type createHeartbeatCheckInput struct {
	Name          string   `json:"name" jsonschema:"display name for the check"`
	Interval      int64    `json:"interval" jsonschema:"expected ping interval in minutes, alerts if no ping received within this period, defaults to 5"`
	ContactGroups []string `json:"contact_groups" jsonschema:"contact group names to notify on alerts, use list_contacts tool to discover"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign, use create_tag to create new tags first"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
}

func (c *checksHandler) HandleCreateHeartbeatCheck(ctx context.Context, _ *mcp.CallToolRequest, in createHeartbeatCheckInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}

	interval := in.Interval
	if interval == 0 {
		interval = 5
	}

	var contactGroups *[]string
	if len(in.ContactGroups) > 0 {
		contactGroups = &in.ContactGroups
	}

	check := upapi.CheckHeartbeat{
		Name:          in.Name,
		Interval:      interval,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Notes:         in.Notes,
	}

	created, err := client.Checks().CreateHeartbeat(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create heartbeat check: %w", err)
	}

	return textResult(fmt.Sprintf("Created heartbeat check #%d: %s\nHeartbeat URL: %s", created.PK, created.Name, created.HeartbeatURL)), nil, nil
}
