package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerGetAlertTool(srv *mcp.Server, h *alertsHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_alert",
		Description: "Get detailed information about an alert including location, output, and resolution status",
	}, h.HandleGetAlert)
}

type getAlertInput struct {
	ID int64 `json:"id" jsonschema:"alert ID as returned by list_alerts"`
}

func (h *alertsHandler) HandleGetAlert(ctx context.Context, _ *mcp.CallToolRequest, in getAlertInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	var sb strings.Builder
	if err := h.loadAlert(ctx, client, in.ID, &sb); err != nil {
		return nil, nil, err
	}

	return textResult(sb.String()), nil, nil
}
