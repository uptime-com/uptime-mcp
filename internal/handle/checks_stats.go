package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerGetCheckStatsTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_check_stats",
		Description: "Get statistics for a monitoring check including uptime percentage and outages",
	}, h.HandleGetCheckStats)
}

type getCheckStatsInput struct {
	ID        int64  `json:"id"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

func (c *checksHandler) HandleGetCheckStats(ctx context.Context, _ *mcp.CallToolRequest, in getCheckStatsInput) (*mcp.CallToolResult, any, error) {
	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	opts := upapi.CheckStatsOptions{
		StartDate: in.StartDate,
		EndDate:   in.EndDate,
	}

	stats, err := c.service.Stats(ctx, upapi.PrimaryKey(in.ID), opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get check stats: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Statistics for check #%d\n", in.ID)
	if in.StartDate != "" && in.EndDate != "" {
		fmt.Fprintf(&sb, "Period: %s to %s\n\n", in.StartDate, in.EndDate)
	}

	// Calculate totals from statistics
	var totalOutages, totalDowntime int64
	for _, s := range stats {
		totalOutages += s.Outages
		totalDowntime += s.DowntimeSecs
	}

	fmt.Fprintf(&sb, "Totals:\n")
	fmt.Fprintf(&sb, "  Outages: %d\n", totalOutages)
	fmt.Fprintf(&sb, "  Downtime: %d seconds\n\n", totalDowntime)

	if len(stats) > 0 {
		fmt.Fprintf(&sb, "Daily breakdown:\n")
		for _, s := range stats {
			fmt.Fprintf(&sb, "  %s: %d outages, %d sec downtime\n", s.Date, s.Outages, s.DowntimeSecs)
		}
	}

	return textResult(sb.String()), nil, nil
}
