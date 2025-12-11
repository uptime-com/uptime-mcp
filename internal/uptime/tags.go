package uptime

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	api "github.com/uptime-com/uptime-client-go"
)

// ListTagsInput defines parameters for listing tags.
type ListTagsInput struct {
	Search   string `json:"search,omitempty" jsonschema:"description=Search term to filter tags"`
	Page     int    `json:"page,omitempty" jsonschema:"description=Page number (default 1)"`
	PageSize int    `json:"page_size,omitempty" jsonschema:"description=Results per page (default 25, max 100)"`
}

var listTagsTool = &mcp.Tool{
	Name:        "list_tags",
	Description: "List all check tags with optional search filtering",
}

func (p *Provider) handleListTags(ctx context.Context, _ *mcp.CallToolRequest, in ListTagsInput) (*mcp.CallToolResult, any, error) {
	client, err := getClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	opts := &api.TagListOptions{
		Search:   in.Search,
		Page:     in.Page,
		PageSize: in.PageSize,
	}

	tags, _, err := client.Tags.List(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list tags: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d tags:\n\n", len(tags))
	for _, t := range tags {
		fmt.Fprintf(&sb, "- [%d] %s (color: #%s)\n", t.PK, t.Tag, t.ColorHex)
	}

	return textResult(sb.String()), nil, nil
}

// GetTagInput defines parameters for getting a single tag.
type GetTagInput struct {
	ID int `json:"id" jsonschema:"description=Tag ID (pk)"`
}

var getTagTool = &mcp.Tool{
	Name:        "get_tag",
	Description: "Get details of a specific tag by ID",
}

func (p *Provider) handleGetTag(ctx context.Context, _ *mcp.CallToolRequest, in GetTagInput) (*mcp.CallToolResult, any, error) {
	client, err := getClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	tag, _, err := client.Tags.Get(ctx, in.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get tag: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Tag #%d\n", tag.PK)
	fmt.Fprintf(&sb, "Name: %s\n", tag.Tag)
	fmt.Fprintf(&sb, "Color: #%s\n", tag.ColorHex)

	return textResult(sb.String()), nil, nil
}

// CreateTagInput defines parameters for creating a new tag.
type CreateTagInput struct {
	Name  string `json:"name" jsonschema:"description=Tag name"`
	Color string `json:"color,omitempty" jsonschema:"description=Hex color code without # (e.g. FF5733)"`
}

var createTagTool = &mcp.Tool{
	Name:        "create_tag",
	Description: "Create a new check tag",
}

func (p *Provider) handleCreateTag(ctx context.Context, _ *mcp.CallToolRequest, in CreateTagInput) (*mcp.CallToolResult, any, error) {
	client, err := getClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}

	tag := &api.Tag{
		Tag:      in.Name,
		ColorHex: in.Color,
	}

	created, _, err := client.Tags.Create(ctx, tag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return textResult(fmt.Sprintf("Created tag #%d: %s", created.PK, created.Tag)), nil, nil
}

// UpdateTagInput defines parameters for updating an existing tag.
type UpdateTagInput struct {
	ID    int    `json:"id" jsonschema:"description=Tag ID (pk) to update"`
	Name  string `json:"name,omitempty" jsonschema:"description=New tag name"`
	Color string `json:"color,omitempty" jsonschema:"description=New hex color code without # (e.g. FF5733)"`
}

var updateTagTool = &mcp.Tool{
	Name:        "update_tag",
	Description: "Update an existing check tag",
}

func (p *Provider) handleUpdateTag(ctx context.Context, _ *mcp.CallToolRequest, in UpdateTagInput) (*mcp.CallToolResult, any, error) {
	client, err := getClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}
	if in.Name == "" && in.Color == "" {
		return nil, nil, fmt.Errorf("at least one of name or color is required")
	}

	tag := &api.Tag{
		PK:       in.ID,
		Tag:      in.Name,
		ColorHex: in.Color,
	}

	updated, _, err := client.Tags.Update(ctx, tag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update tag: %w", err)
	}

	return textResult(fmt.Sprintf("Updated tag #%d: %s", updated.PK, updated.Tag)), nil, nil
}

// DeleteTagInput defines parameters for deleting a tag.
type DeleteTagInput struct {
	ID int `json:"id" jsonschema:"description=Tag ID (pk) to delete"`
}

var deleteTagTool = &mcp.Tool{
	Name:        "delete_tag",
	Description: "Delete a check tag by ID",
}

func (p *Provider) handleDeleteTag(ctx context.Context, _ *mcp.CallToolRequest, in DeleteTagInput) (*mcp.CallToolResult, any, error) {
	client, err := getClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	_, err = client.Tags.Delete(ctx, in.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to delete tag: %w", err)
	}

	return textResult(fmt.Sprintf("Successfully deleted tag #%d", in.ID)), nil, nil
}
