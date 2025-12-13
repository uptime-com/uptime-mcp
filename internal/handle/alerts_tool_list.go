package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerListAlertsTool(srv *mcp.Server, h *alertsHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_alerts",
		Description: "List alerts (incidents from monitoring locations) with optional filtering by check, type, state, or date range",
	}, h.HandleListAlerts)
}

type listAlertsInput struct {
	CheckID   int64  `json:"check_id,omitempty"`
	Type      string `json:"type,omitempty"`
	Tag       string `json:"tag,omitempty"`
	Resolved  *bool  `json:"resolved,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Search    string `json:"search,omitempty"`
	Page      int64  `json:"page,omitempty"`
	PageSize  int64  `json:"page_size,omitempty"`
}

func (h *alertsHandler) HandleListAlerts(ctx context.Context, _ *mcp.CallToolRequest, in listAlertsInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	pageSize := in.PageSize
	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	page := in.Page
	if page == 0 {
		page = 1
	}

	opts := upapi.AlertListOptions{
		CheckPK:                    in.CheckID,
		CheckMonitoringServiceType: in.Type,
		CheckTag:                   in.Tag,
		StateIsUp:                  in.Resolved,
		StartDate:                  in.StartDate,
		EndDate:                    in.EndDate,
		Search:                     in.Search,
		Page:                       page,
		PageSize:                   pageSize,
	}

	result, err := client.Alerts().List(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list alerts: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(formatPaginationHeader(result.TotalCount, page, pageSize, len(result.Items)))
	for _, alert := range result.Items {
		status := "down"
		if alert.StateIsUp {
			status = "resolved"
		}
		if alert.Ignored {
			status += " (ignored)"
		}
		fmt.Fprintf(&sb, "- [%d] %s (%s) - %s\n", alert.PK, alert.CheckName, alert.CheckMonitoringServiceType, status)
		fmt.Fprintf(&sb, "  Location: %s (%s)\n", alert.Location, alert.MonitoringServerName)
		if alert.CreatedAt != nil {
			fmt.Fprintf(&sb, "  Created: %s\n", alert.CreatedAt.Format("2006-01-02 15:04:05"))
		}
		if alert.StateIsUp && alert.ResolvedAt != nil {
			fmt.Fprintf(&sb, "  Resolved: %s\n", alert.ResolvedAt.Format("2006-01-02 15:04:05"))
		}
	}

	return textResult(sb.String()), nil, nil
}
