package handle

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

const statusPageURIPrefix = "uptime://statuspages/"

func registerStatusPageResource(srv *mcp.Server, h *statusPagesHandler) {
	srv.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: statusPageURIPrefix + "{id}",
		Name:        "status_page",
		Description: "Uptime.com status page details",
		MIMEType:    "text/plain",
	}, h.handleStatusPageResource)
}

func (h *statusPagesHandler) handleStatusPageResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	uri := req.Params.URI

	idStr := strings.TrimPrefix(uri, statusPageURIPrefix)
	if idStr == uri {
		return nil, fmt.Errorf("invalid status page URI: %s", uri)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid status page ID: %s", idStr)
	}

	var sb strings.Builder
	if err := h.loadStatusPage(ctx, client, id, &sb); err != nil {
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

func (h *statusPagesHandler) loadStatusPage(ctx context.Context, client upapi.API, id int64, sb *strings.Builder) error {
	sp, err := client.StatusPages().Get(ctx, upapi.PrimaryKey(id))
	if err != nil {
		return fmt.Errorf("failed to get status page: %w", err)
	}

	fmt.Fprintf(sb, "Status Page #%d\n", sp.PK)
	fmt.Fprintf(sb, "Name: %s\n", sp.Name)
	fmt.Fprintf(sb, "PageType: %s\n", sp.PageType)
	fmt.Fprintf(sb, "Slug: %s\n", sp.Slug)
	fmt.Fprintf(sb, "VisibilityLevel: %s\n", sp.VisibilityLevel)
	fmt.Fprintf(sb, "Description: %s\n", sp.Description)
	fmt.Fprintf(sb, "CNAME: %s\n", sp.CNAME)
	fmt.Fprintf(sb, "AllowSubscriptions: %t\n", sp.AllowSubscriptions)
	fmt.Fprintf(sb, "Timezone: %s\n", sp.Timezone)

	return nil
}
