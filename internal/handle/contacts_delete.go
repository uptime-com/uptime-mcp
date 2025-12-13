package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerDeleteContactTool(srv *mcp.Server, h *contactsHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "delete_contact",
		Description: "Delete a contact by ID",
	}, h.handleDeleteContact)
}

type deleteContactInput struct {
	ID int64 `json:"id"`
}

func (h *contactsHandler) handleDeleteContact(ctx context.Context, _ *mcp.CallToolRequest, in deleteContactInput) (*mcp.CallToolResult, any, error) {
	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	err := h.service.Delete(ctx, upapi.PrimaryKey(in.ID))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to delete contact: %w", err)
	}

	return textResult(fmt.Sprintf("Successfully deleted contact #%d", in.ID)), nil, nil
}
