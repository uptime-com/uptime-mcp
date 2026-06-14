package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateCloudStatusCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_cloudstatus_check",
		Description: "Create a new Cloud Status monitoring check that tracks the status of a third-party cloud provider service (e.g. AWS EC2, Azure DevOps, Cloudflare DNS). Use list_locations for valid probe locations and list_contacts for contact group names.",
	}, h.HandleCreateCloudStatusCheck)
}

type createCloudStatusCheckInput struct {
	Name          string   `json:"name" jsonschema:"display name for the check"`
	ServiceName   string   `json:"service_name" jsonschema:"cloud provider service to monitor, e.g. AWS - Amazon Elastic Compute Cloud, Azure - Azure DevOps, Cloudflare - DNS"`
	Locations     []string `json:"locations" jsonschema:"probe location identifiers, use list_locations tool to discover valid values"`
	ContactGroups []string `json:"contact_groups" jsonschema:"contact group names to notify on alerts, use list_contacts tool to discover"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign, use create_tag to create new tags first"`
	IsPaused      *bool    `json:"is_paused,omitempty" jsonschema:"whether the check starts in a paused state"`
}

func (c *checksHandler) HandleCreateCloudStatusCheck(ctx context.Context, _ *mcp.CallToolRequest, in createCloudStatusCheckInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" || in.ServiceName == "" {
		return nil, nil, fmt.Errorf("name and service_name are required")
	}

	var contactGroups *[]string
	if len(in.ContactGroups) > 0 {
		contactGroups = &in.ContactGroups
	}

	check := upapi.CheckCloudStatus{
		Name:          in.Name,
		Locations:     in.Locations,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		IsPaused:      in.IsPaused,
		CloudStatusConfig: upapi.CheckCloudStatusConfig{
			ServiceName: in.ServiceName,
		},
	}

	created, err := client.Checks().CreateCloudStatus(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Cloud Status check: %w", err)
	}

	return textResult(fmt.Sprintf("Created Cloud Status check #%d: %s (monitoring %s)", created.PK, created.Name, in.ServiceName)), nil, nil
}
