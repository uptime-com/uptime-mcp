package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerGetStatusPageTool(srv *mcp.Server, h *statusPagesHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_status_page",
		Description: "Get detailed information about a status page",
	}, h.HandleGetStatusPage)
}

type getStatusPageInput struct {
	ID int64 `json:"id"`
}

func (h *statusPagesHandler) HandleGetStatusPage(ctx context.Context, _ *mcp.CallToolRequest, in getStatusPageInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	var sb strings.Builder
	if err := h.loadStatusPage(ctx, client, in.ID, &sb); err != nil {
		return nil, nil, err
	}

	return textResult(sb.String()), nil, nil
}
