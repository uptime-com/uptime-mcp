package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateRUMCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_rum_check",
		Description: "Create a new Real User Monitoring check. Requires placing a JavaScript snippet on your website to collect real user performance data. Locations are assigned automatically. Use list_contacts for contact group names.",
	}, h.HandleCreateRUMCheck)
}

type createRUMCheckInput struct {
	Name          string   `json:"name" jsonschema:"display name for the check"`
	Address       string   `json:"address" jsonschema:"website URL to monitor, e.g. https://example.com"`
	ContactGroups []string `json:"contact_groups" jsonschema:"contact group names to notify on alerts, use list_contacts tool to discover"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign, use create_tag to create new tags first"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	Threshold     int64    `json:"threshold,omitempty" jsonschema:"performance threshold in milliseconds"`
}

func (c *checksHandler) HandleCreateRUMCheck(ctx context.Context, _ *mcp.CallToolRequest, in createRUMCheckInput) (*mcp.CallToolResult, any, error) {
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

	check := upapi.CheckRUM{
		Name:          in.Name,
		Address:       in.Address,
		Locations:     []string{"AUTO"},
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Threshold:     in.Threshold,
		Notes:         in.Notes,
	}

	created, err := client.Checks().CreateRUM(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create RUM check: %w", err)
	}

	return textResult(fmt.Sprintf("Created RUM check #%d: %s", created.PK, created.Name)), nil, nil
}
