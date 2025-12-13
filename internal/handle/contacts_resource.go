package handle

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

const contactURIPrefix = "https://uptime.com/api/v1/contacts/"

func registerContactResource(srv *mcp.Server, h *contactsHandler) {
	srv.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: contactURIPrefix + "{id}",
		Name:        "contact",
		Description: "Uptime.com contact details",
		MIMEType:    "text/plain",
	}, h.handleContactResource)
}

func (h *contactsHandler) handleContactResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	uri := req.Params.URI

	idStr := strings.TrimPrefix(uri, contactURIPrefix)
	if idStr == uri || idStr == "" {
		return nil, fmt.Errorf("invalid contact URI: %s", uri)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid contact ID: %s", idStr)
	}

	contact, err := h.service.Get(ctx, upapi.PrimaryKey(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Contact #%d\n", contact.PK)
	fmt.Fprintf(&sb, "Name: %s\n", contact.Name)

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      uri,
			MIMEType: "text/plain",
			Text:     sb.String(),
		}},
	}, nil
}
