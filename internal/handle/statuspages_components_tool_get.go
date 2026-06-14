package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerGetStatusPageComponentTool(srv *mcp.Server, h *statusPagesHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_status_page_component",
		Description: "Get detailed information about a status page component",
	}, h.HandleGetStatusPageComponent)
}

type getStatusPageComponentInput struct {
	StatusPageID int64 `json:"status_page_id" jsonschema:"status page ID"`
	ID           int64 `json:"id" jsonschema:"component ID"`
}

func (h *statusPagesHandler) HandleGetStatusPageComponent(ctx context.Context, _ *mcp.CallToolRequest, in getStatusPageComponentInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.StatusPageID == 0 {
		return nil, nil, fmt.Errorf("status_page_id is required")
	}
	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	component, err := client.StatusPages().Components(upapi.PrimaryKey(in.StatusPageID)).Get(ctx, upapi.PrimaryKey(in.ID))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get status page component: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Component #%d\n", component.PK)
	fmt.Fprintf(&sb, "Name: %s\n", component.Name)
	fmt.Fprintf(&sb, "Description: %s\n", component.Description)
	fmt.Fprintf(&sb, "Status: %s\n", component.Status)
	fmt.Fprintf(&sb, "IsGroup: %t\n", component.IsGroup)
	if component.GroupID != nil {
		fmt.Fprintf(&sb, "GroupID: %d\n", *component.GroupID)
	}
	if component.ServiceID != nil {
		fmt.Fprintf(&sb, "ServiceID: %d\n", *component.ServiceID)
	}

	return textResult(sb.String()), nil, nil
}
