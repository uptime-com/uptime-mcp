package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerGetCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_check",
		Description: "Get detailed information about a monitoring check including configuration and current status",
	}, h.HandleGetCheck)
}

type getCheckInput struct {
	ID int64 `json:"id" jsonschema:"check ID as returned by list_checks"`
}

func (h *checksHandler) HandleGetCheck(ctx context.Context, _ *mcp.CallToolRequest, in getCheckInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	var sb strings.Builder
	if err := h.loadCheck(ctx, client, in.ID, &sb); err != nil {
		return nil, nil, err
	}

	return textResult(sb.String()), nil, nil
}
