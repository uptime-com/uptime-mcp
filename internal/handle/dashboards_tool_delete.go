package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerDeleteDashboardTool(srv *mcp.Server, h *dashboardsHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "delete_dashboard",
		Description: "Delete a dashboard by ID",
	}, h.HandleDeleteDashboard)
}

type deleteDashboardInput struct {
	ID int64 `json:"id" jsonschema:"dashboard ID"`
}

func (h *dashboardsHandler) HandleDeleteDashboard(ctx context.Context, _ *mcp.CallToolRequest, in deleteDashboardInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	err = client.Dashboards().Delete(ctx, upapi.PrimaryKey(in.ID))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to delete dashboard: %w", err)
	}

	return textResult(fmt.Sprintf("Successfully deleted dashboard #%d", in.ID)), nil, nil
}
