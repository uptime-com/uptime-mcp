package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api "github.com/uptime-com/uptime-client-go"
	"go.uber.org/fx"
)

var CreateTagToolModule = fx.Module("tool.create_tag",
	fx.Invoke(func(srv *mcp.Server) {
		mcp.AddTool(srv, &mcp.Tool{
			Name:        "create_tag",
			Description: "Create a new check tag",
		}, HandleCreateTag)
	}),
)

type createTagInput struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

func HandleCreateTag(ctx context.Context, _ *mcp.CallToolRequest, in createTagInput) (*mcp.CallToolResult, any, error) {
	client, err := getClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}

	tag := &api.Tag{
		Tag:      in.Name,
		ColorHex: in.Color,
	}

	created, _, err := client.Tags.Create(ctx, tag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return textResult(fmt.Sprintf("Created tag #%d: %s", created.PK, created.Tag)), nil, nil
}
