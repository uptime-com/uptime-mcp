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
	fx.Invoke(func(srv *mcp.Server) {
		mcp.AddTool(srv, &mcp.Tool{
			Name:        "list_tags",
			Description: "List all check tags with optional search filtering",
		}, HandleListTags)
	}),
)

type listTagsInput struct {
	Search   string `json:"search,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

func HandleListTags(ctx context.Context, _ *mcp.CallToolRequest, in listTagsInput) (*mcp.CallToolResult, any, error) {
	client, err := getClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	opts := &api.TagListOptions{
		Search:   in.Search,
		Page:     in.Page,
		PageSize: in.PageSize,
	}

	tags, _, err := client.Tags.List(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list tags: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d tags:\n\n", len(tags))
	for _, t := range tags {
		fmt.Fprintf(&sb, "- [%d] %s (color: #%s)\n", t.PK, t.Tag, t.ColorHex)
	}

	return textResult(sb.String()), nil, nil
}
