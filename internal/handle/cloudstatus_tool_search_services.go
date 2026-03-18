package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerSearchCloudStatusServicesTool(srv *mcp.Server, h *cloudStatusHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "search_cloudstatus_services",
		Description: "Search for cloud status services by provider and/or name. Returns service names that can be used as service_name when creating a Cloud Status check. Use list_cloudstatus_providers first to discover valid provider names.",
	}, h.HandleSearchCloudStatusServices)
}

type searchCloudStatusServicesInput struct {
	Provider string `json:"provider,omitempty" jsonschema:"exact provider name to filter by, e.g. AWS, Azure, GCP"`
	Search   string `json:"search,omitempty" jsonschema:"search text to match against service names, case-insensitive"`
	Page     int    `json:"page,omitempty" jsonschema:"page number, defaults to 1"`
	PageSize int    `json:"page_size,omitempty" jsonschema:"results per page, defaults to 25"`
}

func (h *cloudStatusHandler) HandleSearchCloudStatusServices(_ context.Context, _ *mcp.CallToolRequest, in searchCloudStatusServicesInput) (*mcp.CallToolResult, any, error) {
	if in.Provider == "" && in.Search == "" {
		return nil, nil, fmt.Errorf("at least one of provider or search is required")
	}

	page := in.Page
	if page < 1 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize < 1 {
		pageSize = 25
	}

	result := h.index.Search(in.Provider, in.Search, page, pageSize)

	var sb strings.Builder
	sb.WriteString(formatPaginationHeader(int64(result.TotalCount), int64(page), int64(pageSize), len(result.Services)))
	for _, s := range result.Services {
		fmt.Fprintf(&sb, "- [%s] %s\n", s.Group, s.Name)
	}

	if len(result.Services) == 0 {
		sb.WriteString("No services found matching the criteria.\n")
	}

	return textResult(sb.String()), nil, nil
}
