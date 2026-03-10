package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdateRUMCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_rum_check",
		Description: "Update an existing Real User Monitoring check by ID. Only provided fields are changed.",
	}, h.HandleUpdateRUMCheck)
}

type updateRUMCheckInput struct {
	ID            int64    `json:"id" jsonschema:"check ID"`
	Name          string   `json:"name,omitempty" jsonschema:"display name for the check"`
	Address       string   `json:"address,omitempty" jsonschema:"website URL to monitor"`
	ContactGroups []string `json:"contact_groups,omitempty" jsonschema:"contact group names to notify on alerts"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	Threshold     int64    `json:"threshold,omitempty" jsonschema:"performance threshold in milliseconds"`
}

func (c *checksHandler) HandleUpdateRUMCheck(ctx context.Context, _ *mcp.CallToolRequest, in updateRUMCheckInput) (*mcp.CallToolResult, any, error) {
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

	check := upapi.CheckRUM{
		Name:          in.Name,
		Address:       in.Address,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Threshold:     in.Threshold,
		Notes:         in.Notes,
	}

	updated, err := client.Checks().UpdateRUM(ctx, upapi.PrimaryKey(in.ID), check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update RUM check: %w", err)
	}

	return textResult(fmt.Sprintf("Updated RUM check #%d: %s", updated.PK, updated.Name)), nil, nil
}
