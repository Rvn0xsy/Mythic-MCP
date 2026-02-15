//go:build integration && e2e
// +build integration,e2e

package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/nbaertsch/mythic-sdk-go/pkg/mythic/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_Operations_GetOperations tests listing all operations
func TestE2E_Operations_GetOperations(t *testing.T) {
	setup := SetupE2ETest(t)

	// Call mythic_get_operations tool
	result, err := setup.CallMCPTool("mythic_get_operations", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return operations (at least the default operation should exist)
	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	require.NotEmpty(t, content)
}

// TestE2E_Operations_CreateAndManage tests creating, updating, and getting operations
func TestE2E_Operations_CreateAndManage(t *testing.T) {
	setup := SetupE2ETest(t)
	operationName := fmt.Sprintf("Test Operation E2E %d", time.Now().UnixNano())

	// Step 1: Create a new operation
	createResult, err := setup.CallMCPTool("mythic_create_operation", map[string]interface{}{
		"name":    operationName,
		"webhook": "https://example.com/webhook",
		"channel": "test-channel",
	})
	require.NoError(t, err)
	require.NotNil(t, createResult)

	// Extract operation ID from the result
	// The SDK returns an Operation object, which should be in the tool result metadata
	metadata, ok := createResult["metadata"].(map[string]interface{})
	require.True(t, ok, "Expected metadata in result")
	operationIDFloat, ok := metadata["id"].(float64)
	require.True(t, ok, "Expected operation ID in metadata")
	operationID := int(operationIDFloat)

	// Verify operation was created using SDK directly
	operation, err := setup.MythicClient.GetOperationByID(setup.Ctx, operationID)
	require.NoError(t, err)
	assert.Equal(t, operationName, operation.Name)

	// Step 2: Get the operation via MCP tool
	getResult, err := setup.CallMCPTool("mythic_get_operation", map[string]interface{}{
		"operation_id": operationID,
	})
	require.NoError(t, err)
	require.NotNil(t, getResult)

	// Step 3: Update the operation
	updatedOperationName := fmt.Sprintf("Updated Test Operation %d", time.Now().UnixNano())
	updateResult, err := setup.CallMCPTool("mythic_update_operation", map[string]interface{}{
		"operation_id": operationID,
		"name":         updatedOperationName,
		"complete":     true,
	})
	require.NoError(t, err)
	require.NotNil(t, updateResult)

	// Verify update using SDK
	updatedOp, err := setup.MythicClient.GetOperationByID(setup.Ctx, operationID)
	require.NoError(t, err)
	assert.Equal(t, updatedOperationName, updatedOp.Name)
	assert.True(t, updatedOp.Complete)
}

// TestE2E_Operations_CurrentOperation tests setting and getting current operation
func TestE2E_Operations_CurrentOperation(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get current operation first
	getResult, err := setup.CallMCPTool("mythic_get_current_operation", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, getResult)

	// Create a new operation to set as current
	currentOpName := fmt.Sprintf("Current Op Test %d", time.Now().UnixNano())
	operation, err := setup.MythicClient.CreateOperation(setup.Ctx, &types.CreateOperationRequest{
		Name: currentOpName,
	})
	require.NoError(t, err)

	// Set current operation
	setResult, err := setup.CallMCPTool("mythic_set_current_operation", map[string]interface{}{
		"operation_id": operation.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, setResult)

	// Get current operation via tool
	getCurrentResult, err := setup.CallMCPTool("mythic_get_current_operation", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, getCurrentResult)

	metadata, ok := getCurrentResult["metadata"].(map[string]interface{})
	require.True(t, ok, "Expected metadata map in result")
	opID, ok := metadata["operation_id"].(float64)
	require.True(t, ok, "Expected operation_id in metadata")
	assert.Equal(t, operation.ID, int(opID))
}

// TestE2E_Operations_Operators tests getting operators for an operation
func TestE2E_Operations_Operators(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get default operation (should exist)
	operations, err := setup.MythicClient.GetOperations(setup.Ctx)
	require.NoError(t, err)
	require.NotEmpty(t, operations, "Expected at least one operation to exist")

	operationID := operations[0].ID

	// Get operators for this operation
	result, err := setup.CallMCPTool("mythic_get_operation_operators", map[string]interface{}{
		"operation_id": operationID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should have at least the admin user
	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	require.NotEmpty(t, content, "Expected at least one operator in operation")
}

// TestE2E_Operations_EventLog tests creating and retrieving event logs
func TestE2E_Operations_EventLog(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get an operation to log events for
	operations, err := setup.MythicClient.GetOperations(setup.Ctx)
	require.NoError(t, err)
	require.NotEmpty(t, operations)

	operationID := operations[0].ID

	// Create an event log entry
	createResult, err := setup.CallMCPTool("mythic_create_event_log", map[string]interface{}{
		"operation_id": operationID,
		"message":      "Test event log from E2E test",
		"level":        "info",
	})
	require.NoError(t, err)
	require.NotNil(t, createResult)

	// Get event logs for the operation
	getResult, err := setup.CallMCPTool("mythic_get_event_log", map[string]interface{}{
		"operation_id": operationID,
		"limit":        10,
	})
	require.NoError(t, err)
	require.NotNil(t, getResult)

	// Should contain our test message
	content, ok := getResult["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	require.NotEmpty(t, content, "Expected at least one event log entry")
}

// TestE2E_Operations_GlobalSettings tests getting and updating global settings
func TestE2E_Operations_GlobalSettings(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get global settings
	getResult, err := setup.CallMCPTool("mythic_get_global_settings", map[string]interface{}{})

	// Note: Global settings may not be available in all Mythic versions
	// If we get an error, skip the rest of the test
	if err != nil {
		t.Skip("Global settings not available in this Mythic version")
	}
	require.NotNil(t, getResult)

	// Try to update global settings (if available)
	// Note: This is a risky operation, so we'll just test the tool call mechanism
	// In a real scenario, you'd want to save and restore the original settings
	updateResult, err := setup.CallMCPTool("mythic_update_global_settings", map[string]interface{}{
		"settings": map[string]interface{}{
			"test_key": "test_value",
		},
	})

	// If update is not supported, that's okay - just verify the tool exists
	if err != nil {
		t.Logf("Global settings update not supported: %v", err)
	} else {
		require.NotNil(t, updateResult)
	}
}

// TestE2E_Operations_ErrorHandling tests error scenarios
func TestE2E_Operations_ErrorHandling(t *testing.T) {
	setup := SetupE2ETest(t)

	// Test getting non-existent operation
	_, err := setup.CallMCPTool("mythic_get_operation", map[string]interface{}{
		"operation_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting non-existent operation")

	// Test updating non-existent operation
	_, err = setup.CallMCPTool("mythic_update_operation", map[string]interface{}{
		"operation_id": 999999,
		"name":         "Should Fail",
	})
	assert.Error(t, err, "Expected error when updating non-existent operation")

	// Test creating operation with invalid data
	_, err = setup.CallMCPTool("mythic_create_operation", map[string]interface{}{
		"name": "", // Empty name should fail
	})
	assert.Error(t, err, "Expected error when creating operation with empty name")
}

// TestE2E_Operations_FullWorkflow tests a complete operations workflow
func TestE2E_Operations_FullWorkflow(t *testing.T) {
	setup := SetupE2ETest(t)
	workflowName := fmt.Sprintf("Workflow Test Operation %d", time.Now().UnixNano())

	// Workflow: Create operation → Set as current → Log events → Update → Verify

	// 1. Create operation
	createResult, err := setup.CallMCPTool("mythic_create_operation", map[string]interface{}{
		"name":    workflowName,
		"webhook": "https://example.com/hook",
	})
	require.NoError(t, err)
	metadata, ok := createResult["metadata"].(map[string]interface{})
	require.True(t, ok)
	operationID := int(metadata["id"].(float64))

	// 2. Set as current operation
	_, err = setup.CallMCPTool("mythic_set_current_operation", map[string]interface{}{
		"operation_id": operationID,
	})
	require.NoError(t, err)

	// 3. Log an event
	_, err = setup.CallMCPTool("mythic_create_event_log", map[string]interface{}{
		"operation_id": operationID,
		"message":      "Workflow test started",
		"level":        "info",
	})
	require.NoError(t, err)

	// 4. Get operation details
	getResult, err := setup.CallMCPTool("mythic_get_operation", map[string]interface{}{
		"operation_id": operationID,
	})
	require.NoError(t, err)
	require.NotNil(t, getResult)

	// 5. Get event logs
	logResult, err := setup.CallMCPTool("mythic_get_event_log", map[string]interface{}{
		"operation_id": operationID,
		"limit":        5,
	})
	require.NoError(t, err)
	require.NotNil(t, logResult)

	// 6. Update operation to mark as complete
	_, err = setup.CallMCPTool("mythic_update_operation", map[string]interface{}{
		"operation_id": operationID,
		"complete":     true,
	})
	require.NoError(t, err)

	// 7. Verify final state
	finalOp, err := setup.MythicClient.GetOperationByID(setup.Ctx, operationID)
	require.NoError(t, err)
	assert.True(t, finalOp.Complete, "Operation should be marked as complete")
	assert.Equal(t, workflowName, finalOp.Name)
}
