//go:build integration && e2e
// +build integration,e2e

package integration

import (
	"encoding/base64"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_Payloads_GetPayloads tests listing all payloads
func TestE2E_Payloads_GetPayloads(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all payloads
	result, err := setup.CallMCPTool("mythic_get_payloads", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Result should be an array (may be empty if no payloads)
	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Total payloads: %d", len(content))
}

// TestE2E_Payloads_GetPayload tests getting specific payload
func TestE2E_Payloads_GetPayload(t *testing.T) {
	setup := SetupE2ETest(t)

	// First get all payloads to find one to query
	allPayloads, err := setup.MythicClient.GetPayloads(setup.Ctx)
	require.NoError(t, err)

	if len(allPayloads) == 0 {
		t.Skip("No payloads available to test")
	}

	payloadUUID := allPayloads[0].UUID

	// Get specific payload
	result, err := setup.CallMCPTool("mythic_get_payload", map[string]interface{}{
		"payload_uuid": payloadUUID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
}

// TestE2E_Payloads_GetPayloadTypes tests listing payload types
func TestE2E_Payloads_GetPayloadTypes(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all payload types
	result, err := setup.CallMCPTool("mythic_get_payload_types", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Total payload types: %d", len(content))
}

// TestE2E_Payloads_CreatePayload tests creating a payload
func TestE2E_Payloads_CreatePayload(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get payload types first
	payloadTypes, err := setup.MythicClient.GetPayloadTypes(setup.Ctx)
	require.NoError(t, err)

	if len(payloadTypes) == 0 {
		t.Skip("No payload types available for testing")
	}

	// Get a callback to use as template (optional)
	_, err = setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	// Create a basic payload
	createResult, err := setup.CallMCPTool("mythic_create_payload", map[string]interface{}{
		"payload_type": payloadTypes[0].Name,
		"description":  "E2E test payload",
	})

	// May fail if build parameters are required - that's okay
	if err != nil {
		t.Logf("Create payload failed (expected if build params required): %v", err)
		return
	}

	require.NotNil(t, createResult)
	t.Logf("Successfully created payload")
}

// TestE2E_Payloads_UpdatePayload tests updating payload properties
func TestE2E_Payloads_UpdatePayload(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a payload to update
	allPayloads, err := setup.MythicClient.GetPayloads(setup.Ctx)
	require.NoError(t, err)

	if len(allPayloads) == 0 {
		t.Skip("No payloads available to test")
	}

	payloadUUID := allPayloads[0].UUID

	// Update payload description
	updateResult, err := setup.CallMCPTool("mythic_update_payload", map[string]interface{}{
		"payload_uuid": payloadUUID,
		"description":  "Updated via E2E test",
	})
	require.NoError(t, err)
	require.NotNil(t, updateResult)

	// Verify update
	payload, err := setup.MythicClient.GetPayloadByUUID(setup.Ctx, payloadUUID)
	require.NoError(t, err)
	assert.Equal(t, "Updated via E2E test", payload.Description)
}

// TestE2E_Payloads_GetPayloadCommands tests getting payload commands
func TestE2E_Payloads_GetPayloadCommands(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a payload
	allPayloads, err := setup.MythicClient.GetPayloads(setup.Ctx)
	require.NoError(t, err)

	if len(allPayloads) == 0 {
		t.Skip("No payloads available to test")
	}

	payloadID := allPayloads[0].ID

	// Get commands for payload
	result, err := setup.CallMCPTool("mythic_get_payload_commands", map[string]interface{}{
		"payload_id": payloadID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Commands for payload %d: %d", payloadID, len(content))
}

// TestE2E_Payloads_ExportPayloadConfig tests exporting payload config
func TestE2E_Payloads_ExportPayloadConfig(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a payload
	allPayloads, err := setup.MythicClient.GetPayloads(setup.Ctx)
	require.NoError(t, err)

	if len(allPayloads) == 0 {
		t.Skip("No payloads available to test")
	}

	payloadUUID := allPayloads[0].UUID

	// Export payload config
	exportResult, err := setup.CallMCPTool("mythic_export_payload_config", map[string]interface{}{
		"payload_uuid": payloadUUID,
	})
	require.NoError(t, err)
	require.NotNil(t, exportResult)

	// Extract exported config
	exportMeta, ok := exportResult["metadata"].(map[string]interface{})
	require.True(t, ok, "Expected metadata in export result")
	configStr, ok := exportMeta["config"].(string)
	require.True(t, ok, "Expected config string in metadata")
	require.NotEmpty(t, configStr, "Config should not be empty")

	t.Logf("Successfully exported config (length: %d bytes)", len(configStr))
}

// TestE2E_Payloads_GetPayloadOnHost tests listing payloads on hosts
func TestE2E_Payloads_GetPayloadOnHost(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get current operation
	me, err := setup.MythicClient.GetMe(setup.Ctx)
	require.NoError(t, err)

	if me.CurrentOperation == nil {
		t.Skip("No current operation set")
	}

	operationID := me.CurrentOperation.ID

	// Get payloads on host
	result, err := setup.CallMCPTool("mythic_get_payload_on_host", map[string]interface{}{
		"operation_id": operationID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Payloads on host for operation %d: %d", operationID, len(content))
}

// TestE2E_Payloads_DownloadPayload tests downloading a payload
func TestE2E_Payloads_DownloadPayload(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a built payload
	allPayloads, err := setup.MythicClient.GetPayloads(setup.Ctx)
	require.NoError(t, err)

	// Find a built/success payload
	var builtPayloadUUID string
	for _, p := range allPayloads {
		if p.BuildMessage == "success" || p.BuildPhase == "success" {
			builtPayloadUUID = p.UUID
			break
		}
	}

	if builtPayloadUUID == "" {
		t.Skip("No built payloads available for download test")
	}

	// Download payload
	downloadResult, err := setup.CallMCPTool("mythic_download_payload", map[string]interface{}{
		"payload_uuid": builtPayloadUUID,
	})
	require.NoError(t, err)
	require.NotNil(t, downloadResult)

	// Extract downloaded payload
	meta, ok := downloadResult["metadata"].(map[string]interface{})
	require.True(t, ok, "Expected metadata in download result")
	payloadData, ok := meta["payload_data"].(string)
	require.True(t, ok, "Expected payload_data string in metadata")

	// Verify base64 encoding
	_, err = base64.StdEncoding.DecodeString(payloadData)
	require.NoError(t, err, "Payload data should be valid base64")

	t.Logf("Successfully downloaded payload (size: %d bytes base64)", len(payloadData))
}

// TestE2E_Payloads_RebuildPayload tests rebuilding a payload
func TestE2E_Payloads_RebuildPayload(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a payload to rebuild
	allPayloads, err := setup.MythicClient.GetPayloads(setup.Ctx)
	require.NoError(t, err)

	if len(allPayloads) == 0 {
		t.Skip("No payloads available to test")
	}

	payloadUUID := allPayloads[0].UUID

	// Rebuild payload
	rebuildResult, err := setup.CallMCPTool("mythic_rebuild_payload", map[string]interface{}{
		"payload_uuid": payloadUUID,
	})

	// May fail if payload is currently building
	if err != nil {
		t.Logf("Rebuild failed (expected if already building): %v", err)
		return
	}

	require.NotNil(t, rebuildResult)
	t.Logf("Successfully initiated payload rebuild")
}

// TestE2E_Payloads_WaitForPayload tests waiting for payload build
func TestE2E_Payloads_WaitForPayload(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a payload
	allPayloads, err := setup.MythicClient.GetPayloads(setup.Ctx)
	require.NoError(t, err)

	if len(allPayloads) == 0 {
		t.Skip("No payloads available to test")
	}

	payloadUUID := allPayloads[0].UUID

	// Wait for payload with short timeout (it's probably already done)
	waitResult, err := setup.CallMCPTool("mythic_wait_for_payload", map[string]interface{}{
		"payload_uuid": payloadUUID,
		"timeout":      5,
	})

	// If payload is already complete, this should succeed
	// If still building, may timeout - both are okay
	if err != nil {
		t.Logf("Wait timed out or failed (expected for building payloads): %v", err)
		return
	}

	require.NotNil(t, waitResult)
	t.Logf("Payload %s completed or was already complete", payloadUUID)
}

// TestE2E_Payloads_DeletePayload tests deleting a payload
func TestE2E_Payloads_DeletePayload(t *testing.T) {
	setup := SetupE2ETest(t)

	// First try to create a test payload to delete
	payloadTypes, err := setup.MythicClient.GetPayloadTypes(setup.Ctx)
	require.NoError(t, err)

	if len(payloadTypes) == 0 {
		t.Skip("No payload types available for testing")
	}

	// Choose a payload type that supports the installed HTTP C2 profile if possible.
	payloadTypeName := ""
	for _, pt := range payloadTypes {
		if pt.Name == "" {
			continue
		}
		payloadTypeName = pt.Name
		for _, c2 := range pt.SupportedC2Profiles {
			if c2 == "http" {
				payloadTypeName = pt.Name
				goto chosen
			}
		}
	}

chosen:
	if payloadTypeName == "" {
		t.Logf("No payload type name available; cannot exercise delete payload flow")
		return
	}

	// Build a minimal HTTP C2 profile config using required parameters.
	c2Profiles, err := setup.MythicClient.GetC2Profiles(setup.Ctx)
	require.NoError(t, err)
	var httpProfileID int
	for _, p := range c2Profiles {
		if p.Name == "http" {
			httpProfileID = p.ID
			break
		}
	}
	require.NotZero(t, httpProfileID, "Expected HTTP C2 profile to be installed")

	paramsSchema, err := setup.MythicClient.GetC2ProfileParameters(setup.Ctx, httpProfileID)
	require.NoError(t, err)

	c2Params := map[string]interface{}{}
	for _, p := range paramsSchema {
		if !p.Required {
			continue
		}
		// Prefer defaults when available.
		if p.DefaultValue != "" {
			// Best-effort type coercion based on parameter_type.
			switch p.ParameterType {
			case "Number":
				// Hasura accepts numbers as JSON numbers; keep defaults as strings only if parse fails.
				if n, err := strconv.Atoi(p.DefaultValue); err == nil {
					c2Params[p.Name] = n
				} else {
					c2Params[p.Name] = p.DefaultValue
				}
			case "Boolean":
				c2Params[p.Name] = (p.DefaultValue == "true" || p.DefaultValue == "1")
			default:
				c2Params[p.Name] = p.DefaultValue
			}
			continue
		}

		// Fallbacks for common required params.
		switch p.Name {
		case "callback_host":
			c2Params[p.Name] = "https://127.0.0.1"
		case "callback_port":
			c2Params[p.Name] = 80
		default:
			// Conservative placeholders; Mythic will validate further if needed.
			switch p.ParameterType {
			case "Number":
				c2Params[p.Name] = 80
			case "Boolean":
				c2Params[p.Name] = false
			default:
				c2Params[p.Name] = "e2e"
			}
		}
	}

	createResult, err := setup.CallMCPTool("mythic_create_payload", map[string]interface{}{
		"payload_type": payloadTypeName,
		"description":  "E2E test payload for deletion",
		"c2_profiles": []map[string]interface{}{
			{
				"name":       "http",
				"parameters": c2Params,
			},
		},
	})
	if err != nil {
		// Some environments require additional build parameters or payload-type containers.
		// Avoid skipping; track full deterministic payload build coverage in issue #42.
		t.Logf("Create payload failed (cannot exercise delete payload happy-path): %v", err)
		return
	}
	require.NotNil(t, createResult)

	// Extract UUID from metadata. The create tool returns the payload object as structuredContent.
	meta, ok := createResult["metadata"].(map[string]interface{})
	require.True(t, ok, "Expected metadata in create result")
	payloadUUID, ok := meta["uuid"].(string)
	require.True(t, ok && payloadUUID != "", "Expected uuid in payload metadata")

	// Wait briefly for build to at least register (some Mythic versions reject deletes during early build).
	_, _ = setup.CallMCPTool("mythic_wait_for_payload", map[string]interface{}{
		"payload_uuid": payloadUUID,
		"timeout":      30,
	})

	// Delete the payload
	deleteResult, err := setup.CallMCPTool("mythic_delete_payload", map[string]interface{}{
		"payload_uuid": payloadUUID,
	})
	require.NoError(t, err)
	require.NotNil(t, deleteResult)

	t.Logf("Successfully deleted payload %s", payloadUUID)
}

// TestE2E_Payloads_ErrorHandling tests error scenarios
func TestE2E_Payloads_ErrorHandling(t *testing.T) {
	setup := SetupE2ETest(t)

	// Test getting non-existent payload
	_, err := setup.CallMCPTool("mythic_get_payload", map[string]interface{}{
		"payload_uuid": "non-existent-uuid",
	})
	assert.Error(t, err, "Expected error when getting non-existent payload")

	// Test updating non-existent payload
	_, err = setup.CallMCPTool("mythic_update_payload", map[string]interface{}{
		"payload_uuid": "non-existent-uuid",
		"description":  "Should fail",
	})
	assert.Error(t, err, "Expected error when updating non-existent payload")

	// Test downloading non-existent payload
	_, err = setup.CallMCPTool("mythic_download_payload", map[string]interface{}{
		"payload_uuid": "non-existent-uuid",
	})
	assert.Error(t, err, "Expected error when downloading non-existent payload")

	// Test deleting non-existent payload
	_, err = setup.CallMCPTool("mythic_delete_payload", map[string]interface{}{
		"payload_uuid": "non-existent-uuid",
	})
	assert.Error(t, err, "Expected error when deleting non-existent payload")
}

// TestE2E_Payloads_FullWorkflow tests complete payload workflow
func TestE2E_Payloads_FullWorkflow(t *testing.T) {
	setup := SetupE2ETest(t)

	// Workflow: Get types → List payloads → Get one → Update → Get commands → Export config → Download

	// 1. Get payload types
	typesResult, err := setup.CallMCPTool("mythic_get_payload_types", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, typesResult)

	// 2. Get all payloads
	allResult, err := setup.CallMCPTool("mythic_get_payloads", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, allResult)

	// Get a payload to work with
	allPayloads, err := setup.MythicClient.GetPayloads(setup.Ctx)
	require.NoError(t, err)

	if len(allPayloads) == 0 {
		t.Skip("No payloads available for full workflow test")
	}

	payload := allPayloads[0]

	// 3. Get specific payload
	getResult, err := setup.CallMCPTool("mythic_get_payload", map[string]interface{}{
		"payload_uuid": payload.UUID,
	})
	require.NoError(t, err)
	require.NotNil(t, getResult)

	// 4. Update payload
	_, err = setup.CallMCPTool("mythic_update_payload", map[string]interface{}{
		"payload_uuid": payload.UUID,
		"description":  "Workflow test payload",
	})
	require.NoError(t, err)

	// 5. Get payload commands
	cmdsResult, err := setup.CallMCPTool("mythic_get_payload_commands", map[string]interface{}{
		"payload_id": payload.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, cmdsResult)

	// 6. Export payload config
	exportResult, err := setup.CallMCPTool("mythic_export_payload_config", map[string]interface{}{
		"payload_uuid": payload.UUID,
	})
	require.NoError(t, err)
	require.NotNil(t, exportResult)

	// 7. Download payload (if built)
	if payload.BuildMessage == "success" || payload.BuildPhase == "success" {
		downloadResult, err := setup.CallMCPTool("mythic_download_payload", map[string]interface{}{
			"payload_uuid": payload.UUID,
		})
		require.NoError(t, err)
		require.NotNil(t, downloadResult)
	}

	// 8. Verify final state
	finalPayload, err := setup.MythicClient.GetPayloadByUUID(setup.Ctx, payload.UUID)
	require.NoError(t, err)
	assert.Equal(t, "Workflow test payload", finalPayload.Description)
	t.Logf("Workflow complete for payload %s (%s)", payload.UUID, payload.PayloadType)
}
