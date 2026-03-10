package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerListStatusPageIncidentsTool(srv *mcp.Server, h *statusPagesHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_status_page_incidents",
		Description: "List incidents for a status page",
	}, h.HandleListStatusPageIncidents)
}

type listStatusPageIncidentsInput struct {
	StatusPageID int64  `json:"status_page_id" jsonschema:"status page ID"`
	Search       string `json:"search,omitempty" jsonschema:"filter incidents by name"`
	Page         int64  `json:"page,omitempty" jsonschema:"page number, defaults to 1"`
	PageSize     int64  `json:"page_size,omitempty" jsonschema:"results per page, defaults to 25"`
}

func (h *statusPagesHandler) HandleListStatusPageIncidents(ctx context.Context, _ *mcp.CallToolRequest, in listStatusPageIncidentsInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.StatusPageID == 0 {
		return nil, nil, fmt.Errorf("status_page_id is required")
	}

	pageSize := in.PageSize
	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	page := in.Page
	if page == 0 {
		page = 1
	}

	opts := upapi.StatusPageIncidentListOptions{
		Search:   in.Search,
		Page:     in.Page,
		PageSize: in.PageSize,
	}

	result, err := client.StatusPages().Incidents(upapi.PrimaryKey(in.StatusPageID)).List(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list status page incidents: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(formatPaginationHeader(result.TotalCount, page, pageSize, len(result.Items)))
	for _, inc := range result.Items {
		fmt.Fprintf(&sb, "- [%d] %s (type: %s, starts: %s)\n", inc.PK, inc.Name, inc.IncidentType, inc.StartsAt)
	}

	return textResult(sb.String()), nil, nil
}
