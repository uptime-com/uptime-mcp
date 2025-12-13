package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const outageURIPrefix = "https://uptime.com/api/v1/outages/"

func registerOutageResource(srv *mcp.Server, h *outages) {
	srv.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: outageURIPrefix + "{id}",
		Name:        "outage",
		Description: "Uptime.com outage details",
		MIMEType:    "text/plain",
	}, h.handleOutageResource)
}

func (o *outages) handleOutageResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	uri := req.Params.URI

	id := strings.TrimPrefix(uri, outageURIPrefix)
	if id == uri || id == "" {
		return nil, fmt.Errorf("invalid outage URI: %s", uri)
	}

	outage, _, err := o.service.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get outage: %w", err)
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

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      uri,
			MIMEType: "text/plain",
			Text:     sb.String(),
		}},
	}, nil
}
