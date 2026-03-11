package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerUpdateContactTool(srv *mcp.Server, h *contactsHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "update_contact",
		Description: "Update an existing contact group by ID. Only provided fields are changed.",
	}, h.handleUpdateContact)
}

type updateContactInput struct {
	ID        int64    `json:"id" jsonschema:"contact group ID"`
	Name      string   `json:"name,omitempty" jsonschema:"contact group name"`
	EmailList []string `json:"email_list,omitempty" jsonschema:"email addresses for notifications"`
	SMSList   []string `json:"sms_list,omitempty" jsonschema:"phone numbers for SMS notifications in E.164 format"`
}

func (h *contactsHandler) handleUpdateContact(ctx context.Context, _ *mcp.CallToolRequest, in updateContactInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.ID == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	contact := upapi.Contact{
		Name:      in.Name,
		EmailList: in.EmailList,
		SmsList:   in.SMSList,
	}

	updated, err := client.Contacts().Update(ctx, upapi.PrimaryKey(in.ID), contact)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update contact: %w", err)
	}

	return textResult(fmt.Sprintf("Updated contact #%d: %s", updated.PK, updated.Name)), nil, nil
}
