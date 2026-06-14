package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdateGroupCheckTool(srv *mcp.Server, h *checksHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_group_check",
		Description: "Update an existing group check by ID. Only provided fields are changed.",
	}, h.HandleUpdateGroupCheck)
}

type updateGroupCheckInput struct {
	ID            int64    `json:"id" jsonschema:"check ID"`
	Name          string   `json:"name,omitempty" jsonschema:"display name for the group check"`
	ContactGroups []string `json:"contact_groups,omitempty" jsonschema:"contact group names to notify on alerts"`
	Tags          []string `json:"tags,omitempty" jsonschema:"tag names to assign"`
	Notes         string   `json:"notes,omitempty" jsonschema:"free-text notes for the check"`
	Services      []string `json:"services,omitempty" jsonschema:"check names to include in the group"`
	CheckTags     []string `json:"check_tags,omitempty" jsonschema:"tag names whose checks to include in the group"`
	DownCondition string   `json:"down_condition,omitempty" jsonschema:"condition for group to be considered down, e.g. ANY or ALL"`
}

func (c *checksHandler) HandleUpdateGroupCheck(ctx context.Context, _ *mcp.CallToolRequest, in updateGroupCheckInput) (*mcp.CallToolResult, any, error) {
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

	updated, err := client.Checks().UpdateGroup(ctx, upapi.PrimaryKey(in.ID), check)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update group check: %w", err)
	}

	return textResult(fmt.Sprintf("Updated group check #%d: %s", updated.PK, updated.Name)), nil, nil
}
