//go:build integration && e2e
// +build integration,e2e

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_Processes_GetProcesses tests listing all processes
func TestE2E_Processes_GetProcesses(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all processes
	result, err := setup.CallMCPTool("mythic_get_processes", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Result should be an array (may be empty if no processes)
	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Total processes: %d", len(content))
}

// TestE2E_Processes_GetProcessesByOperation tests listing processes by operation
func TestE2E_Processes_GetProcessesByOperation(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get current operation
	me, err := setup.MythicClient.GetMe(setup.Ctx)
	require.NoError(t, err)

	if me.CurrentOperationID == nil {
		t.Skip("No current operation set")
	}

	operationID := *me.CurrentOperationID

	// Get processes for operation
	result, err := setup.CallMCPTool("mythic_get_processes_by_operation", map[string]interface{}{
		"operation_id": operationID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Processes in operation %d: %d", operationID, len(content))
}

// TestE2E_Processes_GetProcessesByCallback tests listing processes by callback
func TestE2E_Processes_GetProcessesByCallback(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a callback
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available to test")
	}

	callbackID := callbacks[0].DisplayID

	// Get processes for callback
	result, err := setup.CallMCPTool("mythic_get_processes_by_callback", map[string]interface{}{
		"callback_id": callbackID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Processes for callback %d: %d", callbackID, len(content))
}

// TestE2E_Processes_GetProcessTree tests getting process tree
func TestE2E_Processes_GetProcessTree(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a callback
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available to test")
	}

	callbackID := callbacks[0].DisplayID

	// Get process tree for callback
	result, err := setup.CallMCPTool("mythic_get_process_tree", map[string]interface{}{
		"callback_id": callbackID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Process tree nodes for callback %d: %d", callbackID, len(content))
}

// TestE2E_Processes_GetProcessesByHost tests listing processes by host
func TestE2E_Processes_GetProcessesByHost(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get current operation
	me, err := setup.MythicClient.GetMe(setup.Ctx)
	require.NoError(t, err)

	if me.CurrentOperationID == nil {
		t.Skip("No current operation set")
	}

	operationID := *me.CurrentOperationID

	// Get hosts
	hosts, err := setup.MythicClient.GetHosts(setup.Ctx, operationID)
	require.NoError(t, err)

	if len(hosts) == 0 {
		t.Skip("No hosts available to test")
	}

	hostID := hosts[0].ID

	// Get processes for host
	result, err := setup.CallMCPTool("mythic_get_processes_by_host", map[string]interface{}{
		"host_id": hostID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Processes for host %d: %d", hostID, len(content))
}

// TestE2E_Processes_ErrorHandling tests error scenarios
func TestE2E_Processes_ErrorHandling(t *testing.T) {
	setup := SetupE2ETest(t)

	// Test getting processes for non-existent operation
	_, err := setup.CallMCPTool("mythic_get_processes_by_operation", map[string]interface{}{
		"operation_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting processes for non-existent operation")

	// Test getting processes for non-existent callback
	_, err = setup.CallMCPTool("mythic_get_processes_by_callback", map[string]interface{}{
		"callback_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting processes for non-existent callback")

	// Test getting process tree for non-existent callback
	_, err = setup.CallMCPTool("mythic_get_process_tree", map[string]interface{}{
		"callback_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting tree for non-existent callback")

	// Test getting processes for non-existent host
	_, err = setup.CallMCPTool("mythic_get_processes_by_host", map[string]interface{}{
		"host_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting processes for non-existent host")
}

// TestE2E_Processes_FullWorkflow tests complete process workflow
func TestE2E_Processes_FullWorkflow(t *testing.T) {
	setup := SetupE2ETest(t)

	// Workflow: Get all processes → Get by operation → Get by callback → Get tree

	// 1. Get all processes
	allProcessesResult, err := setup.CallMCPTool("mythic_get_processes", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, allProcessesResult)

	// 2. Get current operation
	me, err := setup.MythicClient.GetMe(setup.Ctx)
	require.NoError(t, err)

	if me.CurrentOperationID == nil {
		t.Skip("No current operation set for full workflow test")
	}

	operationID := *me.CurrentOperationID

	// 3. Get processes by operation
	operationProcessesResult, err := setup.CallMCPTool("mythic_get_processes_by_operation", map[string]interface{}{
		"operation_id": operationID,
	})
	require.NoError(t, err)
	require.NotNil(t, operationProcessesResult)

	// Get a callback to work with
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available for full workflow test")
	}

	callback := callbacks[0]

	// 4. Get processes by callback
	callbackProcessesResult, err := setup.CallMCPTool("mythic_get_processes_by_callback", map[string]interface{}{
		"callback_id": callback.DisplayID,
	})
	require.NoError(t, err)
	require.NotNil(t, callbackProcessesResult)

	// 5. Get process tree
	processTreeResult, err := setup.CallMCPTool("mythic_get_process_tree", map[string]interface{}{
		"callback_id": callback.DisplayID,
	})
	require.NoError(t, err)
	require.NotNil(t, processTreeResult)

	t.Logf("Workflow complete for callback %d (%s@%s)", callback.DisplayID, callback.User, callback.Host)
}

// TestE2E_Processes_ProcessDetails tests detailed process information
func TestE2E_Processes_ProcessDetails(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all processes
	processes, err := setup.MythicClient.GetProcesses(setup.Ctx)
	require.NoError(t, err)

	if len(processes) == 0 {
		t.Skip("No processes available to test")
	}

	// Log details for first few processes
	for i, process := range processes {
		if i >= 3 {
			break
		}

		t.Logf("Process %d:", process.ProcessID)
		t.Logf("  - Name: %s", process.Name)
		t.Logf("  - PID: %d", process.ProcessID)
		if process.ParentProcessID != 0 {
			t.Logf("  - Parent PID: %d", process.ParentProcessID)
		}
		if process.Host != nil {
			t.Logf("  - Host: %s", process.Host.Host)
		}
		t.Logf("  - User: %s", process.User)
	}
}
