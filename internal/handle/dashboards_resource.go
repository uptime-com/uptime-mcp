package handle

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

const dashboardURIPrefix = "uptime://dashboards/"

func registerDashboardResource(srv *mcp.Server, h *dashboardsHandler) {
	srv.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: dashboardURIPrefix + "{id}",
		Name:        "dashboard",
		Description: "Uptime.com dashboard details",
		MIMEType:    "text/plain",
	}, h.handleDashboardResource)
}

func (h *dashboardsHandler) handleDashboardResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	uri := req.Params.URI

	idStr := strings.TrimPrefix(uri, dashboardURIPrefix)
	if idStr == uri {
		return nil, fmt.Errorf("invalid dashboard URI: %s", uri)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid dashboard ID: %s", idStr)
	}

	var sb strings.Builder
	if err := h.loadDashboard(ctx, client, id, &sb); err != nil {
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

func (h *dashboardsHandler) loadDashboard(ctx context.Context, client upapi.API, id int64, sb *strings.Builder) error {
	d, err := client.Dashboards().Get(ctx, upapi.PrimaryKey(id))
	if err != nil {
		return fmt.Errorf("failed to get dashboard: %w", err)
	}

	fmt.Fprintf(sb, "Dashboard #%d\n", d.PK)
	fmt.Fprintf(sb, "Name: %s\n", d.Name)
	fmt.Fprintf(sb, "Pinned: %t\n", d.IsPinned)
	fmt.Fprintf(sb, "Ordering: %d\n", d.Ordering)

	if len(d.ServicesSelected) > 0 {
		fmt.Fprintf(sb, "Selected services: %s\n", strings.Join(d.ServicesSelected, ", "))
	}
	if len(d.ServicesTags) > 0 {
		fmt.Fprintf(sb, "Service tags: %s\n", strings.Join(d.ServicesTags, ", "))
	}

	fmt.Fprintf(sb, "\nServices section: show=%t, count=%d, include_up=%t, include_down=%t, include_paused=%t, include_maintenance=%t\n",
		d.ServicesShowSection, d.ServicesNumToShow, d.ServicesIncludeUp, d.ServicesIncludeDown, d.ServicesIncludePaused, d.ServicesIncludeMaintenance)
	fmt.Fprintf(sb, "Services sort: primary=%s, secondary=%s\n", d.ServicesPrimarySort, d.ServicesSecondarySort)
	fmt.Fprintf(sb, "Services display: uptime=%t, response_time=%t\n", d.ServicesShowUptime, d.ServicesShowResponseTime)

	fmt.Fprintf(sb, "\nMetrics section: show=%t, all_checks=%t\n", d.MetricsShowSection, d.MetricsForAllChecks)

	fmt.Fprintf(sb, "\nAlerts section: show=%t, all_checks=%t, include_ignored=%t, include_resolved=%t, count=%d\n",
		d.AlertsShowSection, d.AlertsForAllChecks, d.AlertsIncludeIgnored, d.AlertsincludeResolved, d.AlertsnumToShow)

	return nil
}
