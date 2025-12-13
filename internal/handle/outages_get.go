package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerGetOutageTool(srv *mcp.Server, h *outages) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_outage",
		Description: "Get details of a specific outage including all alerts",
	}, h.HandleGetOutage)
}

type getOutageInput struct {
	ID string `json:"id"`
}

func (o *outages) HandleGetOutage(ctx context.Context, _ *mcp.CallToolRequest, in getOutageInput) (*mcp.CallToolResult, any, error) {
	if in.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	outage, _, err := o.service.Get(ctx, in.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get outage: %w", err)
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

	return textResult(sb.String()), nil, nil
}
