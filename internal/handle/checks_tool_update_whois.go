package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdateWHOISCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_whois_check",
		Description: "Update an existing WHOIS domain expiry monitoring check by ID. Only provided fields are changed.",
	}, h.HandleUpdateWHOISCheck)
}

type updateWHOISCheckInput struct {
	ID            int64    `json:"id" jsonschema:"check ID"`
	Name          string   `json:"name,omitempty" jsonschema:"display name for the check"`
	Address       string   `json:"address,omitempty" jsonschema:"domain name to monitor"`
	ContactGroups []string `json:"contact_groups,omitempty" jsonschema:"contact group names to notify on alerts"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	Threshold     int64    `json:"threshold,omitempty" jsonschema:"days before domain expiry to trigger an alert"`
	ExpectString  string   `json:"expect_string,omitempty" jsonschema:"expected string in WHOIS response"`
}

func (c *checksHandler) HandleUpdateWHOISCheck(ctx context.Context, _ *mcp.CallToolRequest, in updateWHOISCheckInput) (*mcp.CallToolResult, any, error) {
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

	check := upapi.CheckWHOIS{
		Name:          in.Name,
		Address:       in.Address,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Threshold:     in.Threshold,
		ExpectString:  in.ExpectString,
		Notes:         in.Notes,
	}

	updated, err := client.Checks().UpdateWHOIS(ctx, upapi.PrimaryKey(in.ID), check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update WHOIS check: %w", err)
	}

	return textResult(fmt.Sprintf("Updated WHOIS check #%d: %s", updated.PK, updated.Name)), nil, nil
}
