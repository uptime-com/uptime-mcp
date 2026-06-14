package handle

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

const tagURIPrefix = "uptime://tags/"

func registerTagResource(srv *mcp.Server, h *tagsHandler) {
	srv.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: tagURIPrefix + "{id}",
		Name:        "tag",
		Description: "Uptime.com check tag details",
		MIMEType:    "text/plain",
	}, h.handleTagResource)
}

func (h *tagsHandler) handleTagResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	uri := req.Params.URI

	idStr := strings.TrimPrefix(uri, tagURIPrefix)
	if idStr == uri {
		return nil, fmt.Errorf("invalid tag URI: %s", uri)
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid tag ID: %s", idStr)
	}

	var sb strings.Builder
	if err := h.loadTag(ctx, client, id, &sb); err != nil {
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

func (h *tagsHandler) loadTag(ctx context.Context, client upapi.API, id int64, sb *strings.Builder) error {
	tag, err := client.Tags().Get(ctx, upapi.PrimaryKey(id))
	if err != nil {
		return fmt.Errorf("failed to get tag: %w", err)
	}

	fmt.Fprintf(sb, "Tag #%d\n", tag.PK)
	fmt.Fprintf(sb, "Name: %s\n", tag.Tag)
	fmt.Fprintf(sb, "Color: #%s\n", tag.ColorHex)

	return nil
}
