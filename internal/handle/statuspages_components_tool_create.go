package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateStatusPageComponentTool(srv *mcp.Server, h *statusPagesHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_status_page_component",
		Description: "Create a new component on a status page",
	}, h.HandleCreateStatusPageComponent)
}

type createStatusPageComponentInput struct {
	StatusPageID int64  `json:"status_page_id" jsonschema:"status page ID"`
	Name         string `json:"name" jsonschema:"component name"`
	Description  string `json:"description,omitempty" jsonschema:"component description"`
	IsGroup      bool   `json:"is_group,omitempty" jsonschema:"whether this is a group component"`
	GroupID      *int64 `json:"group_id,omitempty" jsonschema:"parent group ID"`
	ServiceID    *int64 `json:"service_id,omitempty" jsonschema:"linked service (check) ID"`
}

func (h *statusPagesHandler) HandleCreateStatusPageComponent(ctx context.Context, _ *mcp.CallToolRequest, in createStatusPageComponentInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.StatusPageID == 0 {
		return nil, nil, fmt.Errorf("status_page_id is required")
	}
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}

	component := upapi.StatusPageComponent{
		Name:        in.Name,
		Description: in.Description,
		IsGroup:     in.IsGroup,
		GroupID:     in.GroupID,
		ServiceID:   in.ServiceID,
	}

	created, err := client.StatusPages().Components(upapi.PrimaryKey(in.StatusPageID)).Create(ctx, component)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create status page component: %w", err)
	}

	return textResult(fmt.Sprintf("Created component #%d: %s", created.PK, created.Name)), nil, nil
}
