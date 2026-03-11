package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerListStatusPageComponentsTool(srv *mcp.Server, h *statusPagesHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_status_page_components",
		Description: "List components for a status page",
	}, h.HandleListStatusPageComponents)
}

type listStatusPageComponentsInput struct {
	StatusPageID int64  `json:"status_page_id" jsonschema:"status page ID"`
	Search       string `json:"search,omitempty" jsonschema:"filter components by name"`
	Page         int64  `json:"page,omitempty" jsonschema:"page number, defaults to 1"`
	PageSize     int64  `json:"page_size,omitempty" jsonschema:"results per page, defaults to 25"`
}

func (h *statusPagesHandler) HandleListStatusPageComponents(ctx context.Context, _ *mcp.CallToolRequest, in listStatusPageComponentsInput) (*mcp.CallToolResult, any, error) {
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

	opts := upapi.StatusPageComponentListOptions{
		Search:   in.Search,
		Page:     in.Page,
		PageSize: in.PageSize,
	}

	result, err := client.StatusPages().Components(upapi.PrimaryKey(in.StatusPageID)).List(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list status page components: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(formatPaginationHeader(result.TotalCount, page, pageSize, len(result.Items)))
	for _, c := range result.Items {
		fmt.Fprintf(&sb, "- [%d] %s (status: %s)", c.PK, c.Name, c.Status)
		if c.IsGroup {
			sb.WriteString(" [group]")
		}
		sb.WriteString("\n")
	}

	return textResult(sb.String()), nil, nil
}
