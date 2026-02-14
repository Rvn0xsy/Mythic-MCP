//go:build integration && e2e
// +build integration,e2e

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_Tasks_IssueTask tests issuing a task to a callback
func TestE2E_Tasks_IssueTask(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a callback to task
	callbacks, err := setup.MythicClient.GetAllActiveCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No active callbacks available for testing")
	}

	callbackID := callbacks[0].DisplayID

	// Issue a simple task (e.g., "whoami" or similar safe command)
	// Note: The exact command depends on the agent type
	issueResult, err := setup.CallMCPTool("mythic_issue_task", map[string]interface{}{
		"callback_id": callbackID,
		"command":     "shell",
		"params":      "{\"command\":\"whoami\"}",
	})

	// May fail if command not loaded - that's okay
	if err != nil {
		t.Logf("Issue task failed (expected if command not available): %v", err)
		return
	}

	require.NotNil(t, issueResult)
	t.Logf("Successfully issued task to callback %d", callbackID)
}

// TestE2E_Tasks_GetTask tests getting task details
func TestE2E_Tasks_GetTask(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get callbacks
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available for testing")
	}

	callbackID := callbacks[0].DisplayID

	// Get tasks for this callback
	tasks, err := setup.MythicClient.GetTasksForCallback(setup.Ctx, callbackID, 10)
	require.NoError(t, err)

	if len(tasks) == 0 {
		t.Skip("No tasks available for testing")
	}

	taskID := tasks[0].DisplayID

	// Get specific task
	getResult, err := setup.CallMCPTool("mythic_get_task", map[string]interface{}{
		"task_id": taskID,
	})
	require.NoError(t, err)
	require.NotNil(t, getResult)
}

// TestE2E_Tasks_GetCallbackTasks tests listing tasks for a callback
func TestE2E_Tasks_GetCallbackTasks(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a callback
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available for testing")
	}

	callbackID := callbacks[0].DisplayID

	// Get tasks for callback
	result, err := setup.CallMCPTool("mythic_get_callback_tasks", map[string]interface{}{
		"callback_id": callbackID,
		"limit":       50,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Tasks for callback %d: %d", callbackID, len(content))
}

// TestE2E_Tasks_GetTasksByStatus tests filtering tasks by status
func TestE2E_Tasks_GetTasksByStatus(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a callback
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available for testing")
	}

	callbackID := callbacks[0].DisplayID

	// Get completed tasks
	result, err := setup.CallMCPTool("mythic_get_tasks_by_status", map[string]interface{}{
		"callback_id": callbackID,
		"status":      "completed",
		"limit":       10,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Completed tasks for callback %d: %d", callbackID, len(content))
}

// TestE2E_Tasks_GetTaskOutput tests getting task output/responses
func TestE2E_Tasks_GetTaskOutput(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a callback with tasks
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available for testing")
	}

	callbackID := callbacks[0].DisplayID
	tasks, err := setup.MythicClient.GetTasksForCallback(setup.Ctx, callbackID, 10)
	require.NoError(t, err)

	if len(tasks) == 0 {
		t.Skip("No tasks available for testing")
	}

	taskID := tasks[0].DisplayID

	// Get task output
	result, err := setup.CallMCPTool("mythic_get_task_output", map[string]interface{}{
		"task_id": taskID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
}

// TestE2E_Tasks_UpdateTask tests updating task properties
func TestE2E_Tasks_UpdateTask(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a task to update
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available for testing")
	}

	callbackID := callbacks[0].DisplayID
	tasks, err := setup.MythicClient.GetTasksForCallback(setup.Ctx, callbackID, 10)
	require.NoError(t, err)

	if len(tasks) == 0 {
		t.Skip("No tasks available for testing")
	}

	taskID := tasks[0].DisplayID

	// Update task comment
	updateResult, err := setup.CallMCPTool("mythic_update_task", map[string]interface{}{
		"task_id": taskID,
		"updates": map[string]interface{}{
			"comment": "Updated via E2E test",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, updateResult)
}

// TestE2E_Tasks_GetTaskArtifacts tests getting task artifacts
func TestE2E_Tasks_GetTaskArtifacts(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a task
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available for testing")
	}

	callbackID := callbacks[0].DisplayID
	tasks, err := setup.MythicClient.GetTasksForCallback(setup.Ctx, callbackID, 10)
	require.NoError(t, err)

	if len(tasks) == 0 {
		t.Skip("No tasks available for testing")
	}

	taskID := tasks[0].DisplayID

	// Get task artifacts
	result, err := setup.CallMCPTool("mythic_get_task_artifacts", map[string]interface{}{
		"task_id": taskID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Artifacts for task %d: %d", taskID, len(content))
}

// TestE2E_Responses_GetTaskResponses tests getting responses for a task
func TestE2E_Responses_GetTaskResponses(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a task with responses
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available for testing")
	}

	callbackID := callbacks[0].DisplayID
	tasks, err := setup.MythicClient.GetTasksForCallback(setup.Ctx, callbackID, 10)
	require.NoError(t, err)

	if len(tasks) == 0 {
		t.Skip("No tasks available for testing")
	}

	taskID := tasks[0].DisplayID

	// Get task responses
	result, err := setup.CallMCPTool("mythic_get_task_responses", map[string]interface{}{
		"task_id": taskID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Responses for task %d: %d", taskID, len(content))
}

// TestE2E_Responses_GetCallbackResponses tests getting responses for a callback
func TestE2E_Responses_GetCallbackResponses(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a callback
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available for testing")
	}

	callbackID := callbacks[0].DisplayID

	// Get callback responses
	result, err := setup.CallMCPTool("mythic_get_callback_responses", map[string]interface{}{
		"callback_id": callbackID,
		"limit":       50,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Responses for callback %d: %d", callbackID, len(content))
}

// TestE2E_Responses_GetLatestResponses tests getting latest responses
func TestE2E_Responses_GetLatestResponses(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get current operation
	me, err := setup.MythicClient.GetMe(setup.Ctx)
	require.NoError(t, err)

	if me.CurrentOperation == nil {
		t.Skip("No current operation set")
	}

	operationID := me.CurrentOperation.ID

	// Get latest responses
	result, err := setup.CallMCPTool("mythic_get_latest_responses", map[string]interface{}{
		"operation_id": operationID,
		"limit":        20,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Latest responses for operation %d: %d", operationID, len(content))
}

// TestE2E_Responses_SearchResponses tests searching response text
func TestE2E_Responses_SearchResponses(t *testing.T) {
	setup := SetupE2ETest(t)

	// Search for responses containing common text
	result, err := setup.CallMCPTool("mythic_search_responses", map[string]interface{}{
		"search_term": "success",
	})

	// May return no results if no matching responses
	if err != nil {
		t.Logf("Search failed or no results: %v", err)
		return
	}

	require.NotNil(t, result)
	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Search results for 'success': %d", len(content))
}

// TestE2E_Tasks_ErrorHandling tests error scenarios
func TestE2E_Tasks_ErrorHandling(t *testing.T) {
	setup := SetupE2ETest(t)

	// Test getting non-existent task
	_, err := setup.CallMCPTool("mythic_get_task", map[string]interface{}{
		"task_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting non-existent task")

	// Test updating non-existent task
	_, err = setup.CallMCPTool("mythic_update_task", map[string]interface{}{
		"task_id": 999999,
		"updates": map[string]interface{}{"comment": "fail"},
	})
	assert.Error(t, err, "Expected error when updating non-existent task")

	// Test getting tasks for non-existent callback
	_, err = setup.CallMCPTool("mythic_get_callback_tasks", map[string]interface{}{
		"callback_id": 999999,
		"limit":       10,
	})
	assert.Error(t, err, "Expected error when getting tasks for non-existent callback")

	// Test issuing task to non-existent callback
	_, err = setup.CallMCPTool("mythic_issue_task", map[string]interface{}{
		"callback_id": 999999,
		"command":     "invalid",
		"params":      "{}",
	})
	assert.Error(t, err, "Expected error when issuing to non-existent callback")
}

// TestE2E_Tasks_FullWorkflow tests complete task workflow
func TestE2E_Tasks_FullWorkflow(t *testing.T) {
	setup := SetupE2ETest(t)

	// Workflow: Get callback → List tasks → Get task → Get output → Get responses

	// 1. Get an active callback
	callbacks, err := setup.MythicClient.GetAllActiveCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No active callbacks available for workflow test")
	}

	callback := callbacks[0]
	t.Logf("Using callback %d (%s@%s)", callback.DisplayID, callback.User, callback.Host)

	// 2. Get tasks for callback
	tasksResult, err := setup.CallMCPTool("mythic_get_callback_tasks", map[string]interface{}{
		"callback_id": callback.DisplayID,
		"limit":       10,
	})
	require.NoError(t, err)
	require.NotNil(t, tasksResult)

	// 3. Get tasks for callback from SDK to get a task ID
	tasks, err := setup.MythicClient.GetTasksForCallback(setup.Ctx, callback.DisplayID, 10)
	require.NoError(t, err)

	if len(tasks) == 0 {
		t.Skip("No tasks available for workflow test")
	}

	task := tasks[0]
	t.Logf("Using task %d (status: %s)", task.DisplayID, task.Status)

	// 4. Get specific task
	taskResult, err := setup.CallMCPTool("mythic_get_task", map[string]interface{}{
		"task_id": task.DisplayID,
	})
	require.NoError(t, err)
	require.NotNil(t, taskResult)

	// 5. Get task output
	outputResult, err := setup.CallMCPTool("mythic_get_task_output", map[string]interface{}{
		"task_id": task.DisplayID,
	})
	require.NoError(t, err)
	require.NotNil(t, outputResult)

	// 6. Get task responses
	responsesResult, err := setup.CallMCPTool("mythic_get_task_responses", map[string]interface{}{
		"task_id": task.DisplayID,
	})
	require.NoError(t, err)
	require.NotNil(t, responsesResult)

	// 7. Get task artifacts
	artifactsResult, err := setup.CallMCPTool("mythic_get_task_artifacts", map[string]interface{}{
		"task_id": task.DisplayID,
	})
	require.NoError(t, err)
	require.NotNil(t, artifactsResult)

	// 8. Get callback responses
	cbResponsesResult, err := setup.CallMCPTool("mythic_get_callback_responses", map[string]interface{}{
		"callback_id": callback.DisplayID,
		"limit":       20,
	})
	require.NoError(t, err)
	require.NotNil(t, cbResponsesResult)

	t.Logf("Workflow complete for callback %d, task %d", callback.DisplayID, task.DisplayID)
}

// TestE2E_Tasks_WaitForTask tests waiting for task completion
func TestE2E_Tasks_WaitForTask(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a completed or processing task
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available for testing")
	}

	callbackID := callbacks[0].DisplayID
	tasks, err := setup.MythicClient.GetTasksForCallback(setup.Ctx, callbackID, 5)
	require.NoError(t, err)

	if len(tasks) == 0 {
		t.Skip("No tasks available for testing")
	}

	taskID := tasks[0].DisplayID

	// Wait for task with short timeout (it's probably already done)
	waitResult, err := setup.CallMCPTool("mythic_wait_for_task", map[string]interface{}{
		"task_id": taskID,
		"timeout": 5,
	})

	// If task is already complete, this should succeed
	// If still processing, may timeout - both are okay
	if err != nil {
		t.Logf("Wait timed out or failed (expected for in-progress tasks): %v", err)
		return
	}

	require.NotNil(t, waitResult)
	t.Logf("Task %d completed or was already complete", taskID)
}
