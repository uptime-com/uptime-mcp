package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerGetDashboardTool(srv *mcp.Server, h *dashboardsHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_dashboard",
		Description: "Get detailed information about a specific dashboard by ID",
	}, h.HandleGetDashboard)
}

type getDashboardInput struct {
	ID int64 `json:"id" jsonschema:"dashboard ID"`
}

func (h *dashboardsHandler) HandleGetDashboard(ctx context.Context, _ *mcp.CallToolRequest, in getDashboardInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	var sb strings.Builder
	if err := h.loadDashboard(ctx, client, in.ID, &sb); err != nil {
		return nil, nil, err
	}

	return textResult(sb.String()), nil, nil
}
