package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateAPICheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "create_api_check",
		Description: "Create a new API monitoring check that executes multi-step HTTP request sequences. The script field is a JSON array of step objects with step_def and values. IMPORTANT: Do not guess the script format — only use this tool if you have precise knowledge of the API check scripting syntax from documentation or context. Use list_locations for valid probe locations and list_contacts for contact group names.",
	}, h.HandleCreateAPICheck)
}

type createAPICheckInput struct {
	Name            string   `json:"name" jsonschema:"display name for the check"`
	Script          string   `json:"script" jsonschema:"JSON array of step objects — do not guess the format, refer to API check scripting documentation"`
	Interval        int64    `json:"interval,omitempty" jsonschema:"check frequency in minutes, defaults to 5"`
	Threshold       int64    `json:"threshold,omitempty" jsonschema:"timeout in seconds, defaults to 30"`
	Locations       []string `json:"locations" jsonschema:"probe location identifiers, use list_locations tool to discover valid values"`
	ContactGroups   []string `json:"contact_groups" jsonschema:"contact group names to notify on alerts, use list_contacts tool to discover"`
	Tags            []string `json:"tags,omitempty" jsonschema:"tag names to assign"`
	Sensitivity     int64    `json:"sensitivity,omitempty" jsonschema:"number of locations that must confirm an outage before alerting, 0 uses account default"`
	NumRetries      int64    `json:"num_retries,omitempty" jsonschema:"number of retries before marking as down"`
	UseIPVersion    string   `json:"use_ip_version,omitempty" jsonschema:"IP version to use: IPV4 or IPV6"`
	Notes           string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	IsPaused        *bool    `json:"is_paused,omitempty" jsonschema:"whether the check starts in a paused state"`
}

func (c *checksHandler) HandleCreateAPICheck(ctx context.Context, _ *mcp.CallToolRequest, in createAPICheckInput) (*mcp.CallToolResult, any, error) {
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

	check := upapi.CheckAPI{
		Name:          in.Name,
		Script:        in.Script,
		Interval:      interval,
		Threshold:     threshold,
		Locations:     in.Locations,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Sensitivity:   in.Sensitivity,
		NumRetries:    in.NumRetries,
		UseIPVersion:  in.UseIPVersion,
		Notes:         in.Notes,
		IsPaused:      in.IsPaused,
	}

	created, err := client.Checks().CreateAPI(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create API check: %w", err)
	}

	return textResult(fmt.Sprintf("Created API check #%d: %s", created.PK, created.Name)), nil, nil
}
