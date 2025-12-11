package uptime

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api "github.com/uptime-com/uptime-client-go"
)

// ListChecksInput defines parameters for listing checks.
type ListChecksInput struct {
	Search   string `json:"search,omitempty" jsonschema:"description=Search term to filter checks by name or address"`
	Tag      string `json:"tag,omitempty" jsonschema:"description=Filter by tag name"`
	Type     string `json:"type,omitempty" jsonschema:"description=Filter by check type (e.g. HTTP, DNS, SSL)"`
	IsPaused bool   `json:"is_paused,omitempty" jsonschema:"description=Filter by paused status"`
	Page     int    `json:"page,omitempty" jsonschema:"description=Page number (default 1)"`
	PageSize int    `json:"page_size,omitempty" jsonschema:"description=Results per page (default 25, max 100)"`
}

var listChecksTool = &mcp.Tool{
	Name:        "list_checks",
	Description: "List monitoring checks with optional filtering by search term, tag, or check type",
}

func (p *Provider) handleListChecks(ctx context.Context, _ *mcp.ServerSession, req *mcp.CallToolParamsFor[ListChecksInput]) (*mcp.CallToolResultFor[any], error) {
	client, err := getClient(ctx)
	if err != nil {
		return errorResult(err), nil
	}

	in := req.Arguments
	opts := &api.CheckListOptions{
		Search:                in.Search,
		MonitoringServiceType: in.Type,
		IsPaused:              in.IsPaused,
		Page:                  in.Page,
		PageSize:              in.PageSize,
	}
	if in.Tag != "" {
		opts.Tag = []string{in.Tag}
	}

	checks, _, err := client.Checks.List(ctx, opts)
	if err != nil {
		return errorResult(fmt.Errorf("failed to list checks: %w", err)), nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d checks:\n\n", len(checks))
	for _, c := range checks {
		fmt.Fprintf(&sb, "- [%d] %s (%s) - %s\n", c.PK, c.Name, c.CheckType, c.Address)
	}

	return textResult(sb.String()), nil
}

// GetCheckInput defines parameters for getting a single check.
type GetCheckInput struct {
	ID int `json:"id" jsonschema:"description=Check ID (pk)"`
}

var getCheckTool = &mcp.Tool{
	Name:        "get_check",
	Description: "Get details of a specific monitoring check by ID",
}

func (p *Provider) handleGetCheck(ctx context.Context, _ *mcp.ServerSession, req *mcp.CallToolParamsFor[GetCheckInput]) (*mcp.CallToolResultFor[any], error) {
	client, err := getClient(ctx)
	if err != nil {
		return errorResult(err), nil
	}

	if req.Arguments.ID == 0 {
		return errorResult(fmt.Errorf("id is required")), nil
	}

	check, _, err := client.Checks.Get(ctx, req.Arguments.ID)
	if err != nil {
		return errorResult(fmt.Errorf("failed to get check: %w", err)), nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Check #%d: %s\n", check.PK, check.Name)
	fmt.Fprintf(&sb, "Type: %s\n", check.CheckType)
	fmt.Fprintf(&sb, "Address: %s\n", check.Address)
	if check.Port > 0 {
		fmt.Fprintf(&sb, "Port: %d\n", check.Port)
	}
	fmt.Fprintf(&sb, "Interval: %d seconds\n", check.Interval)
	fmt.Fprintf(&sb, "Sensitivity: %d\n", check.Sensitivity)
	if len(check.Locations) > 0 {
		fmt.Fprintf(&sb, "Locations: %s\n", strings.Join(check.Locations, ", "))
	}
	if len(check.Tags) > 0 {
		fmt.Fprintf(&sb, "Tags: %s\n", strings.Join(check.Tags, ", "))
	}
	if check.Notes != "" {
		fmt.Fprintf(&sb, "Notes: %s\n", check.Notes)
	}

	return textResult(sb.String()), nil
}

// DeleteCheckInput defines parameters for deleting a check.
type DeleteCheckInput struct {
	ID int `json:"id" jsonschema:"description=Check ID (pk) to delete"`
}

var deleteCheckTool = &mcp.Tool{
	Name:        "delete_check",
	Description: "Delete a monitoring check by ID",
}

func (p *Provider) handleDeleteCheck(ctx context.Context, _ *mcp.ServerSession, req *mcp.CallToolParamsFor[DeleteCheckInput]) (*mcp.CallToolResultFor[any], error) {
	client, err := getClient(ctx)
	if err != nil {
		return errorResult(err), nil
	}

	if req.Arguments.ID == 0 {
		return errorResult(fmt.Errorf("id is required")), nil
	}

	_, err = client.Checks.Delete(ctx, req.Arguments.ID)
	if err != nil {
		return errorResult(fmt.Errorf("failed to delete check: %w", err)), nil
	}

	return textResult(fmt.Sprintf("Successfully deleted check #%d", req.Arguments.ID)), nil
}

// GetCheckStatsInput defines parameters for getting check statistics.
type GetCheckStatsInput struct {
	ID        int    `json:"id" jsonschema:"description=Check ID (pk)"`
	StartDate string `json:"start_date,omitempty" jsonschema:"description=Start date (YYYY-MM-DD format)"`
	EndDate   string `json:"end_date,omitempty" jsonschema:"description=End date (YYYY-MM-DD format)"`
}

var getCheckStatsTool = &mcp.Tool{
	Name:        "get_check_stats",
	Description: "Get statistics for a monitoring check including uptime percentage and outages",
}

func (p *Provider) handleGetCheckStats(ctx context.Context, _ *mcp.ServerSession, req *mcp.CallToolParamsFor[GetCheckStatsInput]) (*mcp.CallToolResultFor[any], error) {
	client, err := getClient(ctx)
	if err != nil {
		return errorResult(err), nil
	}

	if req.Arguments.ID == 0 {
		return errorResult(fmt.Errorf("id is required")), nil
	}

	in := req.Arguments
	opts := &api.CheckStatsOptions{
		StartDate: in.StartDate,
		EndDate:   in.EndDate,
	}

	stats, _, err := client.Checks.Stats(ctx, in.ID, opts)
	if err != nil {
		return errorResult(fmt.Errorf("failed to get check stats: %w", err)), nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Statistics for check #%d\n", in.ID)
	fmt.Fprintf(&sb, "Period: %s to %s\n\n", stats.StartDate, stats.EndDate)

	fmt.Fprintf(&sb, "Totals:\n")
	fmt.Fprintf(&sb, "  Outages: %d\n", stats.Totals.Outages)
	fmt.Fprintf(&sb, "  Downtime: %d seconds\n\n", stats.Totals.DowntimeSecs)

	if len(stats.Statistics) > 0 {
		fmt.Fprintf(&sb, "Daily breakdown:\n")
		for _, s := range stats.Statistics {
			fmt.Fprintf(&sb, "  %s: %d outages, %d sec downtime\n", s.Date, s.Outages, s.DowntimeSecs)
		}
	}

	return textResult(sb.String()), nil
}
