package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateWebhookCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_webhook_check",
		Description: "Create a new webhook monitoring check. The server generates a unique URL that receives status updates from external systems. No locations needed. Use list_contacts for contact group names.",
	}, h.HandleCreateWebhookCheck)
}

type createWebhookCheckInput struct {
	Name          string   `json:"name" jsonschema:"display name for the check"`
	ContactGroups []string `json:"contact_groups" jsonschema:"contact group names to notify on alerts, use list_contacts tool to discover"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign, use create_tag to create new tags first"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
}

func (c *checksHandler) HandleCreateWebhookCheck(ctx context.Context, _ *mcp.CallToolRequest, in createWebhookCheckInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
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

	created, err := client.Checks().CreateWebhook(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create webhook check: %w", err)
	}

	return textResult(fmt.Sprintf("Created webhook check #%d: %s\nWebhook URL: %s", created.PK, created.Name, created.WebhookURL)), nil, nil
}
