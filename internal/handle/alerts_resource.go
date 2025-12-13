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

	// Location info
	fmt.Fprintf(sb, "Location: %s\n", alert.Location)
	fmt.Fprintf(sb, "Server: %s\n", alert.MonitoringServerName)
	if alert.MonitoringServerIPv4 != nil {
		fmt.Fprintf(sb, "IPv4: %s\n", alert.MonitoringServerIPv4.String())
	}
	if alert.MonitoringServerIPv6 != nil {
		fmt.Fprintf(sb, "IPv6: %s\n", alert.MonitoringServerIPv6.String())
	}

	// Status
	if alert.StateIsUp {
		fmt.Fprintf(sb, "Status: Resolved\n")
	} else {
		fmt.Fprintf(sb, "Status: Down\n")
	}
	if alert.Ignored {
		fmt.Fprintf(sb, "Ignored: yes\n")
	}

	// Timestamps
	if alert.CreatedAt != nil {
		fmt.Fprintf(sb, "Created: %s\n", alert.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	if alert.ResolvedAt != nil {
		fmt.Fprintf(sb, "Resolved: %s\n", alert.ResolvedAt.Format("2006-01-02 15:04:05"))
	}

	// Output
	if alert.Output != "" {
		fmt.Fprintf(sb, "\nOutput:\n%s\n", alert.Output)
	}

	return nil
}
