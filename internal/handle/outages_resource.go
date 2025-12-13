package handle

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

const outageURIPrefix = "uptime://outages/"

func registerOutageResource(srv *mcp.Server, h *outagesHandler) {
	srv.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: outageURIPrefix + "{id}",
		Name:        "outage",
		Description: "Uptime.com outage details",
		MIMEType:    "text/plain",
	}, h.handleOutageResource)
}

func (h *outagesHandler) handleOutageResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	uri := req.Params.URI

	idStr := strings.TrimPrefix(uri, outageURIPrefix)
	if idStr == uri || idStr == "" {
		return nil, fmt.Errorf("invalid outage URI: %s", uri)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid outage ID: %s", idStr)
	}

	var sb strings.Builder
	if err := h.loadOutage(ctx, client, id, &sb); err != nil {
		return nil, err
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      uri,
			MIMEType: "text/plain",
			Text:     sb.String(),
		}},
	}, nil
}

func (h *outagesHandler) loadOutage(ctx context.Context, client upapi.API, id int64, sb *strings.Builder) error {
	outage, err := client.Outages().Get(ctx, upapi.PrimaryKey(id))
	if err != nil {
		return fmt.Errorf("failed to get outage: %w", err)
	}

	fmt.Fprintf(sb, "Outage #%d\n", outage.PK)
	fmt.Fprintf(sb, "Check: %s (#%d)\n", outage.CheckName, outage.CheckPK)
	fmt.Fprintf(sb, "Type: %s\n", outage.CheckMonitoringServiceType)
	fmt.Fprintf(sb, "Address: %s\n", outage.CheckAddress)
	fmt.Fprintf(sb, "Created: %s\n", outage.CreatedAt.Format("2006-01-02 15:04:05"))

	if outage.StateIsUp {
		fmt.Fprintf(sb, "Status: Resolved\n")
		fmt.Fprintf(sb, "Resolved: %s\n", outage.ResolvedAt.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(sb, "Duration: %d seconds\n", outage.DurationSecs)
	} else {
		fmt.Fprintf(sb, "Status: Ongoing\n")
		fmt.Fprintf(sb, "Locations down: %d\n", outage.NumLocationsDown)
	}

	if outage.Ignored {
		fmt.Fprintf(sb, "Ignored: yes\n")
	}

	if outage.AllAlerts != nil && len(*outage.AllAlerts) > 0 {
		fmt.Fprintf(sb, "\nAlerts:\n")
		for _, a := range *outage.AllAlerts {
			fmt.Fprintf(sb, "  - %s (%s): %s\n", a.MonitoringServerName, a.Location, a.Output)
		}
	}

	return nil
}
