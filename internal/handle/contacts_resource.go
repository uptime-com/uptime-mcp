package handle

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

const contactURIPrefix = "uptime://contacts/"

func registerContactResource(srv *mcp.Server, h *contactsHandler) {
	srv.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: contactURIPrefix + "{id}",
		Name:        "contact",
		Description: "Uptime.com contact details",
		MIMEType:    "text/plain",
	}, h.handleContactResource)
}

func (h *contactsHandler) handleContactResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	uri := req.Params.URI

	idStr := strings.TrimPrefix(uri, contactURIPrefix)
	if idStr == uri || idStr == "" {
		return nil, fmt.Errorf("invalid contact URI: %s", uri)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid contact ID: %s", idStr)
	}

	var sb strings.Builder
	if err := h.loadContact(ctx, client, id, &sb); err != nil {
		return nil, err
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      uri,
			MIMEType: "text/plain",
			Text:     sb.String(),
		}},
	}, nil
}

func (h *contactsHandler) loadContact(ctx context.Context, client upapi.API, id int64, sb *strings.Builder) error {
	contact, err := client.Contacts().Get(ctx, upapi.PrimaryKey(id))
	if err != nil {
		return fmt.Errorf("failed to get contact: %w", err)
	}

	fmt.Fprintf(sb, "Contact #%d\n", contact.PK)
	fmt.Fprintf(sb, "Name: %s\n", contact.Name)

	return nil
}
