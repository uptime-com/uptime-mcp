package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdateDashboardTool(srv *mcp.Server, h *dashboardsHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_dashboard",
		Description: "Update an existing dashboard by ID. Only provided fields are changed.",
	}, h.HandleUpdateDashboard)
}

type updateDashboardInput struct {
	ID                         int64    `json:"id" jsonschema:"dashboard ID"`
	Name                       string   `json:"name,omitempty" jsonschema:"dashboard name"`
	ServicesSelected           []string `json:"services_selected,omitempty" jsonschema:"check IDs to include on the dashboard"`
	ServicesTags               []string `json:"services_tags,omitempty" jsonschema:"tag names to filter checks"`
	IsPinned                   *bool    `json:"is_pinned,omitempty" jsonschema:"pin dashboard to top of list"`
	ServicesShowSection        *bool    `json:"services_show_section,omitempty" jsonschema:"show services section"`
	ServicesNumToShow          *int64   `json:"services_num_to_show,omitempty" jsonschema:"number of services to display"`
	ServicesIncludeUp          *bool    `json:"services_include_up,omitempty" jsonschema:"include checks with UP status"`
	ServicesIncludeDown        *bool    `json:"services_include_down,omitempty" jsonschema:"include checks with DOWN status"`
	ServicesIncludePaused      *bool    `json:"services_include_paused,omitempty" jsonschema:"include paused checks"`
	ServicesIncludeMaintenance *bool    `json:"services_include_maintenance,omitempty" jsonschema:"include checks in maintenance"`
	MetricsShowSection         *bool    `json:"metrics_show_section,omitempty" jsonschema:"show metrics section"`
	MetricsForAllChecks        *bool    `json:"metrics_for_all_checks,omitempty" jsonschema:"show metrics for all checks"`
	AlertsShowSection          *bool    `json:"alerts_show_section,omitempty" jsonschema:"show alerts section"`
	AlertsForAllChecks         *bool    `json:"alerts_for_all_checks,omitempty" jsonschema:"show alerts for all checks"`
}

func (h *dashboardsHandler) HandleUpdateDashboard(ctx context.Context, _ *mcp.CallToolRequest, in updateDashboardInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	// Fetch current dashboard to merge with updates
	current, err := client.Dashboards().Get(ctx, upapi.PrimaryKey(in.ID))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get dashboard: %w", err)
	}

	dashboard := *current

	if in.Name != "" {
		dashboard.Name = in.Name
	}
	if in.ServicesSelected != nil {
		dashboard.ServicesSelected = in.ServicesSelected
	}
	if in.ServicesTags != nil {
		dashboard.ServicesTags = in.ServicesTags
	}
	if in.IsPinned != nil {
		dashboard.IsPinned = *in.IsPinned
	}
	if in.ServicesShowSection != nil {
		dashboard.ServicesShowSection = *in.ServicesShowSection
	}
	if in.ServicesNumToShow != nil {
		dashboard.ServicesNumToShow = *in.ServicesNumToShow
	}
	if in.ServicesIncludeUp != nil {
		dashboard.ServicesIncludeUp = *in.ServicesIncludeUp
	}
	if in.ServicesIncludeDown != nil {
		dashboard.ServicesIncludeDown = *in.ServicesIncludeDown
	}
	if in.ServicesIncludePaused != nil {
		dashboard.ServicesIncludePaused = *in.ServicesIncludePaused
	}
	if in.ServicesIncludeMaintenance != nil {
		dashboard.ServicesIncludeMaintenance = *in.ServicesIncludeMaintenance
	}
	if in.MetricsShowSection != nil {
		dashboard.MetricsShowSection = *in.MetricsShowSection
	}
	if in.MetricsForAllChecks != nil {
		dashboard.MetricsForAllChecks = *in.MetricsForAllChecks
	}
	if in.AlertsShowSection != nil {
		dashboard.AlertsShowSection = *in.AlertsShowSection
	}
	if in.AlertsForAllChecks != nil {
		dashboard.AlertsForAllChecks = *in.AlertsForAllChecks
	}

	updated, err := client.Dashboards().Update(ctx, upapi.PrimaryKey(in.ID), dashboard)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update dashboard: %w", err)
	}

	return textResult(fmt.Sprintf("Updated dashboard #%d: %s", updated.PK, updated.Name)), nil, nil
}
