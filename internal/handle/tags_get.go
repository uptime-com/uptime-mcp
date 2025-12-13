package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerGetTagTool(srv *mcp.Server, h *tags) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_tag",
		Description: "Get details of a specific tag by ID",
	}, h.HandleGetTag)
}

type getTagInput struct {
	ID int `json:"id"`
}

func (t *tags) HandleGetTag(ctx context.Context, _ *mcp.CallToolRequest, in getTagInput) (*mcp.CallToolResult, any, error) {
	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	tag, _, err := t.service.Get(ctx, in.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get tag: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Tag #%d\n", tag.PK)
	fmt.Fprintf(&sb, "Name: %s\n", tag.Tag)
	fmt.Fprintf(&sb, "Color: #%s\n", tag.ColorHex)

	return textResult(sb.String()), nil, nil
}
