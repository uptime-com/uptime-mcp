package handle

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateContactTool(srv *mcp.Server, h *contactsHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_contact",
		Description: "Create a new contact with at least one contact method (email, SMS, phone, or integration)",
	}, h.handleCreateContact)
}

type createContactInput struct {
	Name      string   `json:"name" jsonschema:"contact group name"`
	EmailList []string `json:"email_list,omitempty" jsonschema:"email addresses for notifications"`
	SMSList   []string `json:"sms_list,omitempty" jsonschema:"phone numbers for SMS notifications in E.164 format"`
}

func (h *contactsHandler) handleCreateContact(ctx context.Context, _ *mcp.CallToolRequest, in createContactInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	if len(in.EmailList) == 0 && len(in.SMSList) == 0 {
		return nil, nil, fmt.Errorf("at least one contact method required (email_list or sms_list)")
	}

	contact := upapi.Contact{
		Name:      in.Name,
		EmailList: in.EmailList,
		SmsList:   in.SMSList,
	}

	created, err := client.Contacts().Create(ctx, contact)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create contact: %w", err)
	}

	return textResult(fmt.Sprintf("Created contact #%d: %s", created.PK, created.Name)), nil, nil
}
