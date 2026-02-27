package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateGroupCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_group_check",
		Description: "Create a new group check that aggregates multiple existing checks. Use list_checks to find check names and list_tags to find tag names for grouping.",
	}, h.HandleCreateGroupCheck)
}

type createGroupCheckInput struct {
	Name          string   `json:"name" jsonschema:"display name for the group check"`
	ContactGroups []string `json:"contact_groups" jsonschema:"contact group names to notify on alerts, use list_contacts tool to discover"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign, use create_tag to create new tags first"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	Services      []string `json:"services,omitempty" jsonschema:"check names to include in the group, use list_checks to discover"`
	CheckTags     []string `json:"check_tags,omitempty" jsonschema:"tag names whose checks to include in the group"`
	DownCondition string   `json:"down_condition,omitempty" jsonschema:"condition for group to be considered down, e.g. ANY or ALL"`
}

func (c *checksHandler) HandleCreateGroupCheck(ctx context.Context, _ *mcp.CallToolRequest, in createGroupCheckInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}

	var contactGroups *[]string
	if len(in.ContactGroups) > 0 {
		contactGroups = &in.ContactGroups
	}

	check := upapi.CheckGroup{
		Name:          in.Name,
		ContactGroups: contactGroups,
		Tags:          in.Tags,
		Notes:         in.Notes,
		Config: upapi.CheckGroupConfig{
			CheckServices:      in.Services,
			CheckTags:          in.CheckTags,
			CheckDownCondition: in.DownCondition,
		},
	}

	created, err := client.Checks().CreateGroup(ctx, check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create group check: %w", err)
	}

	return textResult(fmt.Sprintf("Created group check #%d: %s", created.PK, created.Name)), nil, nil
}
