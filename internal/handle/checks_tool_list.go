package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

const defaultPageSize int64 = 25

func registerListChecksTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_checks",
		Description: "List monitoring checks with optional filtering. Returns paginated results (default 25 per page).",
	}, h.HandleListChecks)
}

// listChecksInput defines parameters for listing checksHandler.
type listChecksInput struct {
	Search   string `json:"search,omitempty"`
	Tag      string `json:"tag,omitempty"`
	Type     string `json:"type,omitempty"`
	IsPaused bool   `json:"is_paused,omitempty"`
	Page     int64  `json:"page,omitempty"`
	PageSize int64  `json:"page_size,omitempty"`
}

func (c *checksHandler) HandleListChecks(ctx context.Context, _ *mcp.CallToolRequest, in listChecksInput) (*mcp.CallToolResult, any, error) {
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

	opts := upapi.CheckListOptions{
		Search:                in.Search,
		MonitoringServiceType: in.Type,
		IsPaused:              in.IsPaused,
		Page:                  in.Page,
		PageSize:              in.PageSize,
	}
	if in.Tag != "" {
		opts.Tag = []string{in.Tag}
	}

	result, err := client.Checks().List(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list checks: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(formatPaginationHeader(result.TotalCount, page, pageSize, len(result.Items)))
	for _, ch := range result.Items {
		fmt.Fprintf(&sb, "- [%d] %s (%s) - %s\n", ch.PK, ch.Name, ch.CheckType, ch.Address)
	}

	return textResult(sb.String()), nil, nil
}
