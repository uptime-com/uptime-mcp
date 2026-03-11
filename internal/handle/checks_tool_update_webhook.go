package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdateWebhookCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_webhook_check",
		Description: "Update an existing webhook monitoring check by ID. Only provided fields are changed.",
	}, h.HandleUpdateWebhookCheck)
}

type updateWebhookCheckInput struct {
	ID            int64    `json:"id" jsonschema:"check ID"`
	Name          string   `json:"name,omitempty" jsonschema:"display name for the check"`
	ContactGroups []string `json:"contact_groups,omitempty" jsonschema:"contact group names to notify on alerts"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
}

func (c *checksHandler) HandleUpdateWebhookCheck(ctx context.Context, _ *mcp.CallToolRequest, in updateWebhookCheckInput) (*mcp.CallToolResult, any, error) {
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

	check := upapi.CheckWebhook{
		Name:          in.Name,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Notes:         in.Notes,
	}

	updated, err := client.Checks().UpdateWebhook(ctx, upapi.PrimaryKey(in.ID), check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update webhook check: %w", err)
	}

	return textResult(fmt.Sprintf("Updated webhook check #%d: %s", updated.PK, updated.Name)), nil, nil
}
