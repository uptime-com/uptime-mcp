package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreatePageSpeedCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_pagespeed_check",
		Description: "Create a new page speed monitoring check using Google Lighthouse. Use list_locations for valid probe locations and list_contacts for contact group names.",
	}, h.HandleCreatePageSpeedCheck)
}

type createPageSpeedCheckInput struct {
	Name                 string   `json:"name" jsonschema:"display name for the check"`
	Address              string   `json:"address" jsonschema:"URL of the page to measure, e.g. https://example.com"`
	Interval             int64    `json:"interval" jsonschema:"check frequency in minutes, defaults to 5"`
	Locations            []string `json:"locations" jsonschema:"probe location identifiers, use list_locations tool to discover valid values"`
	ContactGroups        []string `json:"contact_groups" jsonschema:"contact group names to notify on alerts, use list_contacts tool to discover"`
	Tags                 []string `json:"tags,omitempty" jsonschema:"tag names to assign, use create_tag to create new tags first"`
	Notes                string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	EmulatedDevice       string   `json:"emulated_device,omitempty" jsonschema:"device to emulate, e.g. desktop or mobile"`
	ConnectionThrottling string   `json:"connection_throttling,omitempty" jsonschema:"network throttling profile to simulate"`
	UptimeGradeThreshold string   `json:"uptime_grade_threshold,omitempty" jsonschema:"minimum Lighthouse grade threshold, check fails if below this grade"`
}

func (c *checksHandler) HandleCreatePageSpeedCheck(ctx context.Context, _ *mcp.CallToolRequest, in createPageSpeedCheckInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" || in.Address == "" {
		return nil, nil, fmt.Errorf("name and address are required")
	}

	interval := in.Interval
	if interval == 0 {
		interval = 5
	}

	var contactGroups *[]string
	if len(in.ContactGroups) > 0 {
		contactGroups = &in.ContactGroups
	}

	check := upapi.CheckPageSpeed{
		Name:          in.Name,
		Address:       in.Address,
		Interval:      interval,
		Locations:     in.Locations,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Notes:         in.Notes,
		Config: upapi.CheckPageSpeedConfig{
			EmulatedDevice:       in.EmulatedDevice,
			ConnectionThrottling: in.ConnectionThrottling,
			UptimeGradeThreshold: in.UptimeGradeThreshold,
		},
	}

	created, err := client.Checks().CreatePageSpeed(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create page speed check: %w", err)
	}

	return textResult(fmt.Sprintf("Created page speed check #%d: %s", created.PK, created.Name)), nil, nil
}
