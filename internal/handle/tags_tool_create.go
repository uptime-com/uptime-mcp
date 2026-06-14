package handle

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func registerCreateTagTool(srv *mcp.Server, h *tagsHandler) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "create_tag",
		Description: "Create a new check tag",
	}, h.HandleCreateTag)
}

type createTagInput struct {
	Name  string `json:"name" jsonschema:"tag name"`
	Color string `json:"color,omitempty" jsonschema:"hex color code, e.g. #FF5733, random color assigned if omitted"`
}

// tagColorPalette is an extended palette based on Uptime.com brand colors.
// Three tones (light, base, dark) for each of 12 hues.
var tagColorPalette = []string{
	// Grays
	"#bcc1c5", "#999999", "#4d4d4d",
	// Reds
	"#f19d93", "#eb5a46", "#b8392e",
	// Oranges
	"#ffd6a5", "#ffab4a", "#d48c2e",
	// Yellows
	"#f7e96b", "#f2d600", "#c4ab00",
	// Greens
	"#a3d98a", "#61bd4f", "#49953c",
	// Mints
	"#8af0bc", "#51e898", "#34b96f",
	// Cyans
	"#66d8eb", "#00c2e0", "#009bb3",
	// Blues
	"#4da3d9", "#0079bf", "#005f99",
	// Purples
	"#d5a3ec", "#c377e0", "#9b51c6",
	// Pinks
	"#ffb3e0", "#ff80ce", "#d45ca8",
	// Teals
	"#6fd4c0", "#2baf9a", "#1f8c7a",
	// Indigos
	"#8c9edb", "#5c6fba", "#3f4f99",
}

func randomTagColor() string {
	return tagColorPalette[rand.IntN(len(tagColorPalette))]
}

func (t *tagsHandler) HandleCreateTag(ctx context.Context, _ *mcp.CallToolRequest, in createTagInput) (*mcp.CallToolResult, any, error) {
	client, err := clientFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}

	color := in.Color
	if color == "" {
		color = randomTagColor()
	}

	tag := upapi.Tag{
		Tag:      in.Name,
		ColorHex: color,
	}

	created, err := client.Tags().Create(ctx, tag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return textResult(fmt.Sprintf("Created tag #%d: %s", created.PK, created.Tag)), nil, nil
}
