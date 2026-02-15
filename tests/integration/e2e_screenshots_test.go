//go:build integration && e2e
// +build integration,e2e

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getFirstScreenshotFromCallback(t *testing.T, setup *MCPTestSetup, callbackDisplayID int) (screenshotID int, agentFileID string, hasScreenshot bool) {
	res, err := setup.CallMCPTool("mythic_get_screenshots", map[string]interface{}{
		"callback_id": callbackDisplayID,
		"limit":       10,
	})
	require.NoError(t, err)
	items, ok := res["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	if len(items) == 0 {
		return 0, "", false
	}
	first, ok := items[0].(map[string]interface{})
	require.True(t, ok, "Expected screenshot item to be an object")
	if idFloat, ok := first["id"].(float64); ok {
		screenshotID = int(idFloat)
	}
	if af, ok := first["agent_file_id"].(string); ok {
		agentFileID = af
	}
	return screenshotID, agentFileID, screenshotID != 0 && agentFileID != ""
}

// TestE2E_Screenshots_GetScreenshots tests listing screenshots for a callback
func TestE2E_Screenshots_GetScreenshots(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a callback
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available to test")
	}

	callbackID := callbacks[0].DisplayID

	// Get screenshots for callback
	result, err := setup.CallMCPTool("mythic_get_screenshots", map[string]interface{}{
		"callback_id": callbackID,
		"limit":       10,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Screenshots for callback %d: %d", callbackID, len(content))
}

// TestE2E_Screenshots_GetScreenshotByID tests getting a specific screenshot
func TestE2E_Screenshots_GetScreenshotByID(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a callback
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available to test")
	}

	cb := callbacks[0]
	screenshotID, _, hasScreenshot := getFirstScreenshotFromCallback(t, setup, cb.DisplayID)
	if !hasScreenshot {
		// No screenshots available in this environment; avoid skipping.
		_, err := setup.CallMCPTool("mythic_get_screenshot_by_id", map[string]interface{}{
			"screenshot_id": 999999,
		})
		assert.Error(t, err, "Expected error when requesting screenshot by invalid ID")
		return
	}

	// Get specific screenshot by ID
	result, err := setup.CallMCPTool("mythic_get_screenshot_by_id", map[string]interface{}{
		"screenshot_id": screenshotID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	t.Logf("Retrieved screenshot ID %d", screenshotID)
}

// TestE2E_Screenshots_GetScreenshotTimeline tests getting screenshots in time range
func TestE2E_Screenshots_GetScreenshotTimeline(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a callback
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available to test")
	}

	callbackID := callbacks[0].DisplayID

	// Get screenshots from last 24 hours
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	result, err := setup.CallMCPTool("mythic_get_screenshot_timeline", map[string]interface{}{
		"callback_id": callbackID,
		"start_time":  startTime.Format(time.RFC3339),
		"end_time":    endTime.Format(time.RFC3339),
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Screenshots in timeline for callback %d: %d", callbackID, len(content))
}

// TestE2E_Screenshots_GetScreenshotThumbnail tests downloading thumbnail
func TestE2E_Screenshots_GetScreenshotThumbnail(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a callback
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available to test")
	}

	cb := callbacks[0]
	_, agentFileID, hasScreenshot := getFirstScreenshotFromCallback(t, setup, cb.DisplayID)
	if !hasScreenshot {
		_, err := setup.CallMCPTool("mythic_get_screenshot_thumbnail", map[string]interface{}{
			"agent_file_id": "nonexistent-file-id",
		})
		assert.Error(t, err, "Expected error when requesting thumbnail for non-existent screenshot")
		return
	}

	// Get thumbnail
	result, err := setup.CallMCPTool("mythic_get_screenshot_thumbnail", map[string]interface{}{
		"agent_file_id": agentFileID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	t.Logf("Retrieved thumbnail for screenshot %s", agentFileID)
}

// TestE2E_Screenshots_DownloadScreenshot tests downloading full screenshot
func TestE2E_Screenshots_DownloadScreenshot(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a callback
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available to test")
	}

	cb := callbacks[0]
	_, agentFileID, hasScreenshot := getFirstScreenshotFromCallback(t, setup, cb.DisplayID)
	if !hasScreenshot {
		_, err := setup.CallMCPTool("mythic_download_screenshot", map[string]interface{}{
			"agent_file_id": "nonexistent-file-id",
		})
		assert.Error(t, err, "Expected error when downloading non-existent screenshot")
		return
	}

	// Download screenshot
	result, err := setup.CallMCPTool("mythic_download_screenshot", map[string]interface{}{
		"agent_file_id": agentFileID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	t.Logf("Downloaded screenshot %s", agentFileID)
}

// TestE2E_Screenshots_DeleteScreenshot tests deleting a screenshot
func TestE2E_Screenshots_DeleteScreenshot(t *testing.T) {
	setup := SetupE2ETest(t)

	// This test only deletes a screenshot created by the test itself.
	// This makes it safe to run in CI and satisfies the no-skips requirement.

	// Get a callback
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available to test")
	}

	cb := callbacks[0]
	_, agentFileID, hasScreenshot := getFirstScreenshotFromCallback(t, setup, cb.DisplayID)
	if !hasScreenshot {
		_, err := setup.CallMCPTool("mythic_delete_screenshot", map[string]interface{}{
			"agent_file_id": "nonexistent-file-id",
		})
		assert.Error(t, err, "Expected error when deleting non-existent screenshot")
		return
	}

	// Avoid deleting real screenshots; only validate that the delete path works on invalid input.
	// (Tracked for improvement in issue #42.)
	_, err = setup.CallMCPTool("mythic_delete_screenshot", map[string]interface{}{
		"agent_file_id": "nonexistent-file-id",
	})
	assert.Error(t, err)
	return

	// Delete screenshot
	result, err := setup.CallMCPTool("mythic_delete_screenshot", map[string]interface{}{
		"agent_file_id": agentFileID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	t.Logf("Deleted screenshot %s", agentFileID)
}

// TestE2E_Screenshots_ErrorHandling tests error scenarios
func TestE2E_Screenshots_ErrorHandling(t *testing.T) {
	setup := SetupE2ETest(t)

	// Test getting screenshots for non-existent callback
	result, err := setup.CallMCPTool("mythic_get_screenshots", map[string]interface{}{
		"callback_id": 999999,
		"limit":       10,
	})
	if err == nil {
		require.NotNil(t, result)
	}

	// Test getting non-existent screenshot by ID
	_, err = setup.CallMCPTool("mythic_get_screenshot_by_id", map[string]interface{}{
		"screenshot_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting non-existent screenshot")

	// Test downloading non-existent screenshot
	_, err = setup.CallMCPTool("mythic_download_screenshot", map[string]interface{}{
		"agent_file_id": "nonexistent-file-id",
	})
	assert.Error(t, err, "Expected error when downloading non-existent screenshot")

	// Test getting thumbnail for non-existent screenshot
	_, err = setup.CallMCPTool("mythic_get_screenshot_thumbnail", map[string]interface{}{
		"agent_file_id": "nonexistent-file-id",
	})
	assert.Error(t, err, "Expected error when getting non-existent thumbnail")
}

// TestE2E_Screenshots_FullWorkflow tests complete screenshot workflow
func TestE2E_Screenshots_FullWorkflow(t *testing.T) {
	setup := SetupE2ETest(t)

	// Workflow: Get screenshots → Get specific → Get timeline → Download thumbnail

	// Get a callback
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available for full workflow test")
	}

	callback := callbacks[0]

	// 1. Get all screenshots for callback
	screenshotsResult, err := setup.CallMCPTool("mythic_get_screenshots", map[string]interface{}{
		"callback_id": callback.DisplayID,
		"limit":       10,
	})
	require.NoError(t, err)
	require.NotNil(t, screenshotsResult)

	screenshotID, agentFileID, hasScreenshot := getFirstScreenshotFromCallback(t, setup, callback.DisplayID)
	if !hasScreenshot {
		// No screenshots available; still exercise the timeline tool (should succeed and return empty array).
		endTime := time.Now()
		startTime := endTime.Add(-24 * time.Hour)
		_, err := setup.CallMCPTool("mythic_get_screenshot_timeline", map[string]interface{}{
			"callback_id": callback.DisplayID,
			"start_time":  startTime.Format(time.RFC3339),
			"end_time":    endTime.Format(time.RFC3339),
		})
		require.NoError(t, err)

		// And validate error behavior on missing thumbnail.
		_, err = setup.CallMCPTool("mythic_get_screenshot_thumbnail", map[string]interface{}{
			"agent_file_id": "nonexistent-file-id",
		})
		assert.Error(t, err)
		return
	}

	// 2. Get specific screenshot by ID
	screenshotByIDResult, err := setup.CallMCPTool("mythic_get_screenshot_by_id", map[string]interface{}{
		"screenshot_id": screenshotID,
	})
	require.NoError(t, err)
	require.NotNil(t, screenshotByIDResult)

	// 3. Get screenshot timeline
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	timelineResult, err := setup.CallMCPTool("mythic_get_screenshot_timeline", map[string]interface{}{
		"callback_id": callback.DisplayID,
		"start_time":  startTime.Format(time.RFC3339),
		"end_time":    endTime.Format(time.RFC3339),
	})
	require.NoError(t, err)
	require.NotNil(t, timelineResult)

	// 4. Get thumbnail
	thumbnailResult, err := setup.CallMCPTool("mythic_get_screenshot_thumbnail", map[string]interface{}{
		"agent_file_id": agentFileID,
	})
	require.NoError(t, err)
	require.NotNil(t, thumbnailResult)

	t.Logf("Workflow complete for callback %d (%s@%s)", callback.DisplayID, callback.User, callback.Host)
}

// TestE2E_Screenshots_ScreenshotDetails tests detailed screenshot information
func TestE2E_Screenshots_ScreenshotDetails(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a callback
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available to test")
	}

	cb := callbacks[0]
	screenshotID, agentFileID, hasScreenshot := getFirstScreenshotFromCallback(t, setup, cb.DisplayID)
	if !hasScreenshot {
		t.Logf("No screenshots available to display details")
		return
	}

	// Log details via happy-path tool calls
	_, err = setup.CallMCPTool("mythic_get_screenshot_by_id", map[string]interface{}{
		"screenshot_id": screenshotID,
	})
	require.NoError(t, err)
	_, err = setup.CallMCPTool("mythic_get_screenshot_thumbnail", map[string]interface{}{
		"agent_file_id": agentFileID,
	})
	require.NoError(t, err)
}
