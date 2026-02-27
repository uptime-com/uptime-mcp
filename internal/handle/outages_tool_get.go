package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerGetOutageTool(srv *mcp.Server, h *outagesHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_outage",
		Description: "Get detailed information about an outage including alerts and resolution status",
	}, h.HandleGetOutage)
}

type getOutageInput struct {
	ID int64 `json:"id" jsonschema:"outage ID as returned by list_outages"`
}

func (h *outagesHandler) HandleGetOutage(ctx context.Context, _ *mcp.CallToolRequest, in getOutageInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	var sb strings.Builder
	if err := h.loadOutage(ctx, client, in.ID, &sb); err != nil {
		return nil, nil, err
	}

	return textResult(sb.String()), nil, nil
}
