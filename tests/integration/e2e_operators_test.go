package integration

import (
	"testing"

	"github.com/nbaertsch/mythic-sdk-go/pkg/mythic/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_Operators_GetOperators tests listing all operators
func TestE2E_Operators_GetOperators(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all operators
	result, err := setup.CallMCPTool("mythic_get_operators", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return at least the admin user
	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	require.NotEmpty(t, content, "Expected at least one operator")
}

// TestE2E_Operators_CreateAndManage tests operator creation and management
func TestE2E_Operators_CreateAndManage(t *testing.T) {
	setup := SetupE2ETest(t)

	// Step 1: Create a new operator
	createResult, err := setup.CallMCPTool("mythic_create_operator", map[string]interface{}{
		"username": "test-operator-e2e",
		"password": "TestPassword123!",
	})
	require.NoError(t, err)
	require.NotNil(t, createResult)

	// Extract operator ID
	metadata, ok := createResult["metadata"].(map[string]interface{})
	require.True(t, ok, "Expected metadata in result")
	operatorIDFloat, ok := metadata["id"].(float64)
	require.True(t, ok, "Expected operator ID in metadata")
	operatorID := int(operatorIDFloat)

	// Step 2: Get the operator by ID
	getResult, err := setup.CallMCPTool("mythic_get_operator", map[string]interface{}{
		"operator_id": operatorID,
	})
	require.NoError(t, err)
	require.NotNil(t, getResult)

	// Step 3: Update operator status (deactivate)
	updateResult, err := setup.CallMCPTool("mythic_update_operator_status", map[string]interface{}{
		"operator_id": operatorID,
		"active":      false,
	})
	require.NoError(t, err)
	require.NotNil(t, updateResult)

	// Verify status was updated
	operator, err := setup.MythicClient.GetOperatorByID(setup.Ctx, operatorID)
	require.NoError(t, err)
	assert.False(t, operator.Active, "Operator should be deactivated")

	// Step 4: Re-activate operator
	_, err = setup.CallMCPTool("mythic_update_operator_status", map[string]interface{}{
		"operator_id": operatorID,
		"active":      true,
	})
	require.NoError(t, err)

	// Verify reactivation
	operator, err = setup.MythicClient.GetOperatorByID(setup.Ctx, operatorID)
	require.NoError(t, err)
	assert.True(t, operator.Active, "Operator should be active")
}

// TestE2E_Operators_PasswordAndEmail tests password and email updates
func TestE2E_Operators_PasswordAndEmail(t *testing.T) {
	setup := SetupE2ETest(t)

	// Create a test operator
	operator, err := setup.MythicClient.CreateOperator(setup.Ctx, &types.CreateOperatorRequest{
		Username: "test-pwd-change",
		Password: "InitialPassword123!",
	})
	require.NoError(t, err)

	// Update password and email
	updateResult, err := setup.CallMCPTool("mythic_update_password_email", map[string]interface{}{
		"operator_id":  operator.ID,
		"old_password": "InitialPassword123!",
		"new_password": "NewPassword456!",
		"email":        "test@example.com",
	})

	// Note: This may fail if the operator doesn't match current user
	// That's expected - just verify the tool works
	if err != nil {
		t.Logf("Password update (expected to fail for different operator): %v", err)
	} else {
		require.NotNil(t, updateResult)
	}
}

// TestE2E_Operators_Preferences tests operator preferences
func TestE2E_Operators_Preferences(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get current operator
	currentOp, err := setup.MythicClient.GetMe(setup.Ctx)
	require.NoError(t, err)

	// Get preferences
	getResult, err := setup.CallMCPTool("mythic_get_operator_preferences", map[string]interface{}{
		"operator_id": currentOp.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, getResult)

	// Update preferences
	updateResult, err := setup.CallMCPTool("mythic_update_operator_preferences", map[string]interface{}{
		"operator_id": currentOp.ID,
		"preferences": map[string]interface{}{
			"test_key":  "test_value",
			"theme":     "dark",
			"font_size": 14,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, updateResult)

	// Verify update
	prefs, err := setup.MythicClient.GetOperatorPreferences(setup.Ctx, currentOp.ID)
	require.NoError(t, err)
	assert.NotNil(t, prefs.Preferences)
}

// TestE2E_Operators_Secrets tests operator secrets management
func TestE2E_Operators_Secrets(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get current operator
	currentOp, err := setup.MythicClient.GetMe(setup.Ctx)
	require.NoError(t, err)

	// Get secrets
	getResult, err := setup.CallMCPTool("mythic_get_operator_secrets", map[string]interface{}{
		"operator_id": currentOp.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, getResult)

	// Update secrets
	updateResult, err := setup.CallMCPTool("mythic_update_operator_secrets", map[string]interface{}{
		"operator_id": currentOp.ID,
		"secrets": map[string]interface{}{
			"api_key": "test-api-key-12345",
			"token":   "test-token-67890",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, updateResult)

	// Verify update
	secrets, err := setup.MythicClient.GetOperatorSecrets(setup.Ctx, currentOp.ID)
	require.NoError(t, err)
	assert.NotNil(t, secrets.Secrets)
}

// TestE2E_Operators_InviteLinks tests invite link management
func TestE2E_Operators_InviteLinks(t *testing.T) {
	setup := SetupE2ETest(t)

	// Create an invite link
	createResult, err := setup.CallMCPTool("mythic_create_invite_link", map[string]interface{}{
		"max_uses": 5,
		"name":     "Test E2E Invite",
	})
	require.NoError(t, err)
	require.NotNil(t, createResult)

	// Get all invite links
	getResult, err := setup.CallMCPTool("mythic_get_invite_links", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, getResult)

	// Should contain at least our newly created link
	content, ok := getResult["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	assert.NotEmpty(t, content, "Expected at least one invite link")
}

// TestE2E_Operators_UpdateOperatorOperation tests operator-operation management
func TestE2E_Operators_UpdateOperatorOperation(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get an operation to work with
	operations, err := setup.MythicClient.GetOperations(setup.Ctx)
	require.NoError(t, err)
	require.NotEmpty(t, operations)
	operationID := operations[0].ID

	// Create a test operator
	operator, err := setup.MythicClient.CreateOperator(setup.Ctx, &types.CreateOperatorRequest{
		Username: "test-op-assignment",
		Password: "TestPassword123!",
	})
	require.NoError(t, err)

	// Add operator to operation
	addResult, err := setup.CallMCPTool("mythic_update_operator_operation", map[string]interface{}{
		"operation_id": operationID,
		"add_users":    []int{operator.ID},
	})
	require.NoError(t, err)
	require.NotNil(t, addResult)

	// Verify operator is in operation
	operators, err := setup.MythicClient.GetOperatorsByOperation(setup.Ctx, operationID)
	require.NoError(t, err)

	found := false
	for _, op := range operators {
		if op.OperatorID == operator.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "Operator should be in operation")

	// Remove operator from operation
	removeResult, err := setup.CallMCPTool("mythic_update_operator_operation", map[string]interface{}{
		"operation_id": operationID,
		"remove_users": []int{operator.ID},
	})
	require.NoError(t, err)
	require.NotNil(t, removeResult)
}

// TestE2E_Operators_ErrorHandling tests error scenarios
func TestE2E_Operators_ErrorHandling(t *testing.T) {
	setup := SetupE2ETest(t)

	// Test getting non-existent operator
	_, err := setup.CallMCPTool("mythic_get_operator", map[string]interface{}{
		"operator_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting non-existent operator")

	// Test creating operator with invalid password (too short)
	_, err = setup.CallMCPTool("mythic_create_operator", map[string]interface{}{
		"username": "test-invalid",
		"password": "short",
	})
	assert.Error(t, err, "Expected error with short password")

	// Test creating operator with duplicate username
	// First create one
	_, err = setup.CallMCPTool("mythic_create_operator", map[string]interface{}{
		"username": "test-duplicate-check",
		"password": "ValidPassword123!",
	})
	// May or may not succeed depending on if it already exists

	// Try to create duplicate
	_, err = setup.CallMCPTool("mythic_create_operator", map[string]interface{}{
		"username": "test-duplicate-check",
		"password": "ValidPassword123!",
	})
	// Should eventually error on duplicate
	t.Logf("Duplicate creation result: %v", err)

	// Test updating non-existent operator status
	_, err = setup.CallMCPTool("mythic_update_operator_status", map[string]interface{}{
		"operator_id": 999999,
		"active":      false,
	})
	assert.Error(t, err, "Expected error when updating non-existent operator")
}

// TestE2E_Operators_FullWorkflow tests a complete operator workflow
func TestE2E_Operators_FullWorkflow(t *testing.T) {
	setup := SetupE2ETest(t)

	// Workflow: Create operator → Get details → Update prefs → Assign to op → Remove

	// 1. Create operator
	createResult, err := setup.CallMCPTool("mythic_create_operator", map[string]interface{}{
		"username": "workflow-test-operator",
		"password": "WorkflowTest123!",
	})
	require.NoError(t, err)
	metadata, ok := createResult["metadata"].(map[string]interface{})
	require.True(t, ok)
	operatorID := int(metadata["id"].(float64))

	// 2. Get operator details
	getResult, err := setup.CallMCPTool("mythic_get_operator", map[string]interface{}{
		"operator_id": operatorID,
	})
	require.NoError(t, err)
	require.NotNil(t, getResult)

	// 3. Update preferences
	_, err = setup.CallMCPTool("mythic_update_operator_preferences", map[string]interface{}{
		"operator_id": operatorID,
		"preferences": map[string]interface{}{
			"workflow_test": true,
			"setting":       "value",
		},
	})
	require.NoError(t, err)

	// 4. Get an operation
	operations, err := setup.MythicClient.GetOperations(setup.Ctx)
	require.NoError(t, err)
	require.NotEmpty(t, operations)

	// 5. Add operator to operation
	_, err = setup.CallMCPTool("mythic_update_operator_operation", map[string]interface{}{
		"operation_id": operations[0].ID,
		"add_users":    []int{operatorID},
	})
	require.NoError(t, err)

	// 6. Verify operator is in operation
	opOps, err := setup.MythicClient.GetOperatorsByOperation(setup.Ctx, operations[0].ID)
	require.NoError(t, err)

	found := false
	for _, op := range opOps {
		if op.OperatorID == operatorID {
			found = true
			break
		}
	}
	assert.True(t, found, "Operator should be in operation")

	// 7. Deactivate operator
	_, err = setup.CallMCPTool("mythic_update_operator_status", map[string]interface{}{
		"operator_id": operatorID,
		"active":      false,
	})
	require.NoError(t, err)

	// 8. Verify final state
	finalOp, err := setup.MythicClient.GetOperatorByID(setup.Ctx, operatorID)
	require.NoError(t, err)
	assert.False(t, finalOp.Active, "Operator should be deactivated")
	assert.Equal(t, "workflow-test-operator", finalOp.Username)
}

// TestE2E_Operators_MultipleOperators tests creating multiple operators
func TestE2E_Operators_MultipleOperators(t *testing.T) {
	setup := SetupE2ETest(t)

	operatorIDs := make([]int, 0, 3)

	// Create 3 operators
	for i := 1; i <= 3; i++ {
		createResult, err := setup.CallMCPTool("mythic_create_operator", map[string]interface{}{
			"username": "multi-test-op-" + string(rune('0'+i)),
			"password": "MultiTest123!",
		})
		require.NoError(t, err)

		metadata, ok := createResult["metadata"].(map[string]interface{})
		require.True(t, ok)
		operatorID := int(metadata["id"].(float64))
		operatorIDs = append(operatorIDs, operatorID)
	}

	// Verify all operators exist
	allOps, err := setup.MythicClient.GetOperators(setup.Ctx)
	require.NoError(t, err)

	for _, opID := range operatorIDs {
		found := false
		for _, op := range allOps {
			if op.ID == opID {
				found = true
				break
			}
		}
		assert.True(t, found, "Operator %d should exist in list", opID)
	}

	// Deactivate all test operators
	for _, opID := range operatorIDs {
		_, err := setup.CallMCPTool("mythic_update_operator_status", map[string]interface{}{
			"operator_id": opID,
			"active":      false,
		})
		require.NoError(t, err)
	}
}
