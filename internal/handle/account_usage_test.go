package handle

import (
	"fmt"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/uptime-com/uptime-client-go/v2/pkg/upapi"
)

func TestHandleGetAccountUsage(t *testing.T) {
	t.Run("returns formatted usage data", func(t *testing.T) {
		usage := upapi.AccountUsage{
			"Account":                    "Test",
			"Checks Used":                float64(18),
			"Checks Allocated":           float64(100),
			"Content Matching Available": true,
		}

		svc := newAccountUsageServiceMock(t)
		svc.EXPECT().Get(mock.Anything).Return(&usage, nil)

		client := newClientMock(t)
		client.EXPECT().AccountUsage().Return(svc)

		h := &accountUsageHandler{}
		ctx := testContext(t, client)
		result, _, err := h.handleGetAccountUsage(ctx, nil, getAccountUsageInput{})

		require.NoError(t, err)
		require.Len(t, result.Content, 1)
		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "Account: Test")
		assert.Contains(t, text, "Checks Used: 18")
		assert.Contains(t, text, "Checks Allocated: 100")
		assert.Contains(t, text, "Content Matching Available: true")
	})

	t.Run("returns error on API failure", func(t *testing.T) {
		svc := newAccountUsageServiceMock(t)
		svc.EXPECT().Get(mock.Anything).Return(nil, fmt.Errorf("unauthorized"))

		client := newClientMock(t)
		client.EXPECT().AccountUsage().Return(svc)

		h := &accountUsageHandler{}
		ctx := testContext(t, client)
		_, _, err := h.handleGetAccountUsage(ctx, nil, getAccountUsageInput{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})
}
