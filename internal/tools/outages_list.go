package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api "github.com/uptime-com/uptime-client-go"
	"go.uber.org/fx"
)

var ListOutagesToolModule = fx.Module("tool.list_outages",
	fx.Invoke(func(srv *mcp.Server) {
		mcp.AddTool(srv, &mcp.Tool{
			Name:        "list_outages",
			Description: "List outages across all monitored checks with optional filtering",
		}, HandleListOutages)
	}),
)

type listOutagesInput struct {
	Search   string `json:"search,omitempty"`
	Type     string `json:"type,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

func HandleListOutages(ctx context.Context, _ *mcp.CallToolRequest, in listOutagesInput) (*mcp.CallToolResult, any, error) {
	client, err := getClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	opts := &api.OutageListOptions{
		Search:                     in.Search,
		CheckMonitoringServiceType: in.Type,
		Page:                       in.Page,
		PageSize:                   in.PageSize,
	}

	outages, _, err := client.Outages.List(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list outages: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d outages:\n\n", len(outages))
	for _, o := range outages {
		status := "ongoing"
		if o.StateIsUp {
			status = "resolved"
		}
		fmt.Fprintf(&sb, "- [%d] %s (%s) - %s\n", o.PK, o.CheckName, o.CheckMonitoringServiceType, status)
		fmt.Fprintf(&sb, "  Created: %s\n", o.CreatedAt.Format("2006-01-02 15:04:05"))
		if o.StateIsUp {
			fmt.Fprintf(&sb, "  Resolved: %s (duration: %d sec)\n", o.ResolvedAt.Format("2006-01-02 15:04:05"), o.DurationSecs)
		}
	}

	return textResult(sb.String()), nil, nil
}
