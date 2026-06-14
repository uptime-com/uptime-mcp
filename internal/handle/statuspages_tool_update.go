package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdateStatusPageTool(srv *mcp.Server, h *statusPagesHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_status_page",
		Description: "Update an existing status page",
	}, h.HandleUpdateStatusPage)
}

type updateStatusPageInput struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name,omitempty"`
	Description        string `json:"description,omitempty"`
	VisibilityLevel    string `json:"visibility_level,omitempty"`
	AllowSubscriptions *bool  `json:"allow_subscriptions,omitempty"`
	Timezone           string `json:"timezone,omitempty"`
	CNAME              string `json:"cname,omitempty"`
	ContactEmail       string `json:"contact_email,omitempty"`
}

func (h *statusPagesHandler) HandleUpdateStatusPage(ctx context.Context, _ *mcp.CallToolRequest, in updateStatusPageInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	sp := upapi.StatusPage{
		Name:            in.Name,
		Description:     in.Description,
		VisibilityLevel: in.VisibilityLevel,
		Timezone:        in.Timezone,
		CNAME:           in.CNAME,
		ContactEmail:    in.ContactEmail,
	}
	if in.AllowSubscriptions != nil {
		sp.AllowSubscriptions = in.AllowSubscriptions
	}

	updated, err := client.StatusPages().Update(ctx, upapi.PrimaryKey(in.ID), sp)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update status page: %w", err)
	}

	return textResult(fmt.Sprintf("Updated status page #%d: %s", updated.PK, updated.Name)), nil, nil
}
