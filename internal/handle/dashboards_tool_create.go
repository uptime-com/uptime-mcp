package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateDashboardTool(srv *mcp.Server, h *dashboardsHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_dashboard",
		Description: "Create a new dashboard. Filter checks by tag names (services_tags) or specific check IDs (services_selected).",
	}, h.HandleCreateDashboard)
}

type createDashboardInput struct {
	Name                       string   `json:"name" jsonschema:"dashboard name"`
	ServicesSelected           []string `json:"services_selected,omitempty" jsonschema:"check IDs to include on the dashboard"`
	ServicesTags               []string `json:"services_tags,omitempty" jsonschema:"tag names to filter checks, use list_tags to discover"`
	IsPinned                   bool     `json:"is_pinned,omitempty" jsonschema:"pin dashboard to top of list"`
	ServicesShowSection        bool     `json:"services_show_section,omitempty" jsonschema:"show services section"`
	ServicesNumToShow          int64    `json:"services_num_to_show,omitempty" jsonschema:"number of services to display"`
	ServicesIncludeUp          bool     `json:"services_include_up,omitempty" jsonschema:"include checks with UP status"`
	ServicesIncludeDown        bool     `json:"services_include_down,omitempty" jsonschema:"include checks with DOWN status"`
	ServicesIncludePaused      bool     `json:"services_include_paused,omitempty" jsonschema:"include paused checks"`
	ServicesIncludeMaintenance bool     `json:"services_include_maintenance,omitempty" jsonschema:"include checks in maintenance"`
	MetricsShowSection         bool     `json:"metrics_show_section,omitempty" jsonschema:"show metrics section"`
	MetricsForAllChecks        bool     `json:"metrics_for_all_checks,omitempty" jsonschema:"show metrics for all checks"`
	AlertsShowSection          bool     `json:"alerts_show_section,omitempty" jsonschema:"show alerts section"`
	AlertsForAllChecks         bool     `json:"alerts_for_all_checks,omitempty" jsonschema:"show alerts for all checks"`
	AlertsNumToShow            int64    `json:"alerts_num_to_show,omitempty" jsonschema:"number of alerts to display, defaults to 10"`
	ServicesPrimarySort        string   `json:"services_primary_sort,omitempty" jsonschema:"primary sort for services, e.g. name_asc, name_desc, status, defaults to name_asc"`
	ServicesSecondarySort      string   `json:"services_secondary_sort,omitempty" jsonschema:"secondary sort for services, e.g. name_asc, status, defaults to status"`
}

func (h *dashboardsHandler) HandleCreateDashboard(ctx context.Context, _ *mcp.CallToolRequest, in createDashboardInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}

	servicesSelected := in.ServicesSelected
	if servicesSelected == nil {
		servicesSelected = []string{}
	}

	servicesNumToShow := in.ServicesNumToShow
	if servicesNumToShow == 0 {
		servicesNumToShow = 10
	}

	alertsNumToShow := in.AlertsNumToShow
	if alertsNumToShow == 0 {
		alertsNumToShow = 10
	}

	servicesPrimarySort := in.ServicesPrimarySort
	if servicesPrimarySort == "" {
		servicesPrimarySort = "name_asc"
	}

	servicesSecondarySort := in.ServicesSecondarySort
	if servicesSecondarySort == "" {
		servicesSecondarySort = "status"
	}

	dashboard := upapi.Dashboard{
		Name:                       in.Name,
		ServicesSelected:           servicesSelected,
		ServicesTags:               in.ServicesTags,
		IsPinned:                   in.IsPinned,
		ServicesShowSection:        in.ServicesShowSection,
		ServicesNumToShow:          servicesNumToShow,
		ServicesIncludeUp:          in.ServicesIncludeUp,
		ServicesIncludeDown:        in.ServicesIncludeDown,
		ServicesIncludePaused:      in.ServicesIncludePaused,
		ServicesIncludeMaintenance: in.ServicesIncludeMaintenance,
		ServicesPrimarySort:        servicesPrimarySort,
		ServicesSecondarySort:      servicesSecondarySort,
		MetricsShowSection:         in.MetricsShowSection,
		MetricsForAllChecks:        in.MetricsForAllChecks,
		AlertsShowSection:          in.AlertsShowSection,
		AlertsForAllChecks:         in.AlertsForAllChecks,
		AlertsnumToShow:            alertsNumToShow,
	}

	created, err := client.Dashboards().Create(ctx, dashboard)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create dashboard: %w", err)
	}

	return textResult(fmt.Sprintf("Created dashboard #%d: %s", created.PK, created.Name)), nil, nil
}
