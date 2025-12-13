package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api "github.com/uptime-com/uptime-client-go"
	"go.uber.org/fx"
)

var ListTagsToolModule = fx.Module("tool.list_tags",
	fx.Invoke(func(srv *mcp.Server, t *tags) {
		mcp.AddTool(srv, &mcp.Tool{
			Name:        "list_tags",
			Description: "List all check tags with optional search filtering",
		}, t.HandleListTags)
	}),
)

type listTagsInput struct {
	Search   string `json:"search,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

func (t *tags) HandleListTags(ctx context.Context, _ *mcp.CallToolRequest, in listTagsInput) (*mcp.CallToolResult, any, error) {
	opts := &api.TagListOptions{
		Search:   in.Search,
		Page:     in.Page,
		PageSize: in.PageSize,
	}

	tagList, _, err := t.service.List(ctx, opts)
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
