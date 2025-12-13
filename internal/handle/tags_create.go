package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateTagTool(srv *mcp.Server, h *tags) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_tag",
		Description: "Create a new check tag",
	}, h.HandleCreateTag)
}

type createTagInput struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

func (t *tags) HandleCreateTag(ctx context.Context, _ *mcp.CallToolRequest, in createTagInput) (*mcp.CallToolResult, any, error) {
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}

	tag := upapi.Tag{
		Tag:      in.Name,
		ColorHex: in.Color,
	}

	created, err := t.service.Create(ctx, tag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return textResult(fmt.Sprintf("Created tag #%d: %s", created.PK, created.Tag)), nil, nil
}
