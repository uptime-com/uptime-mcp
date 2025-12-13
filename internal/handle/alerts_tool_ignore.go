package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerIgnoreAlertTool(srv *mcp.Server, h *alertsHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "ignore_alert",
		Description: "Toggle the ignored state of an alert. Ignored alerts are excluded from outage calculations and reports.",
	}, h.HandleIgnoreAlert)
}

type ignoreAlertInput struct {
	ID int64 `json:"id"`
}

func (h *alertsHandler) HandleIgnoreAlert(ctx context.Context, _ *mcp.CallToolRequest, in ignoreAlertInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	alert, err := client.Alerts().Ignore(ctx, upapi.PrimaryKey(in.ID))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to toggle alert ignore state: %w", err)
	}

	status := "not ignored"
	if alert.Ignored {
		status = "ignored"
	}

	return textResult(fmt.Sprintf("Alert #%d is now %s", alert.PK, status)), nil, nil
}
