package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateStatusPageTool(srv *mcp.Server, h *statusPagesHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_status_page",
		Description: "Create a new status page",
	}, h.HandleCreateStatusPage)
}

type createStatusPageInput struct {
	Name               string `json:"name" jsonschema:"status page name"`
	PageType           string `json:"page_type,omitempty" jsonschema:"page type, valid: PUBLIC PUBLIC_SLA INTERNAL"`
	Slug               string `json:"slug,omitempty" jsonschema:"URL slug"`
	Description        string `json:"description,omitempty" jsonschema:"page description"`
	VisibilityLevel    string `json:"visibility_level,omitempty" jsonschema:"visibility level, valid: PUBLIC UPTIME_USERS EXTERNAL_USERS"`
	AllowSubscriptions bool   `json:"allow_subscriptions,omitempty" jsonschema:"allow subscriptions"`
	Timezone           string `json:"timezone,omitempty" jsonschema:"timezone"`
	CNAME              string `json:"cname,omitempty" jsonschema:"custom CNAME"`
	ContactEmail       string `json:"contact_email,omitempty" jsonschema:"contact email"`
}

func (h *statusPagesHandler) HandleCreateStatusPage(ctx context.Context, _ *mcp.CallToolRequest, in createStatusPageInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}

	pageType := in.PageType
	if pageType == "" {
		pageType = "PUBLIC"
	}

	visibilityLevel := in.VisibilityLevel
	if visibilityLevel == "" {
		visibilityLevel = "UPTIME_USERS"
	}

	sp := upapi.StatusPage{
		Name:                     in.Name,
		PageType:                 pageType,
		Slug:                     in.Slug,
		Description:              in.Description,
		VisibilityLevel:          visibilityLevel,
		AllowSubscriptions:       optBool(in.AllowSubscriptions),
		Timezone:                 in.Timezone,
		CNAME:                    in.CNAME,
		ContactEmail:             in.ContactEmail,
		UptimeCalculationType:    "BY_INCIDENTS",
		Theme:                    "LEGACY",
		CustomHeaderBgColorHex:   "#002E52",
		CustomHeaderTextColorHex: "#FFFFFF",
	}

	created, err := client.StatusPages().Create(ctx, sp)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create status page: %w", err)
	}

	return textResult(fmt.Sprintf("Created status page #%d: %s", created.PK, created.Name)), nil, nil
}
