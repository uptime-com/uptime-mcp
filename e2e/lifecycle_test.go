//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// TestE2E_CheckLifecycle tests the full lifecycle:
// 1. Create contact
// 2. Create tag
// 3. Confirm tag via resource
// 4. Create checks with that tag and contact
// 5. List and confirm checks
// 6. Delete checks
// 7. Delete tag
// 8. Delete contact
func TestE2E_CheckLifecycle(t *testing.T) {
	session := makeClientSession(t)
	ctx := context.Background()

	// Generate unique names to avoid conflicts
	suffix := time.Now().Unix()
	contactName := fmt.Sprintf("e2e-contact-%d", suffix)
	tagName := fmt.Sprintf("e2e-tag-%d", suffix)

	// 1. Create contact
	t.Log("Step 1: Creating contact...")
	contactResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "create_contact",
		Arguments: map[string]any{
			"name":       contactName,
			"email_list": []string{"noreply@uptime.com"},
		},
	})
	require.NoError(t, err)
	if contactResult.IsError {
		errText := contactResult.Content[0].(*mcp.TextContent).Text
		t.Fatalf("create_contact failed: %s", errText)
	}

	contactText := contactResult.Content[0].(*mcp.TextContent).Text
	t.Logf("Contact created: %s", contactText)

	contactID := extractID(t, contactText, `#(\d+)`)
	require.NotZero(t, contactID, "failed to extract contact ID")

	// Ensure contact cleanup (last in cleanup order)
	defer func() {
		t.Log("Cleanup: Deleting contact...")
		_, _ = session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "delete_contact",
			Arguments: map[string]any{"id": contactID},
		})
	}()

	// 2. Create tag
	t.Log("Step 2: Creating tag...")
	tagResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "create_tag",
		Arguments: map[string]any{
			"name":  tagName,
			"color": "#FF5733",
		},
	})
	require.NoError(t, err)
	require.False(t, tagResult.IsError)

	tagText := tagResult.Content[0].(*mcp.TextContent).Text
	t.Logf("Tag created: %s", tagText)

	tagID := extractID(t, tagText, `#(\d+)`)
	require.NotZero(t, tagID, "failed to extract tag ID")

	// Ensure tag cleanup
	defer func() {
		t.Log("Cleanup: Deleting tag...")
		_, _ = session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "delete_tag",
			Arguments: map[string]any{"id": tagID},
		})
	}()

	// 3. Confirm tag via resource
	t.Log("Step 3: Confirming tag via resource...")
	tagURI := fmt.Sprintf("https://uptime.com/api/v1/check-tags/%d", tagID)
	tagResource, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: tagURI})
	require.NoError(t, err)
	require.Contains(t, tagResource.Contents[0].Text, tagName)
	t.Logf("Tag resource confirmed: %s", tagResource.Contents[0].Text)

	// 4. Create checks with that tag
	t.Log("Step 4: Creating HTTP checks...")
	checkIDs := make([]int, 0, 2)

	for i := 1; i <= 2; i++ {
		checkName := fmt.Sprintf("e2e-check-%d-%d", suffix, i)
		checkResult, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "create_http_check",
			Arguments: map[string]any{
				"name":           checkName,
				"address":        fmt.Sprintf("https://example.com/e2e-test-%d", i),
				"tags":           []string{tagName},
				"contact_groups": []string{contactName},
				"locations":      []string{"Germany-Frankfurt"},
			},
		})
		require.NoError(t, err)
		if checkResult.IsError {
			errText := checkResult.Content[0].(*mcp.TextContent).Text
			t.Fatalf("check creation failed: %s", errText)
		}

		checkText := checkResult.Content[0].(*mcp.TextContent).Text
		t.Logf("Check %d created: %s", i, checkText)

		checkID := extractID(t, checkText, `#(\d+)`)
		require.NotZero(t, checkID, "failed to extract check ID")
		checkIDs = append(checkIDs, checkID)
	}

	// Ensure checks cleanup (in reverse order, before tag)
	defer func() {
		t.Log("Cleanup: Deleting checks...")
		for _, id := range checkIDs {
			_, _ = session.CallTool(ctx, &mcp.CallToolParams{
				Name:      "delete_check",
				Arguments: map[string]any{"id": id},
			})
		}
	}()

	// 5. List and confirm checks with tag filter
	t.Log("Step 5: Listing checks by tag...")
	listResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "list_checks",
		Arguments: map[string]any{
			"tag":       tagName,
			"page_size": 10,
		},
	})
	require.NoError(t, err)

	listText := listResult.Content[0].(*mcp.TextContent).Text
	t.Logf("List result:\n%s", listText)

	// Verify both checks appear in the list
	for _, id := range checkIDs {
		require.Contains(t, listText, fmt.Sprintf("[%d]", id), "check %d not found in list", id)
	}

	// 6. Confirm check via resource
	t.Log("Step 6: Confirming check via resource...")
	checkURI := fmt.Sprintf("https://uptime.com/api/v1/checks/%d", checkIDs[0])
	checkResource, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: checkURI})
	require.NoError(t, err)
	require.Contains(t, checkResource.Contents[0].Text, tagName)
	t.Logf("Check resource confirmed")

	// 7. Delete checks
	t.Log("Step 7: Deleting checks...")
	for _, id := range checkIDs {
		delResult, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "delete_check",
			Arguments: map[string]any{"id": id},
		})
		require.NoError(t, err)
		require.False(t, delResult.IsError)
		t.Logf("Deleted check #%d", id)
	}
	checkIDs = nil // Clear so defer doesn't try again

	// 8. Delete tag
	t.Log("Step 8: Deleting tag...")
	delTagResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "delete_tag",
		Arguments: map[string]any{"id": tagID},
	})
	require.NoError(t, err)
	require.False(t, delTagResult.IsError)
	t.Logf("Deleted tag #%d", tagID)
	tagID = 0 // Clear so defer doesn't try again

	// 9. Delete contact
	t.Log("Step 9: Deleting contact...")
	delContactResult, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "delete_contact",
		Arguments: map[string]any{"id": contactID},
	})
	require.NoError(t, err)
	require.False(t, delContactResult.IsError)
	t.Logf("Deleted contact #%d", contactID)
	contactID = 0 // Clear so defer doesn't try again

	t.Log("Lifecycle test completed successfully!")
}

// extractID extracts the first numeric ID from text using the given regex pattern.
func extractID(t *testing.T, text, pattern string) int {
	t.Helper()
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) < 2 {
		return 0
	}
	id, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}
	return id
}
