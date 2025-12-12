package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/fx"
)

var GetTagToolModule = fx.Module("tool.get_tag",
	fx.Invoke(func(srv *mcp.Server) {
		mcp.AddTool(srv, &mcp.Tool{
			Name:        "get_tag",
			Description: "Get details of a specific tag by ID",
		}, HandleGetTag)
	}),
)

type getTagInput struct {
	ID int `json:"id"`
}

func HandleGetTag(ctx context.Context, _ *mcp.CallToolRequest, in getTagInput) (*mcp.CallToolResult, any, error) {
	client, err := getClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	tag, _, err := client.Tags.Get(ctx, in.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get tag: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Tag #%d\n", tag.PK)
	fmt.Fprintf(&sb, "Name: %s\n", tag.Tag)
	fmt.Fprintf(&sb, "Color: #%s\n", tag.ColorHex)

	return textResult(sb.String()), nil, nil
}
