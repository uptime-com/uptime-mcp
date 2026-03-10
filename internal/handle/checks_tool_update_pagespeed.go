package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdatePageSpeedCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_pagespeed_check",
		Description: "Update an existing page speed monitoring check by ID. Only provided fields are changed.",
	}, h.HandleUpdatePageSpeedCheck)
}

type updatePageSpeedCheckInput struct {
	ID                   int64    `json:"id" jsonschema:"check ID"`
	Name                 string   `json:"name,omitempty" jsonschema:"display name for the check"`
	Address              string   `json:"address,omitempty" jsonschema:"URL of the page to measure"`
	Interval             int64    `json:"interval,omitempty" jsonschema:"check frequency in minutes"`
	Locations            []string `json:"locations,omitempty" jsonschema:"probe location identifiers"`
	ContactGroups        []string `json:"contact_groups,omitempty" jsonschema:"contact group names to notify on alerts"`
	Tags                 []string `json:"tags,omitempty" jsonschema:"tag names to assign"`
	Notes                string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	EmulatedDevice       string   `json:"emulated_device,omitempty" jsonschema:"device to emulate, e.g. desktop or mobile"`
	ConnectionThrottling string   `json:"connection_throttling,omitempty" jsonschema:"network throttling profile to simulate"`
	UptimeGradeThreshold string   `json:"uptime_grade_threshold,omitempty" jsonschema:"minimum Lighthouse grade threshold"`
}

func (c *checksHandler) HandleUpdatePageSpeedCheck(ctx context.Context, _ *mcp.CallToolRequest, in updatePageSpeedCheckInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	var contactGroups *[]string
	if len(in.ContactGroups) > 0 {
		contactGroups = &in.ContactGroups
	}

	check := upapi.CheckPageSpeed{
		Name:          in.Name,
		Address:       in.Address,
		Interval:      in.Interval,
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

	updated, err := client.Checks().UpdatePageSpeed(ctx, upapi.PrimaryKey(in.ID), check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update page speed check: %w", err)
	}

	return textResult(fmt.Sprintf("Updated page speed check #%d: %s", updated.PK, updated.Name)), nil, nil
}
