package handle

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

const alertURIPrefix = "uptime://alerts/"

func registerAlertResource(srv *mcp.Server, h *alertsHandler) {
	srv.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: alertURIPrefix + "{id}",
		Name:        "alert",
		Description: "Uptime.com alert details including location, output, and resolution status",
		MIMEType:    "text/plain",
	}, h.handleAlertResource)
}

func (h *alertsHandler) handleAlertResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	uri := req.Params.URI

	idStr := strings.TrimPrefix(uri, alertURIPrefix)
	if idStr == uri {
		return nil, fmt.Errorf("invalid alert URI: %s", uri)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid alert ID: %s", idStr)
	}

	var sb strings.Builder
	if err := h.loadAlert(ctx, client, id, &sb); err != nil {
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

func (h *alertsHandler) loadAlert(ctx context.Context, client upapi.API, id int64, sb *strings.Builder) error {
	alert, err := client.Alerts().Get(ctx, upapi.PrimaryKey(id))
	if err != nil {
		return fmt.Errorf("failed to get alert: %w", err)
	}

	fmt.Fprintf(sb, "Alert #%d\n", alert.PK)
	fmt.Fprintf(sb, "Check: %s (#%d)\n", alert.CheckName, alert.CheckPK)
	fmt.Fprintf(sb, "Type: %s\n", alert.CheckMonitoringServiceType)
	fmt.Fprintf(sb, "Address: %s\n", alert.CheckAddress)

	if alert.StateIsUp {
		fmt.Fprintf(sb, "Status: Resolved\n")
	} else {
		fmt.Fprintf(sb, "Status: Down\n")
		fmt.Fprintf(sb, "Locations down: %d\n", alert.NumLocationsDown)
	}
	if alert.Ignored {
		fmt.Fprintf(sb, "Ignored: yes\n")
	}
	if alert.CreatedAt != nil {
		fmt.Fprintf(sb, "Created: %s\n", alert.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	writeLocationAlerts(sb, alert.AllAlerts)

	return nil
}

// writeLocationAlerts writes one line per location reporting the check down,
// and nothing when there are none.
func writeLocationAlerts(sb *strings.Builder, alerts []upapi.Alert) {
	if len(alerts) == 0 {
		return
	}
	fmt.Fprintf(sb, "\nAlerts:\n")
	for _, a := range alerts {
		fmt.Fprintf(sb, "  - %s (%s): %s\n", a.MonitoringServerName, a.Location, a.Output)
	}
}
