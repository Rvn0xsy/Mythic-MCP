package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_Attack_GetAttackTechniques tests listing all MITRE ATT&CK techniques
func TestE2E_Attack_GetAttackTechniques(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all attack techniques
	result, err := setup.CallMCPTool("mythic_get_attack_techniques", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Result should be an array
	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Total MITRE ATT&CK techniques: %d", len(content))
}

// TestE2E_Attack_GetAttackTechniqueByID tests getting specific technique by ID
func TestE2E_Attack_GetAttackTechniqueByID(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all techniques first
	techniques, err := setup.MythicClient.GetAttackTechniques(setup.Ctx)
	require.NoError(t, err)

	if len(techniques) == 0 {
		t.Skip("No MITRE ATT&CK techniques available to test")
	}

	attackID := techniques[0].ID

	// Get specific technique by ID
	result, err := setup.CallMCPTool("mythic_get_attack_technique_by_id", map[string]interface{}{
		"attack_id": attackID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	t.Logf("Retrieved technique ID %d", attackID)
}

// TestE2E_Attack_GetAttackTechniqueByTNum tests getting technique by T-number
func TestE2E_Attack_GetAttackTechniqueByTNum(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all techniques first to find a valid T-number
	techniques, err := setup.MythicClient.GetAttackTechniques(setup.Ctx)
	require.NoError(t, err)

	if len(techniques) == 0 {
		t.Skip("No MITRE ATT&CK techniques available to test")
	}

	tNum := techniques[0].TNum

	// Get technique by T-number
	result, err := setup.CallMCPTool("mythic_get_attack_technique_by_tnum", map[string]interface{}{
		"t_number": tNum,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	t.Logf("Retrieved technique %s", tNum)
}

// TestE2E_Attack_GetAttackByTask tests getting techniques used by a task
func TestE2E_Attack_GetAttackByTask(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a task with MITRE mappings
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available to test")
	}

	callbackID := callbacks[0].DisplayID
	tasks, err := setup.MythicClient.GetTasksForCallback(setup.Ctx, callbackID, 10)
	require.NoError(t, err)

	if len(tasks) == 0 {
		t.Skip("No tasks available to test")
	}

	// Get task first to get internal ID
	task, err := setup.MythicClient.GetTask(setup.Ctx, tasks[0].DisplayID)
	require.NoError(t, err)

	// Get attack techniques for task
	result, err := setup.CallMCPTool("mythic_get_attack_by_task", map[string]interface{}{
		"task_id": task.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("MITRE techniques for task %d: %d", task.DisplayID, len(content))
}

// TestE2E_Attack_GetAttackByCommand tests getting techniques for a command
func TestE2E_Attack_GetAttackByCommand(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get commands first
	commands, err := setup.MythicClient.GetCommands(setup.Ctx)
	require.NoError(t, err)

	if len(commands) == 0 {
		t.Skip("No commands available to test")
	}

	commandID := commands[0].ID

	// Get attack techniques for command
	result, err := setup.CallMCPTool("mythic_get_attack_by_command", map[string]interface{}{
		"command_id": commandID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("MITRE techniques for command %d (%s): %d", commandID, commands[0].Cmd, len(content))
}

// TestE2E_Attack_GetAttacksByOperation tests getting all techniques used in operation
func TestE2E_Attack_GetAttacksByOperation(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get current operation
	me, err := setup.MythicClient.GetMe(setup.Ctx)
	require.NoError(t, err)

	if me.CurrentOperationID == nil {
		t.Skip("No current operation set")
	}

	operationID := *me.CurrentOperationID

	// Get attack techniques for operation
	result, err := setup.CallMCPTool("mythic_get_attacks_by_operation", map[string]interface{}{
		"operation_id": operationID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("MITRE techniques used in operation %d: %d", operationID, len(content))
}

// TestE2E_Attack_ErrorHandling tests error scenarios
func TestE2E_Attack_ErrorHandling(t *testing.T) {
	setup := SetupE2ETest(t)

	// Test getting non-existent technique by ID
	_, err := setup.CallMCPTool("mythic_get_attack_technique_by_id", map[string]interface{}{
		"attack_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting non-existent technique")

	// Test getting technique with invalid T-number
	_, err = setup.CallMCPTool("mythic_get_attack_technique_by_tnum", map[string]interface{}{
		"t_number": "T99999",
	})
	assert.Error(t, err, "Expected error when getting invalid T-number")

	// Test getting attacks for non-existent task
	_, err = setup.CallMCPTool("mythic_get_attack_by_task", map[string]interface{}{
		"task_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting attacks for non-existent task")

	// Test getting attacks for non-existent command
	_, err = setup.CallMCPTool("mythic_get_attack_by_command", map[string]interface{}{
		"command_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting attacks for non-existent command")
}

// TestE2E_Attack_FullWorkflow tests complete MITRE ATT&CK workflow
func TestE2E_Attack_FullWorkflow(t *testing.T) {
	setup := SetupE2ETest(t)

	// Workflow: Get all techniques → Get specific technique → Get usage in operation

	// 1. Get all MITRE techniques
	allTechniquesResult, err := setup.CallMCPTool("mythic_get_attack_techniques", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, allTechniquesResult)

	// Get techniques to work with
	techniques, err := setup.MythicClient.GetAttackTechniques(setup.Ctx)
	require.NoError(t, err)

	if len(techniques) == 0 {
		t.Skip("No MITRE techniques available for full workflow test")
	}

	technique := techniques[0]

	// 2. Get specific technique by ID
	techniqueByIDResult, err := setup.CallMCPTool("mythic_get_attack_technique_by_id", map[string]interface{}{
		"attack_id": technique.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, techniqueByIDResult)

	// 3. Get specific technique by T-number
	techniqueByTNumResult, err := setup.CallMCPTool("mythic_get_attack_technique_by_tnum", map[string]interface{}{
		"t_number": technique.TNum,
	})
	require.NoError(t, err)
	require.NotNil(t, techniqueByTNumResult)

	// 4. Get current operation
	me, err := setup.MythicClient.GetMe(setup.Ctx)
	require.NoError(t, err)

	if me.CurrentOperationID != nil {
		// 5. Get all techniques used in operation
		operationTechniquesResult, err := setup.CallMCPTool("mythic_get_attacks_by_operation", map[string]interface{}{
			"operation_id": *me.CurrentOperationID,
		})
		require.NoError(t, err)
		require.NotNil(t, operationTechniquesResult)
	}

	t.Logf("Workflow complete for MITRE technique %s (%s)", technique.TNum, technique.Name)
}

// TestE2E_Attack_TechniqueDetails tests detailed technique information
func TestE2E_Attack_TechniqueDetails(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all techniques
	techniques, err := setup.MythicClient.GetAttackTechniques(setup.Ctx)
	require.NoError(t, err)

	if len(techniques) == 0 {
		t.Skip("No MITRE techniques available to test")
	}

	// Test getting details for first few techniques
	for i, technique := range techniques {
		if i >= 3 {
			break
		}

		t.Logf("Technique %s:", technique.TNum)
		t.Logf("  - Name: %s", technique.Name)
		t.Logf("  - Tactic: %s", technique.Tactic)
		t.Logf("  - OS: %s", technique.OS)
		t.Logf("  - ID: %d", technique.ID)
	}
}
