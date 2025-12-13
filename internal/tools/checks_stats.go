package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api "github.com/uptime-com/uptime-client-go"
	"go.uber.org/fx"
)

var GetCheckStatsToolModule = fx.Module("tool.get_check_stats",
	fx.Invoke(func(srv *mcp.Server, c *checksHandler) {
		mcp.AddTool(srv, &mcp.Tool{
			Name:        "get_check_stats",
			Description: "Get statistics for a monitoring check including uptime percentage and outages",
		}, c.HandleGetCheckStats)
	}),
)

type getCheckStatsInput struct {
	ID        int    `json:"id"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

func (c *checksHandler) HandleGetCheckStats(ctx context.Context, _ *mcp.CallToolRequest, in getCheckStatsInput) (*mcp.CallToolResult, any, error) {
	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	opts := &api.CheckStatsOptions{
		StartDate: in.StartDate,
		EndDate:   in.EndDate,
	}

	stats, _, err := c.service.Stats(ctx, in.ID, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get check stats: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Statistics for check #%d\n", in.ID)
	fmt.Fprintf(&sb, "Period: %s to %s\n\n", stats.StartDate, stats.EndDate)

	fmt.Fprintf(&sb, "Totals:\n")
	fmt.Fprintf(&sb, "  Outages: %d\n", stats.Totals.Outages)
	fmt.Fprintf(&sb, "  Downtime: %d seconds\n\n", stats.Totals.DowntimeSecs)

	if len(stats.Statistics) > 0 {
		fmt.Fprintf(&sb, "Daily breakdown:\n")
		for _, s := range stats.Statistics {
			fmt.Fprintf(&sb, "  %s: %d outages, %d sec downtime\n", s.Date, s.Outages, s.DowntimeSecs)
		}
	}

	return textResult(sb.String()), nil, nil
}
