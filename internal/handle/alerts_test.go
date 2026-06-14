package handle

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func TestHandleListAlerts(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	resolvedAt := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)

	t.Run("returns formatted list of alerts", func(t *testing.T) {
		svc := newAlertsServiceMock(t)
		svc.EXPECT().List(mock.Anything, mock.Anything).Return(&upapi.ListResult[upapi.AlertItem]{
			TotalCount: 2,
			Items: []upapi.AlertItem{
				{
					PK:                         1,
					CheckName:                  "Web Check",
					CheckMonitoringServiceType: "HTTP",
					StateIsUp:                  false,
					Location:                   "New York",
					MonitoringServerName:       "nyc-server-1",
					CreatedAt:                  &createdAt,
				},
				{
					PK:                         2,
					CheckName:                  "API Check",
					CheckMonitoringServiceType: "HTTP",
					StateIsUp:                  true,
					Location:                   "London",
					MonitoringServerName:       "lon-server-1",
					CreatedAt:                  &createdAt,
					ResolvedAt:                 &resolvedAt,
				},
			},
		}, nil)

		client := newClientMock(t)
		client.EXPECT().Alerts().Return(svc)

		h := &alertsHandler{}
		ctx := testContext(t, client)
		result, _, err := h.HandleListAlerts(ctx, nil, listAlertsInput{})

		require.NoError(t, err)
		require.Len(t, result.Content, 1)
		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "Found 2 results")
		assert.Contains(t, text, "[1] Web Check (HTTP) - down")
		assert.Contains(t, text, "[2] API Check (HTTP) - resolved")
		assert.Contains(t, text, "Location: New York")
		assert.Contains(t, text, "Location: London")
	})

	t.Run("shows ignored status", func(t *testing.T) {
		svc := newAlertsServiceMock(t)
		svc.EXPECT().List(mock.Anything, mock.Anything).Return(&upapi.ListResult[upapi.AlertItem]{
			TotalCount: 1,
			Items: []upapi.AlertItem{
				{
					PK:                         1,
					CheckName:                  "Ignored Check",
					CheckMonitoringServiceType: "HTTP",
					StateIsUp:                  false,
					Ignored:                    true,
					Location:                   "Tokyo",
					MonitoringServerName:       "tky-server-1",
					CreatedAt:                  &createdAt,
				},
			},
		}, nil)

		client := newClientMock(t)
		client.EXPECT().Alerts().Return(svc)

		h := &alertsHandler{}
		ctx := testContext(t, client)
		result, _, err := h.HandleListAlerts(ctx, nil, listAlertsInput{})

		require.NoError(t, err)
		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "down (ignored)")
	})

	t.Run("passes filter options to service", func(t *testing.T) {
		resolved := true
		svc := newAlertsServiceMock(t)
		svc.EXPECT().List(mock.Anything, upapi.AlertListOptions{
			CheckPK:                    123,
			CheckMonitoringServiceType: "HTTP",
			CheckTag:                   "production",
			StateIsUp:                  &resolved,
			StartDate:                  "2024-01-01",
			EndDate:                    "2024-01-31",
			Search:                     "error",
			Page:                       2,
			PageSize:                   10,
		}).Return(&upapi.ListResult[upapi.AlertItem]{TotalCount: 0, Items: []upapi.AlertItem{}}, nil)

		client := newClientMock(t)
		client.EXPECT().Alerts().Return(svc)

		h := &alertsHandler{}
		ctx := testContext(t, client)
		_, _, err := h.HandleListAlerts(ctx, nil, listAlertsInput{
			CheckID:   123,
			Type:      "HTTP",
			Tag:       "production",
			Resolved:  &resolved,
			StartDate: "2024-01-01",
			EndDate:   "2024-01-31",
			Search:    "error",
			Page:      2,
			PageSize:  10,
		})

		assert.NoError(t, err)
	})

	t.Run("returns error on service failure", func(t *testing.T) {
		svc := newAlertsServiceMock(t)
		svc.EXPECT().List(mock.Anything, mock.Anything).Return((*upapi.ListResult[upapi.AlertItem])(nil), assert.AnError)

		client := newClientMock(t)
		client.EXPECT().Alerts().Return(svc)

		h := &alertsHandler{}
		ctx := testContext(t, client)
		_, _, err := h.HandleListAlerts(ctx, nil, listAlertsInput{})

		assert.ErrorContains(t, err, "failed to list alerts")
	})

	t.Run("returns error when no client in context", func(t *testing.T) {
		h := &alertsHandler{}
		_, _, err := h.HandleListAlerts(context.Background(), nil, listAlertsInput{})

		assert.ErrorContains(t, err, "no API client in context")
	})
}

func TestHandleGetAlert(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	resolvedAt := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)

	t.Run("returns formatted alert details", func(t *testing.T) {
		svc := newAlertsServiceMock(t)
		svc.EXPECT().Get(mock.Anything, upapi.PrimaryKey(123)).Return(&upapi.AlertItem{
			PK:                         123,
			CheckName:                  "Web Check",
			CheckPK:                    456,
			CheckMonitoringServiceType: "HTTP",
			CheckAddress:               "https://example.com",
			StateIsUp:                  true,
			Location:                   "New York",
			MonitoringServerName:       "nyc-server-1",
			CreatedAt:                  &createdAt,
			ResolvedAt:                 &resolvedAt,
			Output:                     "Connection successful",
		}, nil)

		client := newClientMock(t)
		client.EXPECT().Alerts().Return(svc)

		h := &alertsHandler{}
		ctx := testContext(t, client)
		result, _, err := h.HandleGetAlert(ctx, nil, getAlertInput{ID: 123})

		require.NoError(t, err)
		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "Alert #123")
		assert.Contains(t, text, "Check: Web Check (#456)")
		assert.Contains(t, text, "Type: HTTP")
		assert.Contains(t, text, "Address: https://example.com")
		assert.Contains(t, text, "Location: New York")
		assert.Contains(t, text, "Server: nyc-server-1")
		assert.Contains(t, text, "Status: Resolved")
		assert.Contains(t, text, "Connection successful")
	})

	t.Run("shows down status for unresolved alerts", func(t *testing.T) {
		svc := newAlertsServiceMock(t)
		svc.EXPECT().Get(mock.Anything, upapi.PrimaryKey(123)).Return(&upapi.AlertItem{
			PK:                         123,
			CheckName:                  "Web Check",
			CheckPK:                    456,
			CheckMonitoringServiceType: "HTTP",
			CheckAddress:               "https://example.com",
			StateIsUp:                  false,
			Location:                   "New York",
			MonitoringServerName:       "nyc-server-1",
			CreatedAt:                  &createdAt,
		}, nil)

		client := newClientMock(t)
		client.EXPECT().Alerts().Return(svc)

		h := &alertsHandler{}
		ctx := testContext(t, client)
		result, _, err := h.HandleGetAlert(ctx, nil, getAlertInput{ID: 123})

		require.NoError(t, err)
		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "Status: Down")
	})

	t.Run("shows ignored status", func(t *testing.T) {
		svc := newAlertsServiceMock(t)
		svc.EXPECT().Get(mock.Anything, upapi.PrimaryKey(123)).Return(&upapi.AlertItem{
			PK:                         123,
			CheckName:                  "Web Check",
			CheckPK:                    456,
			CheckMonitoringServiceType: "HTTP",
			CheckAddress:               "https://example.com",
			StateIsUp:                  false,
			Ignored:                    true,
			Location:                   "New York",
			MonitoringServerName:       "nyc-server-1",
			CreatedAt:                  &createdAt,
		}, nil)

		client := newClientMock(t)
		client.EXPECT().Alerts().Return(svc)

		h := &alertsHandler{}
		ctx := testContext(t, client)
		result, _, err := h.HandleGetAlert(ctx, nil, getAlertInput{ID: 123})

		require.NoError(t, err)
		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "Ignored: yes")
	})

	t.Run("returns error when id is missing", func(t *testing.T) {
		client := newClientMock(t)

		h := &alertsHandler{}
		ctx := testContext(t, client)
		_, _, err := h.HandleGetAlert(ctx, nil, getAlertInput{})

		assert.ErrorContains(t, err, "id is required")
	})

	t.Run("returns error on service failure", func(t *testing.T) {
		svc := newAlertsServiceMock(t)
		svc.EXPECT().Get(mock.Anything, mock.Anything).Return((*upapi.AlertItem)(nil), assert.AnError)

		client := newClientMock(t)
		client.EXPECT().Alerts().Return(svc)

		h := &alertsHandler{}
		ctx := testContext(t, client)
		_, _, err := h.HandleGetAlert(ctx, nil, getAlertInput{ID: 123})

		assert.ErrorContains(t, err, "failed to get alert")
	})

	t.Run("returns error when no client in context", func(t *testing.T) {
		h := &alertsHandler{}
		_, _, err := h.HandleGetAlert(context.Background(), nil, getAlertInput{ID: 123})

		assert.ErrorContains(t, err, "no API client in context")
	})
}

func TestHandleIgnoreAlert(t *testing.T) {
	t.Run("toggles alert to ignored state", func(t *testing.T) {
		svc := newAlertsServiceMock(t)
		svc.EXPECT().Ignore(mock.Anything, upapi.PrimaryKey(123)).Return(&upapi.AlertItem{
			PK:      123,
			Ignored: true,
		}, nil)

		client := newClientMock(t)
		client.EXPECT().Alerts().Return(svc)

		h := &alertsHandler{}
		ctx := testContext(t, client)
		result, _, err := h.HandleIgnoreAlert(ctx, nil, ignoreAlertInput{ID: 123})

		require.NoError(t, err)
		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "Alert #123 is now ignored")
	})

	t.Run("toggles alert to not ignored state", func(t *testing.T) {
		svc := newAlertsServiceMock(t)
		svc.EXPECT().Ignore(mock.Anything, upapi.PrimaryKey(456)).Return(&upapi.AlertItem{
			PK:      456,
			Ignored: false,
		}, nil)

		client := newClientMock(t)
		client.EXPECT().Alerts().Return(svc)

		h := &alertsHandler{}
		ctx := testContext(t, client)
		result, _, err := h.HandleIgnoreAlert(ctx, nil, ignoreAlertInput{ID: 456})

		require.NoError(t, err)
		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "Alert #456 is now not ignored")
	})

	t.Run("returns error when id is missing", func(t *testing.T) {
		client := newClientMock(t)

		h := &alertsHandler{}
		ctx := testContext(t, client)
		_, _, err := h.HandleIgnoreAlert(ctx, nil, ignoreAlertInput{})

		assert.ErrorContains(t, err, "id is required")
	})

	t.Run("returns error on service failure", func(t *testing.T) {
		svc := newAlertsServiceMock(t)
		svc.EXPECT().Ignore(mock.Anything, mock.Anything).Return((*upapi.AlertItem)(nil), assert.AnError)

		client := newClientMock(t)
		client.EXPECT().Alerts().Return(svc)

		h := &alertsHandler{}
		ctx := testContext(t, client)
		_, _, err := h.HandleIgnoreAlert(ctx, nil, ignoreAlertInput{ID: 123})

		assert.ErrorContains(t, err, "failed to toggle alert ignore state")
	})

	t.Run("returns error when no client in context", func(t *testing.T) {
		h := &alertsHandler{}
		_, _, err := h.HandleIgnoreAlert(context.Background(), nil, ignoreAlertInput{ID: 123})

		assert.ErrorContains(t, err, "no API client in context")
	})
}

func TestHandleAlertResource(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	t.Run("returns alert resource by URI", func(t *testing.T) {
		svc := newAlertsServiceMock(t)
		svc.EXPECT().Get(mock.Anything, upapi.PrimaryKey(123)).Return(&upapi.AlertItem{
			PK:                         123,
			CheckName:                  "Web Check",
			CheckPK:                    456,
			CheckMonitoringServiceType: "HTTP",
			CheckAddress:               "https://example.com",
			StateIsUp:                  false,
			Location:                   "New York",
			MonitoringServerName:       "nyc-server-1",
			CreatedAt:                  &createdAt,
		}, nil)

		client := newClientMock(t)
		client.EXPECT().Alerts().Return(svc)

		h := &alertsHandler{}
		ctx := testContext(t, client)
		result, err := h.handleAlertResource(ctx, &mcp.ReadResourceRequest{
			Params: &mcp.ReadResourceParams{URI: "uptime://alerts/123"},
		})

		require.NoError(t, err)
		require.Len(t, result.Contents, 1)
		assert.Equal(t, "uptime://alerts/123", result.Contents[0].URI)
		assert.Equal(t, "text/plain", result.Contents[0].MIMEType)
		assert.Contains(t, result.Contents[0].Text, "Alert #123")
		assert.Contains(t, result.Contents[0].Text, "Check: Web Check (#456)")
	})

	t.Run("returns error for invalid URI", func(t *testing.T) {
		client := newClientMock(t)

		h := &alertsHandler{}
		ctx := testContext(t, client)
		_, err := h.handleAlertResource(ctx, &mcp.ReadResourceRequest{
			Params: &mcp.ReadResourceParams{URI: "invalid://uri"},
		})

		assert.ErrorContains(t, err, "invalid alert URI")
	})

	t.Run("returns error for non-numeric ID", func(t *testing.T) {
		client := newClientMock(t)

		h := &alertsHandler{}
		ctx := testContext(t, client)
		_, err := h.handleAlertResource(ctx, &mcp.ReadResourceRequest{
			Params: &mcp.ReadResourceParams{URI: "uptime://alerts/abc"},
		})

		assert.ErrorContains(t, err, "invalid alert ID")
	})

	t.Run("returns error on service failure", func(t *testing.T) {
		svc := newAlertsServiceMock(t)
		svc.EXPECT().Get(mock.Anything, mock.Anything).Return((*upapi.AlertItem)(nil), assert.AnError)

		client := newClientMock(t)
		client.EXPECT().Alerts().Return(svc)

		h := &alertsHandler{}
		ctx := testContext(t, client)
		_, err := h.handleAlertResource(ctx, &mcp.ReadResourceRequest{
			Params: &mcp.ReadResourceParams{URI: "uptime://alerts/123"},
		})

		assert.ErrorContains(t, err, "failed to get alert")
	})

	t.Run("returns error when no client in context", func(t *testing.T) {
		h := &alertsHandler{}
		_, err := h.handleAlertResource(context.Background(), &mcp.ReadResourceRequest{
			Params: &mcp.ReadResourceParams{URI: "uptime://alerts/123"},
		})

		assert.ErrorContains(t, err, "no API client in context")
	})
}
