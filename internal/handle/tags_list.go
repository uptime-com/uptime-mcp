package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerListTagsTool(srv *mcp.Server, h *tags) {
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

func (t *tags) HandleListTags(ctx context.Context, _ *mcp.CallToolRequest, in listTagsInput) (*mcp.CallToolResult, any, error) {
	opts := upapi.TagListOptions{
		Search:   in.Search,
		Page:     in.Page,
		PageSize: in.PageSize,
	}

	tagList, err := t.service.List(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list tags: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d tags:\n\n", len(tagList))
	for _, tag := range tagList {
		fmt.Fprintf(&sb, "- [%d] %s (color: #%s)\n", tag.PK, tag.Tag, tag.ColorHex)
	}

	return textResult(sb.String()), nil, nil
}
