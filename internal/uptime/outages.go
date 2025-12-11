package uptime

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api "github.com/uptime-com/uptime-client-go"
)

// ListOutagesInput defines parameters for listing outages.
type ListOutagesInput struct {
	Search   string `json:"search,omitempty" jsonschema:"description=Search term to filter outages"`
	Type     string `json:"type,omitempty" jsonschema:"description=Filter by check type"`
	Page     int    `json:"page,omitempty" jsonschema:"description=Page number (default 1)"`
	PageSize int    `json:"page_size,omitempty" jsonschema:"description=Results per page (default 25, max 100)"`
}

var listOutagesTool = &mcp.Tool{
	Name:        "list_outages",
	Description: "List outages across all monitored checks with optional filtering",
}

func (p *Provider) handleListOutages(ctx context.Context, _ *mcp.ServerSession, req *mcp.CallToolParamsFor[ListOutagesInput]) (*mcp.CallToolResultFor[any], error) {
	client, err := getClient(ctx)
	if err != nil {
		return errorResult(err), nil
	}

	in := req.Arguments
	opts := &api.OutageListOptions{
		Search:                     in.Search,
		CheckMonitoringServiceType: in.Type,
		Page:                       in.Page,
		PageSize:                   in.PageSize,
	}

	outages, _, err := client.Outages.List(ctx, opts)
	if err != nil {
		return errorResult(fmt.Errorf("failed to list outages: %w", err)), nil
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

	return textResult(sb.String()), nil
}

// GetOutageInput defines parameters for getting a single outage.
type GetOutageInput struct {
	ID string `json:"id" jsonschema:"description=Outage ID (pk)"`
}

var getOutageTool = &mcp.Tool{
	Name:        "get_outage",
	Description: "Get details of a specific outage including all alerts",
}

func (p *Provider) handleGetOutage(ctx context.Context, _ *mcp.ServerSession, req *mcp.CallToolParamsFor[GetOutageInput]) (*mcp.CallToolResultFor[any], error) {
	client, err := getClient(ctx)
	if err != nil {
		return errorResult(err), nil
	}

	if req.Arguments.ID == "" {
		return errorResult(fmt.Errorf("id is required")), nil
	}

	outage, _, err := client.Outages.Get(ctx, req.Arguments.ID)
	if err != nil {
		return errorResult(fmt.Errorf("failed to get outage: %w", err)), nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Outage #%d\n", outage.PK)
	fmt.Fprintf(&sb, "Check: %s (#%d)\n", outage.CheckName, outage.CheckPK)
	fmt.Fprintf(&sb, "Type: %s\n", outage.CheckMonitoringServiceType)
	fmt.Fprintf(&sb, "Address: %s\n", outage.CheckAddresss)
	fmt.Fprintf(&sb, "Created: %s\n", outage.CreatedAt.Format("2006-01-02 15:04:05"))

	if outage.StateIsUp {
		fmt.Fprintf(&sb, "Status: Resolved\n")
		fmt.Fprintf(&sb, "Resolved: %s\n", outage.ResolvedAt.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(&sb, "Duration: %d seconds\n", outage.DurationSecs)
	} else {
		fmt.Fprintf(&sb, "Status: Ongoing\n")
		fmt.Fprintf(&sb, "Locations down: %d\n", outage.NumLocationsDown)
	}

	if outage.Ignored {
		fmt.Fprintf(&sb, "Ignored: yes\n")
	}

	if outage.AllAlerts != nil && len(*outage.AllAlerts) > 0 {
		fmt.Fprintf(&sb, "\nAlerts:\n")
		for _, a := range *outage.AllAlerts {
			fmt.Fprintf(&sb, "  - %s (%s): %s\n", a.MonitoringServerName, a.Location, a.Output)
		}
	}

	return textResult(sb.String()), nil
}
