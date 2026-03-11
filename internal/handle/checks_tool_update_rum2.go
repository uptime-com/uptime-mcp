package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdateRUM2CheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_rum2_check",
		Description: "Update an existing Real User Monitoring v2 check by ID. Only provided fields are changed.",
	}, h.HandleUpdateRUM2Check)
}

type updateRUM2CheckInput struct {
	ID            int64    `json:"id" jsonschema:"check ID"`
	Name          string   `json:"name,omitempty" jsonschema:"display name for the check"`
	Address       string   `json:"address,omitempty" jsonschema:"website URL to monitor"`
	ContactGroups []string `json:"contact_groups,omitempty" jsonschema:"contact group names to notify on alerts"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	Threshold     int64    `json:"threshold,omitempty" jsonschema:"performance threshold in milliseconds"`
}

func (c *checksHandler) HandleUpdateRUM2Check(ctx context.Context, _ *mcp.CallToolRequest, in updateRUM2CheckInput) (*mcp.CallToolResult, any, error) {
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

	check := upapi.CheckRUM2{
		Name:          in.Name,
		Address:       in.Address,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Threshold:     in.Threshold,
		Notes:         in.Notes,
	}

	updated, err := client.Checks().UpdateRUM2(ctx, upapi.PrimaryKey(in.ID), check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update RUM2 check: %w", err)
	}

	return textResult(fmt.Sprintf("Updated RUM2 check #%d: %s", updated.PK, updated.Name)), nil, nil
}
