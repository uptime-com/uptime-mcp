package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdateTransactionCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_transaction_check",
		Description: "Update an existing Transaction monitoring check by ID. Only provided fields are changed. Only modify the script if you have precise knowledge of the Transaction check scripting syntax, do not guess.",
	}, h.HandleUpdateTransactionCheck)
}

type updateTransactionCheckInput struct {
	ID            int64    `json:"id" jsonschema:"check ID"`
	Name          string   `json:"name,omitempty" jsonschema:"display name for the check"`
	Script        string   `json:"script,omitempty" jsonschema:"check script, requires precise knowledge of the format, do not guess"`
	Interval      int64    `json:"interval,omitempty" jsonschema:"check frequency in minutes"`
	Threshold     int64    `json:"threshold,omitempty" jsonschema:"timeout in seconds"`
	Locations     []string `json:"locations,omitempty" jsonschema:"probe location identifiers"`
	ContactGroups []string `json:"contact_groups,omitempty" jsonschema:"contact group names to notify on alerts"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign"`
	Sensitivity   int64    `json:"sensitivity,omitempty" jsonschema:"number of locations that must confirm an outage before alerting"`
	NumRetries    int64    `json:"num_retries,omitempty" jsonschema:"number of retries before marking as down"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	IsPaused      *bool    `json:"is_paused,omitempty" jsonschema:"whether the check is paused"`
}

func (c *checksHandler) HandleUpdateTransactionCheck(ctx context.Context, _ *mcp.CallToolRequest, in updateTransactionCheckInput) (*mcp.CallToolResult, any, error) {
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

	check := upapi.CheckTransaction{
		Name:          in.Name,
		Script:        in.Script,
		Interval:      in.Interval,
		Threshold:     in.Threshold,
		Locations:     in.Locations,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Sensitivity:   in.Sensitivity,
		NumRetries:    in.NumRetries,
		Notes:         in.Notes,
		IsPaused:      in.IsPaused,
	}

	updated, err := client.Checks().UpdateTransaction(ctx, upapi.PrimaryKey(in.ID), check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update Transaction check: %w", err)
	}

	return textResult(fmt.Sprintf("Updated Transaction check #%d: %s", updated.PK, updated.Name)), nil, nil
}
