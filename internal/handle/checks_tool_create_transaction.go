package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateTransactionCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "create_transaction_check",
		Description: "Create a new Transaction monitoring check that executes multi-step browser interactions in Chromium (Puppeteer). Only provide a script if you have precise knowledge of the Transaction check scripting syntax from documentation or context, do not guess. Use list_locations for valid probe locations and list_contacts for contact group names.",
	}, h.HandleCreateTransactionCheck)
}

type createTransactionCheckInput struct {
	Name          string   `json:"name" jsonschema:"display name for the check"`
	Script        string   `json:"script" jsonschema:"check script, requires precise knowledge of the format, do not guess"`
	Interval      int64    `json:"interval,omitempty" jsonschema:"check frequency in minutes, defaults to 5"`
	Threshold     int64    `json:"threshold,omitempty" jsonschema:"timeout in seconds, defaults to 30"`
	Locations     []string `json:"locations" jsonschema:"probe location identifiers, use list_locations tool to discover valid values"`
	ContactGroups []string `json:"contact_groups" jsonschema:"contact group names to notify on alerts, use list_contacts tool to discover"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign"`
	Sensitivity   int64    `json:"sensitivity,omitempty" jsonschema:"number of locations that must confirm an outage before alerting, 0 uses account default"`
	NumRetries    int64    `json:"num_retries,omitempty" jsonschema:"number of retries before marking as down"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	IsPaused      *bool    `json:"is_paused,omitempty" jsonschema:"whether the check starts in a paused state"`
}

func (c *checksHandler) HandleCreateTransactionCheck(ctx context.Context, _ *mcp.CallToolRequest, in createTransactionCheckInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" || in.Script == "" {
		return nil, nil, fmt.Errorf("name and script are required")
	}

	interval := in.Interval
	if interval == 0 {
		interval = 5
	}
	threshold := in.Threshold
	if threshold == 0 {
		threshold = 30
	}

	var contactGroups *[]string
	if len(in.ContactGroups) > 0 {
		contactGroups = &in.ContactGroups
	}

	check := upapi.CheckTransaction{
		Name:          in.Name,
		Script:        in.Script,
		Interval:      interval,
		Threshold:     threshold,
		Locations:     in.Locations,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Sensitivity:   in.Sensitivity,
		NumRetries:    in.NumRetries,
		Notes:         in.Notes,
		IsPaused:      in.IsPaused,
	}

	created, err := client.Checks().CreateTransaction(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Transaction check: %w", err)
	}

	return textResult(fmt.Sprintf("Created Transaction check #%d: %s", created.PK, created.Name)), nil, nil
}
