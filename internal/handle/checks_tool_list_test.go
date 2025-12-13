package handle

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"

	"github.com/uptime-com/uptime-mcp/internal/app"
)

// testContext creates a context with a mock client for testing.
func testContext(t *testing.T, client upapi.API) context.Context {
	t.Helper()
	session := &app.Session{Client: client}
	return app.ContextWithSession(context.Background(), session)
}

func TestHandleListChecks(t *testing.T) {
	t.Run("returns formatted list of checks", func(t *testing.T) {
		svc := newChecksServiceMock(t)
		svc.EXPECT().List(mock.Anything, mock.Anything).Return(&upapi.ListResult[upapi.Check]{
			TotalCount: 2,
			Items: []upapi.Check{
				{PK: 1, Name: "Web Check", MonitoringServiceType: "HTTP", Address: "https://example.com"},
				{PK: 2, Name: "API Check", MonitoringServiceType: "HTTP", Address: "https://api.example.com"},
			},
		}, nil)

		client := newClientMock(t)
		client.EXPECT().Checks().Return(svc)

		h := &checksHandler{}
		ctx := testContext(t, client)
		result, _, err := h.HandleListChecks(ctx, nil, listChecksInput{})

		require.NoError(t, err)
		require.Len(t, result.Content, 1)
		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "Found 2 results")
		assert.Contains(t, text, "[1] Web Check")
		assert.Contains(t, text, "[2] API Check")
	})

	t.Run("passes filter options to service", func(t *testing.T) {
		svc := newChecksServiceMock(t)
		svc.EXPECT().List(mock.Anything, upapi.CheckListOptions{
			Search:                "prod",
			Tag:                   []string{"critical"},
			MonitoringServiceType: "HTTP",
			Page:                  2,
			PageSize:              10,
		}).Return(&upapi.ListResult[upapi.Check]{TotalCount: 0, Items: []upapi.Check{}}, nil)

		client := newClientMock(t)
		client.EXPECT().Checks().Return(svc)

		h := &checksHandler{}
		ctx := testContext(t, client)
		_, _, err := h.HandleListChecks(ctx, nil, listChecksInput{
			Search:   "prod",
			Tag:      "critical",
			Type:     "HTTP",
			Page:     2,
			PageSize: 10,
		})

		assert.NoError(t, err)
	})

	t.Run("returns error on service failure", func(t *testing.T) {
		svc := newChecksServiceMock(t)
		svc.EXPECT().List(mock.Anything, mock.Anything).Return((*upapi.ListResult[upapi.Check])(nil), assert.AnError)

		client := newClientMock(t)
		client.EXPECT().Checks().Return(svc)

		h := &checksHandler{}
		ctx := testContext(t, client)
		_, _, err := h.HandleListChecks(ctx, nil, listChecksInput{})

		assert.ErrorContains(t, err, "failed to list checks")
	})

	t.Run("returns error when no client in context", func(t *testing.T) {
		h := &checksHandler{}
		_, _, err := h.HandleListChecks(context.Background(), nil, listChecksInput{})

		assert.ErrorContains(t, err, "no API client in context")
	})
}
