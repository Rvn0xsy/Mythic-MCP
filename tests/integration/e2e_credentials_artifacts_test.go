package integration

import (
	"testing"

	"github.com/nbaertsch/mythic-sdk-go/pkg/mythic/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_Credentials_CreateAndManage tests credential lifecycle
func TestE2E_Credentials_CreateAndManage(t *testing.T) {
	setup := SetupE2ETest(t)

	// Step 1: Create a credential
	createResult, err := setup.CallMCPTool("mythic_create_credential", map[string]interface{}{
		"type":       "plaintext",
		"account":    "test-user",
		"realm":      "test.example.com",
		"credential": "test-password-123",
		"comment":    "E2E test credential",
	})
	require.NoError(t, err)
	require.NotNil(t, createResult)

	// Extract credential ID
	metadata, ok := createResult["metadata"].(map[string]interface{})
	require.True(t, ok, "Expected metadata in result")
	credIDFloat, ok := metadata["id"].(float64)
	require.True(t, ok, "Expected credential ID in metadata")
	credID := int(credIDFloat)

	// Step 2: Get credential by ID
	getResult, err := setup.CallMCPTool("mythic_get_credential", map[string]interface{}{
		"credential_id": credID,
	})
	require.NoError(t, err)
	require.NotNil(t, getResult)

	// Step 3: Update credential
	updateResult, err := setup.CallMCPTool("mythic_update_credential", map[string]interface{}{
		"credential_id": credID,
		"comment":       "Updated E2E test credential",
	})
	require.NoError(t, err)
	require.NotNil(t, updateResult)

	// Verify update
	cred, err := setup.MythicClient.GetCredentialByID(setup.Ctx, credID)
	require.NoError(t, err)
	assert.Equal(t, "Updated E2E test credential", cred.Comment)

	// Step 4: Delete credential
	deleteResult, err := setup.CallMCPTool("mythic_delete_credential", map[string]interface{}{
		"credential_id": credID,
	})
	require.NoError(t, err)
	require.NotNil(t, deleteResult)
}

// TestE2E_Credentials_GetCredentials tests listing credentials
func TestE2E_Credentials_GetCredentials(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all credentials
	result, err := setup.CallMCPTool("mythic_get_credentials", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Result should contain credentials array
	_, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
}

// TestE2E_Credentials_GetByOperation tests filtering by operation
func TestE2E_Credentials_GetByOperation(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get an operation
	operations, err := setup.MythicClient.GetOperations(setup.Ctx)
	require.NoError(t, err)
	require.NotEmpty(t, operations)
	operationID := operations[0].ID

	// Get credentials for this operation
	result, err := setup.CallMCPTool("mythic_get_operation_credentials", map[string]interface{}{
		"operation_id": operationID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	_, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
}

// TestE2E_Artifacts_CreateAndManage tests artifact lifecycle
func TestE2E_Artifacts_CreateAndManage(t *testing.T) {
	setup := SetupE2ETest(t)

	// Step 1: Create an artifact
	createResult, err := setup.CallMCPTool("mythic_create_artifact", map[string]interface{}{
		"artifact": "C:\\\\Windows\\\\System32\\\\test.exe",
		"host":     "test-host.local",
	})
	require.NoError(t, err)
	require.NotNil(t, createResult)

	// Extract artifact ID
	metadata, ok := createResult["metadata"].(map[string]interface{})
	require.True(t, ok, "Expected metadata in result")
	artifactIDFloat, ok := metadata["id"].(float64)
	require.True(t, ok, "Expected artifact ID in metadata")
	artifactID := int(artifactIDFloat)

	// Step 2: Get artifact by ID
	getResult, err := setup.CallMCPTool("mythic_get_artifact", map[string]interface{}{
		"artifact_id": artifactID,
	})
	require.NoError(t, err)
	require.NotNil(t, getResult)

	// Step 3: Update artifact (only host can be updated)
	updateResult, err := setup.CallMCPTool("mythic_update_artifact", map[string]interface{}{
		"artifact_id": artifactID,
		"host":        "updated-host.local",
	})
	require.NoError(t, err)
	require.NotNil(t, updateResult)

	// Verify update
	artifact, err := setup.MythicClient.GetArtifactByID(setup.Ctx, artifactID)
	require.NoError(t, err)
	assert.Equal(t, "updated-host.local", artifact.Host)

	// Step 4: Delete artifact
	deleteResult, err := setup.CallMCPTool("mythic_delete_artifact", map[string]interface{}{
		"artifact_id": artifactID,
	})
	require.NoError(t, err)
	require.NotNil(t, deleteResult)
}

// TestE2E_Artifacts_GetArtifacts tests listing artifacts
func TestE2E_Artifacts_GetArtifacts(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all artifacts
	result, err := setup.CallMCPTool("mythic_get_artifacts", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, result)

	_, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
}

// TestE2E_Artifacts_GetByOperation tests filtering by operation
func TestE2E_Artifacts_GetByOperation(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get an operation
	operations, err := setup.MythicClient.GetOperations(setup.Ctx)
	require.NoError(t, err)
	require.NotEmpty(t, operations)
	operationID := operations[0].ID

	// Get artifacts for this operation
	result, err := setup.CallMCPTool("mythic_get_operation_artifacts", map[string]interface{}{
		"operation_id": operationID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	_, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
}

// TestE2E_Artifacts_GetByHost tests filtering by host
func TestE2E_Artifacts_GetByHost(t *testing.T) {
	setup := SetupE2ETest(t)

	// Create an artifact with specific host
	hostStr := "test-host-for-filter"
	artifact, err := setup.MythicClient.CreateArtifact(setup.Ctx, &types.CreateArtifactRequest{
		Artifact: "test-artifact-path",
		Host:     &hostStr,
	})
	require.NoError(t, err)

	// Get artifacts for this host
	result, err := setup.CallMCPTool("mythic_get_host_artifacts", map[string]interface{}{
		"host": "test-host-for-filter",
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should contain at least our artifact
	_, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	require.NotEmpty(t, content, "Expected at least one artifact")

	// Cleanup
	setup.MythicClient.DeleteArtifact(setup.Ctx, artifact.ID)
}

// TestE2E_Artifacts_GetByType tests filtering by artifact type
func TestE2E_Artifacts_GetByType(t *testing.T) {
	setup := SetupE2ETest(t)

	// Create an artifact
	hostStr := "test-host"
	artifact, err := setup.MythicClient.CreateArtifact(setup.Ctx, &types.CreateArtifactRequest{
		Artifact: "test-registry-key",
		Host:     &hostStr,
	})
	require.NoError(t, err)

	// Get artifacts of a type (note: artifact types are user-defined, not fixed)
	result, err := setup.CallMCPTool("mythic_get_artifacts_by_type", map[string]interface{}{
		"artifact_type": "file",
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Result should be an array (may be empty)
	_, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")

	// Cleanup
	setup.MythicClient.DeleteArtifact(setup.Ctx, artifact.ID)
}

// TestE2E_CredentialsArtifacts_ErrorHandling tests error scenarios
func TestE2E_CredentialsArtifacts_ErrorHandling(t *testing.T) {
	setup := SetupE2ETest(t)

	// Test getting non-existent credential
	_, err := setup.CallMCPTool("mythic_get_credential", map[string]interface{}{
		"credential_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting non-existent credential")

	// Test getting non-existent artifact
	_, err = setup.CallMCPTool("mythic_get_artifact", map[string]interface{}{
		"artifact_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting non-existent artifact")

	// Test creating credential with invalid type
	_, err = setup.CallMCPTool("mythic_create_credential", map[string]interface{}{
		"type":       "", // Empty type
		"account":    "test",
		"credential": "test",
	})
	assert.Error(t, err, "Expected error with empty credential type")

	// Test creating artifact with empty artifact string
	_, err = setup.CallMCPTool("mythic_create_artifact", map[string]interface{}{
		"artifact": "",
	})
	assert.Error(t, err, "Expected error with empty artifact")

	// Test updating non-existent credential
	_, err = setup.CallMCPTool("mythic_update_credential", map[string]interface{}{
		"credential_id": 999999,
		"comment":       "Should fail",
	})
	assert.Error(t, err, "Expected error when updating non-existent credential")

	// Test updating non-existent artifact
	_, err = setup.CallMCPTool("mythic_update_artifact", map[string]interface{}{
		"artifact_id": 999999,
		"host":        "Should fail",
	})
	assert.Error(t, err, "Expected error when updating non-existent artifact")

	// Test deleting non-existent credential
	_, err = setup.CallMCPTool("mythic_delete_credential", map[string]interface{}{
		"credential_id": 999999,
	})
	assert.Error(t, err, "Expected error when deleting non-existent credential")

	// Test deleting non-existent artifact
	_, err = setup.CallMCPTool("mythic_delete_artifact", map[string]interface{}{
		"artifact_id": 999999,
	})
	assert.Error(t, err, "Expected error when deleting non-existent artifact")
}

// TestE2E_CredentialsArtifacts_FullWorkflow tests complete workflow
func TestE2E_CredentialsArtifacts_FullWorkflow(t *testing.T) {
	setup := SetupE2ETest(t)

	// Workflow: Create credential → Create artifact → Query both → Update → Delete

	// 1. Create credential
	credResult, err := setup.CallMCPTool("mythic_create_credential", map[string]interface{}{
		"type":       "plaintext",
		"account":    "workflow-user",
		"realm":      "workflow.test",
		"credential": "workflow-pass",
		"comment":    "Workflow test",
	})
	require.NoError(t, err)
	credMeta, ok := credResult["metadata"].(map[string]interface{})
	require.True(t, ok)
	credID := int(credMeta["id"].(float64))

	// 2. Create artifact
	artResult, err := setup.CallMCPTool("mythic_create_artifact", map[string]interface{}{
		"artifact": "C:\\\\Temp\\\\workflow.exe",
		"host":     "workflow-host",
	})
	require.NoError(t, err)
	artMeta, ok := artResult["metadata"].(map[string]interface{})
	require.True(t, ok)
	artID := int(artMeta["id"].(float64))

	// 3. Query all credentials
	allCredsResult, err := setup.CallMCPTool("mythic_get_credentials", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, allCredsResult)

	// 4. Query all artifacts
	allArtsResult, err := setup.CallMCPTool("mythic_get_artifacts", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, allArtsResult)

	// 5. Update credential
	_, err = setup.CallMCPTool("mythic_update_credential", map[string]interface{}{
		"credential_id": credID,
		"comment":       "Updated workflow test",
	})
	require.NoError(t, err)

	// 6. Update artifact (only host can be updated)
	_, err = setup.CallMCPTool("mythic_update_artifact", map[string]interface{}{
		"artifact_id": artID,
		"host":        "workflow-host-updated",
	})
	require.NoError(t, err)

	// 7. Verify updates
	finalCred, err := setup.MythicClient.GetCredentialByID(setup.Ctx, credID)
	require.NoError(t, err)
	assert.Equal(t, "Updated workflow test", finalCred.Comment)

	finalArt, err := setup.MythicClient.GetArtifactByID(setup.Ctx, artID)
	require.NoError(t, err)
	assert.Equal(t, "workflow-host-updated", finalArt.Host)

	// 8. Cleanup
	_, err = setup.CallMCPTool("mythic_delete_credential", map[string]interface{}{
		"credential_id": credID,
	})
	require.NoError(t, err)

	_, err = setup.CallMCPTool("mythic_delete_artifact", map[string]interface{}{
		"artifact_id": artID,
	})
	require.NoError(t, err)
}
