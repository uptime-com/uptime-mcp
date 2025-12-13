package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerListOutagesTool(srv *mcp.Server, h *outagesHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_outages",
		Description: "List outages across all monitored checks with optional filtering",
	}, h.HandleListOutages)
}

type listOutagesInput struct {
	Search   string `json:"search,omitempty"`
	Type     string `json:"type,omitempty"`
	Page     int64  `json:"page,omitempty"`
	PageSize int64  `json:"page_size,omitempty"`
}

func (o *outagesHandler) HandleListOutages(ctx context.Context, _ *mcp.CallToolRequest, in listOutagesInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	opts := upapi.OutageListOptions{
		Search:                     in.Search,
		CheckMonitoringServiceType: in.Type,
		Page:                       in.Page,
		PageSize:                   in.PageSize,
	}

	outageList, err := client.Outages().List(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list outages: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d outages:\n\n", len(outageList))
	for _, outage := range outageList {
		status := "ongoing"
		if outage.StateIsUp {
			status = "resolved"
		}
		fmt.Fprintf(&sb, "- [%d] %s (%s) - %s\n", outage.PK, outage.CheckName, outage.CheckMonitoringServiceType, status)
		fmt.Fprintf(&sb, "  Created: %s\n", outage.CreatedAt.Format("2006-01-02 15:04:05"))
		if outage.StateIsUp {
			fmt.Fprintf(&sb, "  Resolved: %s (duration: %d sec)\n", outage.ResolvedAt.Format("2006-01-02 15:04:05"), outage.DurationSecs)
		}
	}

	return textResult(sb.String()), nil, nil
}
