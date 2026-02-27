package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerListContactsTool(srv *mcp.Server, h *contactsHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_contacts",
		Description: "List contacts with optional search filter",
	}, h.handleListContacts)
}

type listContactsInput struct {
	Search   string `json:"search,omitempty" jsonschema:"filter contacts by name"`
	Page     int64  `json:"page,omitempty" jsonschema:"page number, defaults to 1"`
	PageSize int64  `json:"page_size,omitempty" jsonschema:"results per page, defaults to 25"`
}

func (h *contactsHandler) handleListContacts(ctx context.Context, _ *mcp.CallToolRequest, in listContactsInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	pageSize := in.PageSize
	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	page := in.Page
	if page == 0 {
		page = 1
	}

	opts := upapi.ContactListOptions{
		Search:   in.Search,
		Page:     in.Page,
		PageSize: in.PageSize,
	}

	result, err := client.Contacts().List(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list contacts: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(formatPaginationHeader(result.TotalCount, page, pageSize, len(result.Items)))
	for _, c := range result.Items {
		fmt.Fprintf(&sb, "- [%d] %s\n", c.PK, c.Name)
	}

	return textResult(sb.String()), nil, nil
}
