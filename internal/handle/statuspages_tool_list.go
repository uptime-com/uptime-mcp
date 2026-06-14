package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerListStatusPagesTool(srv *mcp.Server, h *statusPagesHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_status_pages",
		Description: "List all status pages with optional search filtering",
	}, h.HandleListStatusPages)
}

type listStatusPagesInput struct {
	Search   string `json:"search,omitempty" jsonschema:"filter status pages by name"`
	Page     int64  `json:"page,omitempty" jsonschema:"page number, defaults to 1"`
	PageSize int64  `json:"page_size,omitempty" jsonschema:"results per page, defaults to 25"`
}

func (h *statusPagesHandler) HandleListStatusPages(ctx context.Context, _ *mcp.CallToolRequest, in listStatusPagesInput) (*mcp.CallToolResult, any, error) {
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

	opts := upapi.StatusPageListOptions{
		Search:   in.Search,
		Page:     in.Page,
		PageSize: in.PageSize,
	}

	result, err := client.StatusPages().List(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list status pages: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(formatPaginationHeader(result.TotalCount, page, pageSize, len(result.Items)))
	for _, sp := range result.Items {
		fmt.Fprintf(&sb, "- [%d] %s (%s)\n", sp.PK, sp.Name, sp.PageType)
	}

	return textResult(sb.String()), nil, nil
}
