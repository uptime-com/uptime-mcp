package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerListDashboardsTool(srv *mcp.Server, h *dashboardsHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_dashboards",
		Description: "List all dashboards with optional search filtering",
	}, h.HandleListDashboards)
}

type listDashboardsInput struct {
	Search   string `json:"search,omitempty" jsonschema:"filter dashboards by name"`
	Page     int64  `json:"page,omitempty" jsonschema:"page number, defaults to 1"`
	PageSize int64  `json:"page_size,omitempty" jsonschema:"results per page, defaults to 25"`
}

func (h *dashboardsHandler) HandleListDashboards(ctx context.Context, _ *mcp.CallToolRequest, in listDashboardsInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	pageSize := in.PageSize
	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	page := in.Page
	if page == 0 {
		page = 1
	}

	opts := upapi.DashboardListOptions{
		Search:   in.Search,
		Page:     in.Page,
		PageSize: in.PageSize,
	}

	result, err := client.Dashboards().List(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list dashboards: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(formatPaginationHeader(result.TotalCount, page, pageSize, len(result.Items)))
	for _, d := range result.Items {
		fmt.Fprintf(&sb, "- [%d] %s", d.PK, d.Name)
		if d.IsPinned {
			sb.WriteString(" (pinned)")
		}
		sb.WriteString("\n")
	}

	return textResult(sb.String()), nil, nil
}
