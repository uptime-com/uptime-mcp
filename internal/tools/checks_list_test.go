package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	api "github.com/uptime-com/uptime-client-go"
)

func TestHandleListChecks(t *testing.T) {
	t.Run("returns formatted list of checks", func(t *testing.T) {
		svc := newChecksServiceMock(t)
		svc.EXPECT().List(mock.Anything, mock.Anything).Return([]*api.Check{
			{PK: 1, Name: "Web Check", CheckType: "HTTP", Address: "https://example.com"},
			{PK: 2, Name: "API Check", CheckType: "HTTP", Address: "https://api.example.com"},
		}, nil, nil)

		h := &checksHandler{service: svc}
		result, _, err := h.HandleListChecks(context.Background(), nil, listChecksInput{})

		require.NoError(t, err)
		require.Len(t, result.Content, 1)
		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "Found 2 checks")
		assert.Contains(t, text, "[1] Web Check")
		assert.Contains(t, text, "[2] API Check")
	})

	t.Run("passes filter options to service", func(t *testing.T) {
		svc := newChecksServiceMock(t)
		svc.EXPECT().List(mock.Anything, &api.CheckListOptions{
			Search:                "prod",
			Tag:                   []string{"critical"},
			MonitoringServiceType: "HTTP",
			Page:                  2,
			PageSize:              10,
		}).Return([]*api.Check{}, nil, nil)

		h := &checksHandler{service: svc}
		_, _, err := h.HandleListChecks(context.Background(), nil, listChecksInput{
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
		svc.EXPECT().List(mock.Anything, mock.Anything).Return(nil, nil, assert.AnError)

		h := &checksHandler{service: svc}
		_, _, err := h.HandleListChecks(context.Background(), nil, listChecksInput{})

		assert.ErrorContains(t, err, "failed to list checks")
	})
}