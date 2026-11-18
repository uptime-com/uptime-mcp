package handle

import (
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func TestFormatTagColor(t *testing.T) {
	assert.Equal(t, "#2196f3", formatTagColor("#2196f3"))
	assert.Equal(t, "#2196f3", formatTagColor("2196f3"))
	assert.Equal(t, "", formatTagColor(""))
}

func TestHandleListTags(t *testing.T) {
	svc := newTagsServiceMock(t)
	svc.EXPECT().List(mock.Anything, mock.Anything).Return(&upapi.ListResult[upapi.Tag]{
		TotalCount: 2,
		Items: []upapi.Tag{
			{PK: 1, Tag: "prefixed", ColorHex: "#2196f3"},
			{PK: 2, Tag: "bare", ColorHex: "51e898"},
		},
	}, nil)

	client := newClientMock(t)
	client.EXPECT().Tags().Return(svc)

	h := &tagsHandler{}
	result, _, err := h.HandleListTags(testContext(t, client), nil, listTagsInput{})

	require.NoError(t, err)
	text := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, "[1] prefixed (color: #2196f3)")
	assert.Contains(t, text, "[2] bare (color: #51e898)")
	assert.NotContains(t, text, "##")
}

func TestLoadTag(t *testing.T) {
	svc := newTagsServiceMock(t)
	svc.EXPECT().Get(mock.Anything, upapi.PrimaryKey(1)).Return(&upapi.Tag{PK: 1, Tag: "prefixed", ColorHex: "#2196f3"}, nil)

	client := newClientMock(t)
	client.EXPECT().Tags().Return(svc)

	var sb strings.Builder
	h := &tagsHandler{}
	require.NoError(t, h.loadTag(testContext(t, client), client, 1, &sb))
	assert.Contains(t, sb.String(), "Color: #2196f3\n")
	assert.NotContains(t, sb.String(), "##")
}
