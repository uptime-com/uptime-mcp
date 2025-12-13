package handle

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerGetContactTool(srv *mcp.Server, h *contactsHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_contact",
		Description: "Get detailed information about a contact",
	}, h.HandleGetContact)
}

type getContactInput struct {
	ID int64 `json:"id"`
}

func (h *contactsHandler) HandleGetContact(ctx context.Context, _ *mcp.CallToolRequest, in getContactInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	var sb strings.Builder
	if err := h.loadContact(ctx, client, in.ID, &sb); err != nil {
		return nil, nil, err
	}

	return textResult(sb.String()), nil, nil
}
