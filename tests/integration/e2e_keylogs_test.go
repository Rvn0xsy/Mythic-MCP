//go:build integration && e2e
// +build integration,e2e

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_Keylogs_GetKeylogs tests listing all keylogs
func TestE2E_Keylogs_GetKeylogs(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all keylogs
	result, err := setup.CallMCPTool("mythic_get_keylogs", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Result should be an array (may be empty if no keylogs)
	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Total keylogs: %d", len(content))
}

// TestE2E_Keylogs_GetKeylogsByOperation tests listing keylogs by operation
func TestE2E_Keylogs_GetKeylogsByOperation(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get current operation
	me, err := setup.MythicClient.GetMe(setup.Ctx)
	require.NoError(t, err)

	if me.CurrentOperationID == nil {
		t.Skip("No current operation set")
	}

	operationID := *me.CurrentOperationID

	// Get keylogs for operation
	result, err := setup.CallMCPTool("mythic_get_keylogs_by_operation", map[string]interface{}{
		"operation_id": operationID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Keylogs in operation %d: %d", operationID, len(content))
}

// TestE2E_Keylogs_GetKeylogsByCallback tests listing keylogs by callback
func TestE2E_Keylogs_GetKeylogsByCallback(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a callback
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available to test")
	}

	callbackID := callbacks[0].DisplayID

	// Get keylogs for callback
	result, err := setup.CallMCPTool("mythic_get_keylogs_by_callback", map[string]interface{}{
		"callback_id": callbackID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Keylogs for callback %d: %d", callbackID, len(content))
}

// TestE2E_Keylogs_ErrorHandling tests error scenarios
func TestE2E_Keylogs_ErrorHandling(t *testing.T) {
	setup := SetupE2ETest(t)

	// Test getting keylogs for non-existent operation
	_, err := setup.CallMCPTool("mythic_get_keylogs_by_operation", map[string]interface{}{
		"operation_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting keylogs for non-existent operation")

	// Test getting keylogs for non-existent callback
	_, err = setup.CallMCPTool("mythic_get_keylogs_by_callback", map[string]interface{}{
		"callback_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting keylogs for non-existent callback")
}

// TestE2E_Keylogs_FullWorkflow tests complete keylog workflow
func TestE2E_Keylogs_FullWorkflow(t *testing.T) {
	setup := SetupE2ETest(t)

	// Workflow: Get all keylogs → Get by operation → Get by callback

	// 1. Get all keylogs
	allKeylogsResult, err := setup.CallMCPTool("mythic_get_keylogs", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, allKeylogsResult)

	// 2. Get current operation
	me, err := setup.MythicClient.GetMe(setup.Ctx)
	require.NoError(t, err)

	if me.CurrentOperationID == nil {
		t.Skip("No current operation set for full workflow test")
	}

	operationID := *me.CurrentOperationID

	// 3. Get keylogs by operation
	operationKeylogsResult, err := setup.CallMCPTool("mythic_get_keylogs_by_operation", map[string]interface{}{
		"operation_id": operationID,
	})
	require.NoError(t, err)
	require.NotNil(t, operationKeylogsResult)

	// Get a callback to work with
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available for full workflow test")
	}

	callback := callbacks[0]

	// 4. Get keylogs by callback
	callbackKeylogsResult, err := setup.CallMCPTool("mythic_get_keylogs_by_callback", map[string]interface{}{
		"callback_id": callback.DisplayID,
	})
	require.NoError(t, err)
	require.NotNil(t, callbackKeylogsResult)

	t.Logf("Workflow complete for callback %d (%s@%s)", callback.DisplayID, callback.User, callback.Host)
}

// TestE2E_Keylogs_KeylogDetails tests detailed keylog information
func TestE2E_Keylogs_KeylogDetails(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all keylogs
	keylogs, err := setup.MythicClient.GetKeylogs(setup.Ctx)
	require.NoError(t, err)

	if len(keylogs) == 0 {
		t.Skip("No keylogs available to test")
	}

	// Log details for first few keylogs
	for i, keylog := range keylogs {
		if i >= 3 {
			break
		}

		t.Logf("Keylog %d:", keylog.ID)
		t.Logf("  - Window: %s", keylog.Window)
		t.Logf("  - User: %s", keylog.User)
		t.Logf("  - Keystrokes: %s", keylog.Keystrokes)
		t.Logf("  - Task ID: %d", keylog.TaskID)
	}
}
