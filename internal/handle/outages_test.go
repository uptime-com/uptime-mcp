package handle

import (
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func TestHandleListOutages(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	svc := newOutagesServiceMock(t)
	svc.EXPECT().List(mock.Anything, mock.Anything).Return(&upapi.ListResult[upapi.Outage]{
		Items: []upapi.Outage{
			{PK: 1, CheckName: "Web Check", CheckMonitoringServiceType: "HTTP", CreatedAt: createdAt},
			{PK: 2, CheckName: "API Check", CheckMonitoringServiceType: "HTTP", CreatedAt: createdAt},
		},
	}, nil)

	client := newClientMock(t)
	client.EXPECT().Outages().Return(svc)

	h := &outagesHandler{}
	result, _, err := h.HandleListOutages(testContext(t, client), nil, listOutagesInput{PageSize: 2})

	require.NoError(t, err)
	text := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, "Showing 2 results (page 1).\nUse page=2 to see next page.")
	assert.Contains(t, text, "[1] Web Check (HTTP) - ongoing")
}
