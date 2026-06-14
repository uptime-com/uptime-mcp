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
		Description: "Get historical statistics for a check including uptime percentage, response times, and outage counts. Accepts optional start_date, end_date (YYYY-MM-DD), and location filters.",
	}, h.HandleGetCheckStats)
}

type getCheckStatsInput struct {
	ID        int64  `json:"id"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Location  string `json:"location,omitempty"`
}

func (c *checksHandler) HandleGetCheckStats(ctx context.Context, _ *mcp.CallToolRequest, in getCheckStatsInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	opts := upapi.CheckStatsOptions{
		StartDate: in.StartDate,
		EndDate:   in.EndDate,
		Location:  in.Location,
	}

	stats, err := client.Checks().Stats(ctx, upapi.PrimaryKey(in.ID), opts)
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
	for _, s := range stats.Items {
		totalOutages += s.Outages
		totalDowntime += s.DowntimeSecs
	}

	fmt.Fprintf(&sb, "Totals:\n")
	fmt.Fprintf(&sb, "  Outages: %d\n", totalOutages)
	fmt.Fprintf(&sb, "  Downtime: %d seconds\n\n", totalDowntime)

	if len(stats.Items) > 0 {
		fmt.Fprintf(&sb, "Daily breakdown:\n")
		for _, s := range stats.Items {
			var parts []string
			if s.Uptime != nil {
				parts = append(parts, fmt.Sprintf("%.2f%% uptime", *s.Uptime*100))
			}
			if s.ResponseTime != nil {
				parts = append(parts, fmt.Sprintf("%.0fms response", *s.ResponseTime))
			}
			if s.Outages > 0 {
				parts = append(parts, fmt.Sprintf("%d outage(s), %ds down", s.Outages, s.DowntimeSecs))
			} else {
				parts = append(parts, "0 outages")
			}
			fmt.Fprintf(&sb, "  %s: %s\n", s.Date, strings.Join(parts, ", "))
		}
	}

	return textResult(sb.String()), nil, nil
}
