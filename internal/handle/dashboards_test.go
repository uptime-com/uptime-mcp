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

func TestHandleListDashboards(t *testing.T) {
	t.Run("returns formatted list of dashboards", func(t *testing.T) {
		svc := newDashboardsServiceMock(t)
		svc.EXPECT().List(mock.Anything, mock.Anything).Return(&upapi.ListResult[upapi.Dashboard]{
			TotalCount: 2,
			Items: []upapi.Dashboard{
				{PK: 1, Name: "Production", IsPinned: true},
				{PK: 2, Name: "Staging", IsPinned: false},
			},
		}, nil)

		client := newClientMock(t)
		client.EXPECT().Dashboards().Return(svc)

		h := &dashboardsHandler{}
		ctx := testContext(t, client)
		result, _, err := h.HandleListDashboards(ctx, nil, listDashboardsInput{})

		require.NoError(t, err)
		require.Len(t, result.Content, 1)
		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "Found 2 results")
		assert.Contains(t, text, "[1] Production (pinned)")
		assert.Contains(t, text, "[2] Staging")
		assert.NotContains(t, text, "Staging (pinned)")
	})

	t.Run("returns error on service failure", func(t *testing.T) {
		svc := newDashboardsServiceMock(t)
		svc.EXPECT().List(mock.Anything, mock.Anything).Return((*upapi.ListResult[upapi.Dashboard])(nil), assert.AnError)

		client := newClientMock(t)
		client.EXPECT().Dashboards().Return(svc)

		h := &dashboardsHandler{}
		ctx := testContext(t, client)
		_, _, err := h.HandleListDashboards(ctx, nil, listDashboardsInput{})

		assert.ErrorContains(t, err, "failed to list dashboards")
	})

	t.Run("returns error when no client in context", func(t *testing.T) {
		h := &dashboardsHandler{}
		_, _, err := h.HandleListDashboards(context.Background(), nil, listDashboardsInput{})

		assert.ErrorContains(t, err, "no API client in context")
	})
}

func TestHandleGetDashboard(t *testing.T) {
	t.Run("returns formatted dashboard details", func(t *testing.T) {
		svc := newDashboardsServiceMock(t)
		svc.EXPECT().Get(mock.Anything, upapi.PrimaryKey(1)).Return(&upapi.Dashboard{
			PK:                  1,
			Name:                "Production",
			IsPinned:            true,
			ServicesTags:        []string{"production"},
			ServicesShowSection: true,
			AlertsShowSection:   true,
		}, nil)

		client := newClientMock(t)
		client.EXPECT().Dashboards().Return(svc)

		h := &dashboardsHandler{}
		ctx := testContext(t, client)
		result, _, err := h.HandleGetDashboard(ctx, nil, getDashboardInput{ID: 1})

		require.NoError(t, err)
		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "Dashboard #1")
		assert.Contains(t, text, "Name: Production")
		assert.Contains(t, text, "Pinned: true")
		assert.Contains(t, text, "Service tags: production")
	})

	t.Run("returns error when id is missing", func(t *testing.T) {
		client := newClientMock(t)

		h := &dashboardsHandler{}
		ctx := testContext(t, client)
		_, _, err := h.HandleGetDashboard(ctx, nil, getDashboardInput{})

		assert.ErrorContains(t, err, "id is required")
	})

	t.Run("returns error on service failure", func(t *testing.T) {
		svc := newDashboardsServiceMock(t)
		svc.EXPECT().Get(mock.Anything, mock.Anything).Return((*upapi.Dashboard)(nil), assert.AnError)

		client := newClientMock(t)
		client.EXPECT().Dashboards().Return(svc)

		h := &dashboardsHandler{}
		ctx := testContext(t, client)
		_, _, err := h.HandleGetDashboard(ctx, nil, getDashboardInput{ID: 1})

		assert.ErrorContains(t, err, "failed to get dashboard")
	})

	t.Run("returns error when no client in context", func(t *testing.T) {
		h := &dashboardsHandler{}
		_, _, err := h.HandleGetDashboard(context.Background(), nil, getDashboardInput{ID: 1})

		assert.ErrorContains(t, err, "no API client in context")
	})
}

func TestHandleCreateDashboard(t *testing.T) {
	t.Run("creates dashboard with tags", func(t *testing.T) {
		svc := newDashboardsServiceMock(t)
		svc.EXPECT().Create(mock.Anything, mock.MatchedBy(func(d upapi.Dashboard) bool {
			return d.Name == "Production" && len(d.ServicesTags) == 1 && d.ServicesTags[0] == "prod"
		})).Return(&upapi.Dashboard{PK: 1, Name: "Production"}, nil)

		client := newClientMock(t)
		client.EXPECT().Dashboards().Return(svc)

		h := &dashboardsHandler{}
		ctx := testContext(t, client)
		result, _, err := h.HandleCreateDashboard(ctx, nil, createDashboardInput{
			Name:         "Production",
			ServicesTags: []string{"prod"},
		})

		require.NoError(t, err)
		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "Created dashboard #1: Production")
	})

	t.Run("returns error when name is missing", func(t *testing.T) {
		client := newClientMock(t)

		h := &dashboardsHandler{}
		ctx := testContext(t, client)
		_, _, err := h.HandleCreateDashboard(ctx, nil, createDashboardInput{})

		assert.ErrorContains(t, err, "name is required")
	})

	t.Run("returns error on service failure", func(t *testing.T) {
		svc := newDashboardsServiceMock(t)
		svc.EXPECT().Create(mock.Anything, mock.Anything).Return((*upapi.Dashboard)(nil), assert.AnError)

		client := newClientMock(t)
		client.EXPECT().Dashboards().Return(svc)

		h := &dashboardsHandler{}
		ctx := testContext(t, client)
		_, _, err := h.HandleCreateDashboard(ctx, nil, createDashboardInput{Name: "Test"})

		assert.ErrorContains(t, err, "failed to create dashboard")
	})
}

func TestHandleUpdateDashboard(t *testing.T) {
	t.Run("updates dashboard name", func(t *testing.T) {
		svc := newDashboardsServiceMock(t)
		svc.EXPECT().Get(mock.Anything, upapi.PrimaryKey(1)).Return(&upapi.Dashboard{
			PK:   1,
			Name: "Old Name",
		}, nil)
		svc.EXPECT().Update(mock.Anything, upapi.PrimaryKey(1), mock.MatchedBy(func(d upapi.Dashboard) bool {
			return d.Name == "New Name"
		})).Return(&upapi.Dashboard{PK: 1, Name: "New Name"}, nil)

		client := newClientMock(t)
		client.EXPECT().Dashboards().Return(svc)

		h := &dashboardsHandler{}
		ctx := testContext(t, client)
		result, _, err := h.HandleUpdateDashboard(ctx, nil, updateDashboardInput{ID: 1, Name: "New Name"})

		require.NoError(t, err)
		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "Updated dashboard #1: New Name")
	})

	t.Run("updates boolean fields via pointers", func(t *testing.T) {
		pinned := true
		svc := newDashboardsServiceMock(t)
		svc.EXPECT().Get(mock.Anything, upapi.PrimaryKey(1)).Return(&upapi.Dashboard{
			PK:       1,
			Name:     "Dashboard",
			IsPinned: false,
		}, nil)
		svc.EXPECT().Update(mock.Anything, upapi.PrimaryKey(1), mock.MatchedBy(func(d upapi.Dashboard) bool {
			return d.IsPinned == true
		})).Return(&upapi.Dashboard{PK: 1, Name: "Dashboard", IsPinned: true}, nil)

		client := newClientMock(t)
		client.EXPECT().Dashboards().Return(svc)

		h := &dashboardsHandler{}
		ctx := testContext(t, client)
		_, _, err := h.HandleUpdateDashboard(ctx, nil, updateDashboardInput{ID: 1, IsPinned: &pinned})

		require.NoError(t, err)
	})

	t.Run("returns error when id is missing", func(t *testing.T) {
		client := newClientMock(t)

		h := &dashboardsHandler{}
		ctx := testContext(t, client)
		_, _, err := h.HandleUpdateDashboard(ctx, nil, updateDashboardInput{})

		assert.ErrorContains(t, err, "id is required")
	})

	t.Run("returns error on get failure", func(t *testing.T) {
		svc := newDashboardsServiceMock(t)
		svc.EXPECT().Get(mock.Anything, mock.Anything).Return((*upapi.Dashboard)(nil), assert.AnError)

		client := newClientMock(t)
		client.EXPECT().Dashboards().Return(svc)

		h := &dashboardsHandler{}
		ctx := testContext(t, client)
		_, _, err := h.HandleUpdateDashboard(ctx, nil, updateDashboardInput{ID: 1, Name: "X"})

		assert.ErrorContains(t, err, "failed to get dashboard")
	})
}

func TestHandleDeleteDashboard(t *testing.T) {
	t.Run("deletes dashboard", func(t *testing.T) {
		svc := newDashboardsServiceMock(t)
		svc.EXPECT().Delete(mock.Anything, upapi.PrimaryKey(1)).Return(nil)

		client := newClientMock(t)
		client.EXPECT().Dashboards().Return(svc)

		h := &dashboardsHandler{}
		ctx := testContext(t, client)
		result, _, err := h.HandleDeleteDashboard(ctx, nil, deleteDashboardInput{ID: 1})

		require.NoError(t, err)
		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "Successfully deleted dashboard #1")
	})

	t.Run("returns error when id is missing", func(t *testing.T) {
		client := newClientMock(t)

		h := &dashboardsHandler{}
		ctx := testContext(t, client)
		_, _, err := h.HandleDeleteDashboard(ctx, nil, deleteDashboardInput{})

		assert.ErrorContains(t, err, "id is required")
	})

	t.Run("returns error on service failure", func(t *testing.T) {
		svc := newDashboardsServiceMock(t)
		svc.EXPECT().Delete(mock.Anything, mock.Anything).Return(assert.AnError)

		client := newClientMock(t)
		client.EXPECT().Dashboards().Return(svc)

		h := &dashboardsHandler{}
		ctx := testContext(t, client)
		_, _, err := h.HandleDeleteDashboard(ctx, nil, deleteDashboardInput{ID: 1})

		assert.ErrorContains(t, err, "failed to delete dashboard")
	})
}

func TestHandleDashboardResource(t *testing.T) {
	t.Run("returns dashboard resource by URI", func(t *testing.T) {
		svc := newDashboardsServiceMock(t)
		svc.EXPECT().Get(mock.Anything, upapi.PrimaryKey(1)).Return(&upapi.Dashboard{
			PK:       1,
			Name:     "Production",
			IsPinned: true,
		}, nil)

		client := newClientMock(t)
		client.EXPECT().Dashboards().Return(svc)

		h := &dashboardsHandler{}
		ctx := testContext(t, client)
		result, err := h.handleDashboardResource(ctx, &mcp.ReadResourceRequest{
			Params: &mcp.ReadResourceParams{URI: "uptime://dashboards/1"},
		})

		require.NoError(t, err)
		require.Len(t, result.Contents, 1)
		assert.Equal(t, "uptime://dashboards/1", result.Contents[0].URI)
		assert.Contains(t, result.Contents[0].Text, "Dashboard #1")
		assert.Contains(t, result.Contents[0].Text, "Name: Production")
	})

	t.Run("returns error for invalid URI", func(t *testing.T) {
		client := newClientMock(t)

		h := &dashboardsHandler{}
		ctx := testContext(t, client)
		_, err := h.handleDashboardResource(ctx, &mcp.ReadResourceRequest{
			Params: &mcp.ReadResourceParams{URI: "invalid://uri"},
		})

		assert.ErrorContains(t, err, "invalid dashboard URI")
	})

	t.Run("returns error for non-numeric ID", func(t *testing.T) {
		client := newClientMock(t)

		h := &dashboardsHandler{}
		ctx := testContext(t, client)
		_, err := h.handleDashboardResource(ctx, &mcp.ReadResourceRequest{
			Params: &mcp.ReadResourceParams{URI: "uptime://dashboards/abc"},
		})

		assert.ErrorContains(t, err, "invalid dashboard ID")
	})
}
