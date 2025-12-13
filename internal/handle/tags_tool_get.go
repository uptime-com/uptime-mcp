package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerGetTagTool(srv *mcp.Server, h *tagsHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_tag",
		Description: "Get detailed information about a tag",
	}, h.HandleGetTag)
}

type getTagInput struct {
	ID int64 `json:"id"`
}

func (h *tagsHandler) HandleGetTag(ctx context.Context, _ *mcp.CallToolRequest, in getTagInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	var sb strings.Builder
	if err := h.loadTag(ctx, client, in.ID, &sb); err != nil {
		return nil, nil, err
	}

	return textResult(sb.String()), nil, nil
}
