package handle

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func TestHandleGetCheckStats(t *testing.T) {
	t.Run("returns formatted stats with uptime and response time", func(t *testing.T) {
		uptime := 0.995
		responseTime := 245.5

		svc := newChecksServiceMock(t)
		svc.EXPECT().Stats(mock.Anything, upapi.PrimaryKey(123), upapi.CheckStatsOptions{}).Return(&upapi.ListResult[upapi.CheckStats]{
			Items: []upapi.CheckStats{
				{Date: "2024-12-13", Outages: 0, DowntimeSecs: 0, Uptime: &uptime, ResponseTime: &responseTime},
				{Date: "2024-12-12", Outages: 1, DowntimeSecs: 1800, Uptime: &uptime, ResponseTime: &responseTime},
			},
		}, nil)

		client := newClientMock(t)
		client.EXPECT().Checks().Return(svc)

		h := &checksHandler{}
		ctx := testContext(t, client)
		result, _, err := h.HandleGetCheckStats(ctx, nil, getCheckStatsInput{ID: 123})

		require.NoError(t, err)
		require.Len(t, result.Content, 1)
		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "Statistics for check #123")
		assert.Contains(t, text, "Totals:")
		assert.Contains(t, text, "Outages: 1")
		assert.Contains(t, text, "Downtime: 1800 seconds")
		assert.Contains(t, text, "99.50% uptime")
		assert.Contains(t, text, "246ms response")
		assert.Contains(t, text, "2024-12-12")
		assert.Contains(t, text, "1 outage(s)")
	})

	t.Run("passes date range and location to service", func(t *testing.T) {
		svc := newChecksServiceMock(t)
		svc.EXPECT().Stats(mock.Anything, upapi.PrimaryKey(456), upapi.CheckStatsOptions{
			StartDate: "2024-12-01",
			EndDate:   "2024-12-13",
			Location:  "US-East",
		}).Return(&upapi.ListResult[upapi.CheckStats]{Items: []upapi.CheckStats{}}, nil)

		client := newClientMock(t)
		client.EXPECT().Checks().Return(svc)

		h := &checksHandler{}
		ctx := testContext(t, client)
		result, _, err := h.HandleGetCheckStats(ctx, nil, getCheckStatsInput{
			ID:        456,
			StartDate: "2024-12-01",
			EndDate:   "2024-12-13",
			Location:  "US-East",
		})

		require.NoError(t, err)
		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "Statistics for check #456")
		assert.Contains(t, text, "Period: 2024-12-01 to 2024-12-13")
	})

	t.Run("handles empty stats", func(t *testing.T) {
		svc := newChecksServiceMock(t)
		svc.EXPECT().Stats(mock.Anything, mock.Anything, mock.Anything).Return(&upapi.ListResult[upapi.CheckStats]{
			Items: []upapi.CheckStats{},
		}, nil)

		client := newClientMock(t)
		client.EXPECT().Checks().Return(svc)

		h := &checksHandler{}
		ctx := testContext(t, client)
		result, _, err := h.HandleGetCheckStats(ctx, nil, getCheckStatsInput{ID: 789})

		require.NoError(t, err)
		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "Statistics for check #789")
		assert.Contains(t, text, "Totals:")
		assert.Contains(t, text, "Outages: 0")
	})

	t.Run("handles nil uptime and response time", func(t *testing.T) {
		svc := newChecksServiceMock(t)
		svc.EXPECT().Stats(mock.Anything, mock.Anything, mock.Anything).Return(&upapi.ListResult[upapi.CheckStats]{
			Items: []upapi.CheckStats{
				{Date: "2024-12-13", Outages: 0, DowntimeSecs: 0, Uptime: nil, ResponseTime: nil},
			},
		}, nil)

		client := newClientMock(t)
		client.EXPECT().Checks().Return(svc)

		h := &checksHandler{}
		ctx := testContext(t, client)
		result, _, err := h.HandleGetCheckStats(ctx, nil, getCheckStatsInput{ID: 100})

		require.NoError(t, err)
		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "2024-12-13: 0 outages")
		assert.NotContains(t, text, "uptime")
		assert.NotContains(t, text, "response")
	})

	t.Run("returns error when id is missing", func(t *testing.T) {
		h := &checksHandler{}
		ctx := testContext(t, newClientMock(t))
		_, _, err := h.HandleGetCheckStats(ctx, nil, getCheckStatsInput{})

		assert.ErrorContains(t, err, "id is required")
	})

	t.Run("returns error on service failure", func(t *testing.T) {
		svc := newChecksServiceMock(t)
		svc.EXPECT().Stats(mock.Anything, mock.Anything, mock.Anything).Return((*upapi.ListResult[upapi.CheckStats])(nil), assert.AnError)

		client := newClientMock(t)
		client.EXPECT().Checks().Return(svc)

		h := &checksHandler{}
		ctx := testContext(t, client)
		_, _, err := h.HandleGetCheckStats(ctx, nil, getCheckStatsInput{ID: 123})

		assert.ErrorContains(t, err, "failed to get check stats")
	})

	t.Run("returns error when no client in context", func(t *testing.T) {
		h := &checksHandler{}
		_, _, err := h.HandleGetCheckStats(context.Background(), nil, getCheckStatsInput{ID: 123})

		assert.ErrorContains(t, err, "no API client in context")
	})
}
