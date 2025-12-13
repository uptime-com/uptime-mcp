package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerListTagsTool(srv *mcp.Server, h *tagsHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_tags",
		Description: "List all check tags with optional search filtering",
	}, h.HandleListTags)
}

type listTagsInput struct {
	Search   string `json:"search,omitempty"`
	Page     int64  `json:"page,omitempty"`
	PageSize int64  `json:"page_size,omitempty"`
}

func (t *tagsHandler) HandleListTags(ctx context.Context, _ *mcp.CallToolRequest, in listTagsInput) (*mcp.CallToolResult, any, error) {
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

	opts := upapi.TagListOptions{
		Search:   in.Search,
		Page:     in.Page,
		PageSize: in.PageSize,
	}

	result, err := client.Tags().List(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list tags: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(formatPaginationHeader(result.TotalCount, page, pageSize, len(result.Items)))
	for _, tag := range result.Items {
		fmt.Fprintf(&sb, "- [%d] %s (color: #%s)\n", tag.PK, tag.Tag, tag.ColorHex)
	}

	return textResult(sb.String()), nil, nil
}
