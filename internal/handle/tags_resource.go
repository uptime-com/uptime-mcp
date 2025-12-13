package handle

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const tagURIPrefix = "https://uptime.com/api/v1/check-tags/"

func registerTagResource(srv *mcp.Server, h *tags) {
	srv.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: tagURIPrefix + "{id}",
		Name:        "tag",
		Description: "Uptime.com check tag details",
		MIMEType:    "text/plain",
	}, h.handleTagResource)
}

func (t *tags) handleTagResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	uri := req.Params.URI

	idStr := strings.TrimPrefix(uri, tagURIPrefix)
	if idStr == uri {
		return nil, fmt.Errorf("invalid tag URI: %s", uri)
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid tag ID: %s", idStr)
	}

	tag, _, err := t.service.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Tag #%d\n", tag.PK)
	fmt.Fprintf(&sb, "Name: %s\n", tag.Tag)
	fmt.Fprintf(&sb, "Color: #%s\n", tag.ColorHex)

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      uri,
			MIMEType: "text/plain",
			Text:     sb.String(),
		}},
	}, nil
}
