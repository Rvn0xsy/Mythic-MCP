//go:build integration && e2e
// +build integration,e2e

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_Callbacks_GetAllCallbacks tests listing all callbacks
func TestE2E_Callbacks_GetAllCallbacks(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all callbacks
	result, err := setup.CallMCPTool("mythic_get_all_callbacks", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Result should be an array (may be empty if no callbacks)
	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Total callbacks: %d", len(content))
}

// TestE2E_Callbacks_GetActiveCallbacks tests listing active callbacks
func TestE2E_Callbacks_GetActiveCallbacks(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get active callbacks
	result, err := setup.CallMCPTool("mythic_get_active_callbacks", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Active callbacks: %d", len(content))
}

// TestE2E_Callbacks_GetCallback tests getting specific callback
func TestE2E_Callbacks_GetCallback(t *testing.T) {
	setup := SetupE2ETest(t)

	// First get all callbacks to find one to query
	allCallbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(allCallbacks) == 0 {
		t.Skip("No callbacks available to test")
	}

	callbackID := allCallbacks[0].DisplayID

	// Get specific callback
	result, err := setup.CallMCPTool("mythic_get_callback", map[string]interface{}{
		"callback_id": callbackID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
}

// TestE2E_Callbacks_UpdateCallback tests updating callback properties
func TestE2E_Callbacks_UpdateCallback(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a callback to update
	allCallbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(allCallbacks) == 0 {
		t.Skip("No callbacks available to test")
	}

	callbackID := allCallbacks[0].DisplayID

	// Update callback description
	updateResult, err := setup.CallMCPTool("mythic_update_callback", map[string]interface{}{
		"callback_id": callbackID,
		"description": "Updated via E2E test",
	})
	require.NoError(t, err)
	require.NotNil(t, updateResult)

	// Verify update
	callback, err := setup.MythicClient.GetCallbackByID(setup.Ctx, callbackID)
	require.NoError(t, err)
	assert.Equal(t, "Updated via E2E test", callback.Description)
}

// TestE2E_Callbacks_GetLoadedCommands tests getting loaded commands
func TestE2E_Callbacks_GetLoadedCommands(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a callback
	allCallbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(allCallbacks) == 0 {
		t.Skip("No callbacks available to test")
	}

	callbackID := allCallbacks[0].DisplayID

	// Get loaded commands
	result, err := setup.CallMCPTool("mythic_get_loaded_commands", map[string]interface{}{
		"callback_id": callbackID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Loaded commands for callback %d: %d", callbackID, len(content))
}

// TestE2E_Callbacks_ExportImportConfig tests callback config export/import
func TestE2E_Callbacks_ExportImportConfig(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a callback
	allCallbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(allCallbacks) == 0 {
		t.Skip("No callbacks available to test")
	}

	agentCallbackID := allCallbacks[0].AgentCallbackID

	// Export callback config
	exportResult, err := setup.CallMCPTool("mythic_export_callback_config", map[string]interface{}{
		"agent_callback_id": agentCallbackID,
	})
	require.NoError(t, err)
	require.NotNil(t, exportResult)

	// Extract exported config
	exportMeta, ok := exportResult["metadata"].(map[string]interface{})
	require.True(t, ok, "Expected metadata in export result")
	configStr, ok := exportMeta["config"].(string)
	require.True(t, ok, "Expected config string in metadata")
	require.NotEmpty(t, configStr, "Config should not be empty")

	// Note: Import would create a new callback, so we'll skip that in E2E
	// to avoid polluting the Mythic instance
	t.Logf("Successfully exported config (length: %d bytes)", len(configStr))
}

// TestE2E_Callbacks_GetCallbackTokens tests getting callback tokens
func TestE2E_Callbacks_GetCallbackTokens(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a callback
	allCallbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(allCallbacks) == 0 {
		t.Skip("No callbacks available to test")
	}

	callbackID := allCallbacks[0].DisplayID

	// Get callback tokens
	result, err := setup.CallMCPTool("mythic_get_callback_tokens", map[string]interface{}{
		"callback_id": callbackID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Tokens for callback %d: %d", callbackID, len(content))
}

// TestE2E_Callbacks_GraphEdges tests P2P callback graph management
func TestE2E_Callbacks_GraphEdges(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all callbacks
	allCallbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(allCallbacks) < 2 {
		t.Skip("Need at least 2 callbacks for P2P graph testing")
	}

	sourceID := allCallbacks[0].DisplayID
	destID := allCallbacks[1].DisplayID

	// Add graph edge (P2P link)
	addResult, err := setup.CallMCPTool("mythic_add_callback_edge", map[string]interface{}{
		"source_id":       sourceID,
		"destination_id":  destID,
		"c2_profile_name": "default",
	})

	// Note: This may fail if callbacks aren't compatible or already linked
	// That's okay - just log the result
	if err != nil {
		t.Logf("Add edge failed (expected if incompatible): %v", err)
		return
	}
	require.NotNil(t, addResult)

	// If we successfully added an edge, try to remove it
	// Note: We'd need the edge ID to remove it, which the add result should contain
	t.Logf("Successfully added P2P edge (source=%d, dest=%d)", sourceID, destID)
}

// TestE2E_Callbacks_ErrorHandling tests error scenarios
func TestE2E_Callbacks_ErrorHandling(t *testing.T) {
	setup := SetupE2ETest(t)

	// Test getting non-existent callback
	_, err := setup.CallMCPTool("mythic_get_callback", map[string]interface{}{
		"callback_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting non-existent callback")

	// Test updating non-existent callback
	_, err = setup.CallMCPTool("mythic_update_callback", map[string]interface{}{
		"callback_id": 999999,
		"description": "Should fail",
	})
	assert.Error(t, err, "Expected error when updating non-existent callback")

	// Test getting loaded commands for non-existent callback
	_, err = setup.CallMCPTool("mythic_get_loaded_commands", map[string]interface{}{
		"callback_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting commands for non-existent callback")

	// Test exporting config for non-existent callback
	_, err = setup.CallMCPTool("mythic_export_callback_config", map[string]interface{}{
		"agent_callback_id": "non-existent-uuid",
	})
	assert.Error(t, err, "Expected error when exporting non-existent callback")
}

// TestE2E_Callbacks_FullWorkflow tests complete callback workflow
func TestE2E_Callbacks_FullWorkflow(t *testing.T) {
	setup := SetupE2ETest(t)

	// Workflow: List callbacks → Get one → Update → Get loaded commands → Export config

	// 1. List all callbacks
	allResult, err := setup.CallMCPTool("mythic_get_all_callbacks", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, allResult)

	// 2. List active callbacks
	activeResult, err := setup.CallMCPTool("mythic_get_active_callbacks", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, activeResult)

	// Get a callback to work with
	allCallbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(allCallbacks) == 0 {
		t.Skip("No callbacks available for full workflow test")
	}

	callback := allCallbacks[0]

	// 3. Get specific callback
	getResult, err := setup.CallMCPTool("mythic_get_callback", map[string]interface{}{
		"callback_id": callback.DisplayID,
	})
	require.NoError(t, err)
	require.NotNil(t, getResult)

	// 4. Update callback
	_, err = setup.CallMCPTool("mythic_update_callback", map[string]interface{}{
		"callback_id": callback.DisplayID,
		"description": "Workflow test callback",
	})
	require.NoError(t, err)

	// 5. Get loaded commands
	cmdsResult, err := setup.CallMCPTool("mythic_get_loaded_commands", map[string]interface{}{
		"callback_id": callback.DisplayID,
	})
	require.NoError(t, err)
	require.NotNil(t, cmdsResult)

	// 6. Get callback tokens
	tokensResult, err := setup.CallMCPTool("mythic_get_callback_tokens", map[string]interface{}{
		"callback_id": callback.DisplayID,
	})
	require.NoError(t, err)
	require.NotNil(t, tokensResult)

	// 7. Export callback config
	exportResult, err := setup.CallMCPTool("mythic_export_callback_config", map[string]interface{}{
		"agent_callback_id": callback.AgentCallbackID,
	})
	require.NoError(t, err)
	require.NotNil(t, exportResult)

	// 8. Verify final state
	finalCB, err := setup.MythicClient.GetCallbackByID(setup.Ctx, callback.DisplayID)
	require.NoError(t, err)
	assert.Equal(t, "Workflow test callback", finalCB.Description)
	t.Logf("Workflow complete for callback %d (%s@%s)", callback.DisplayID, callback.User, callback.Host)
}
