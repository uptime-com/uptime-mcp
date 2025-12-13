package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api "github.com/uptime-com/uptime-client-go"
	"go.uber.org/fx"
)

var UpdateTagToolModule = fx.Module("tool.update_tag",
	fx.Invoke(func(srv *mcp.Server, t *tags) {
		mcp.AddTool(srv, &mcp.Tool{
			Name:        "update_tag",
			Description: "Update an existing check tag",
		}, t.HandleUpdateTag)
	}),
)

type updateTagInput struct {
	ID    int    `json:"id"`
	Name  string `json:"name,omitempty"`
	Color string `json:"color,omitempty"`
}

func (t *tags) HandleUpdateTag(ctx context.Context, _ *mcp.CallToolRequest, in updateTagInput) (*mcp.CallToolResult, any, error) {
	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}
	if in.Name == "" && in.Color == "" {
		return nil, nil, fmt.Errorf("at least one of name or color is required")
	}

	tag := &api.Tag{
		PK:       in.ID,
		Tag:      in.Name,
		ColorHex: in.Color,
	}

	updated, _, err := t.service.Update(ctx, tag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update tag: %w", err)
	}

	return textResult(fmt.Sprintf("Updated tag #%d: %s", updated.PK, updated.Tag)), nil, nil
}
