package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateWHOISCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_whois_check",
		Description: "Create a new WHOIS domain expiry monitoring check. Monitors domain registration expiry dates. Locations are assigned automatically. Use list_contacts for contact group names.",
	}, h.HandleCreateWHOISCheck)
}

type createWHOISCheckInput struct {
	Name          string   `json:"name" jsonschema:"display name for the check"`
	Address       string   `json:"address" jsonschema:"domain name to monitor, e.g. example.com"`
	ContactGroups []string `json:"contact_groups" jsonschema:"contact group names to notify on alerts, use list_contacts tool to discover"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign, use create_tag to create new tags first"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	Threshold     int64    `json:"threshold,omitempty" jsonschema:"days before domain expiry to trigger an alert"`
	ExpectString  string   `json:"expect_string,omitempty" jsonschema:"expected string in WHOIS response, check fails if not found"`
}

func (c *checksHandler) HandleCreateWHOISCheck(ctx context.Context, _ *mcp.CallToolRequest, in createWHOISCheckInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" || in.Address == "" {
		return nil, nil, fmt.Errorf("name and address are required")
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

	created, err := client.Checks().CreateWHOIS(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create WHOIS check: %w", err)
	}

	return textResult(fmt.Sprintf("Created WHOIS check #%d: %s", created.PK, created.Name)), nil, nil
}
