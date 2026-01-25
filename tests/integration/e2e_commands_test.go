package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_Commands_GetCommands tests listing all commands
func TestE2E_Commands_GetCommands(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all commands
	result, err := setup.CallMCPTool("mythic_get_commands", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Result should be an array (may be empty if no commands)
	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Total commands: %d", len(content))
}

// TestE2E_Commands_GetCommandParameters tests listing all command parameters
func TestE2E_Commands_GetCommandParameters(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all command parameters
	result, err := setup.CallMCPTool("mythic_get_command_parameters", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Result should be an array
	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Total command parameters: %d", len(content))
}

// TestE2E_Commands_GetCommandWithParameters tests getting command with parameters
func TestE2E_Commands_GetCommandWithParameters(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get payload types first
	payloadTypes, err := setup.MythicClient.GetPayloadTypes(setup.Ctx)
	require.NoError(t, err)

	if len(payloadTypes) == 0 {
		t.Skip("No payload types available to test")
	}

	payloadTypeID := payloadTypes[0].ID

	// Get commands for this payload type
	commands, err := setup.MythicClient.GetCommands(setup.Ctx)
	require.NoError(t, err)

	// Find a command for this payload type
	var commandName string
	for _, cmd := range commands {
		if cmd.PayloadTypeID == payloadTypeID {
			commandName = cmd.Cmd
			break
		}
	}

	if commandName == "" {
		t.Skip("No commands available for this payload type")
	}

	// Get command with parameters
	result, err := setup.CallMCPTool("mythic_get_command_with_parameters", map[string]interface{}{
		"payload_type_id": payloadTypeID,
		"command_name":    commandName,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	t.Logf("Retrieved command '%s' with parameters for payload type %d", commandName, payloadTypeID)
}

// TestE2E_Commands_GetLoadedCommandsForCallback tests getting loaded commands
// Note: This is already implemented in callbacks tools, but we test it here too
func TestE2E_Commands_GetLoadedCommandsForCallback(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a callback
	callbacks, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)

	if len(callbacks) == 0 {
		t.Skip("No callbacks available to test")
	}

	callbackID := callbacks[0].DisplayID

	// Get loaded commands (via callbacks tool)
	result, err := setup.CallMCPTool("mythic_get_loaded_commands", map[string]interface{}{
		"callback_id": callbackID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Loaded commands for callback %d: %d", callbackID, len(content))
}

// TestE2E_Commands_ErrorHandling tests error scenarios
func TestE2E_Commands_ErrorHandling(t *testing.T) {
	setup := SetupE2ETest(t)

	// Test getting command with invalid payload type
	_, err := setup.CallMCPTool("mythic_get_command_with_parameters", map[string]interface{}{
		"payload_type_id": 999999,
		"command_name":    "invalid",
	})
	assert.Error(t, err, "Expected error when getting command with invalid payload type")

	// Test getting loaded commands for non-existent callback
	_, err = setup.CallMCPTool("mythic_get_loaded_commands", map[string]interface{}{
		"callback_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting commands for non-existent callback")
}

// TestE2E_Commands_FullWorkflow tests complete command workflow
func TestE2E_Commands_FullWorkflow(t *testing.T) {
	setup := SetupE2ETest(t)

	// Workflow: Get all commands → Get parameters → Get specific command with parameters

	// 1. Get all commands
	allCommandsResult, err := setup.CallMCPTool("mythic_get_commands", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, allCommandsResult)

	// 2. Get all command parameters
	allParamsResult, err := setup.CallMCPTool("mythic_get_command_parameters", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, allParamsResult)

	// Get data to work with
	commands, err := setup.MythicClient.GetCommands(setup.Ctx)
	require.NoError(t, err)

	if len(commands) == 0 {
		t.Skip("No commands available for full workflow test")
	}

	command := commands[0]

	// 3. Get specific command with parameters
	cmdWithParamsResult, err := setup.CallMCPTool("mythic_get_command_with_parameters", map[string]interface{}{
		"payload_type_id": command.PayloadTypeID,
		"command_name":    command.Cmd,
	})
	require.NoError(t, err)
	require.NotNil(t, cmdWithParamsResult)

	t.Logf("Workflow complete for command '%s' (payload type ID: %d)", command.Cmd, command.PayloadTypeID)
}

// TestE2E_Commands_CommandDetails tests detailed command information
func TestE2E_Commands_CommandDetails(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all commands
	commands, err := setup.MythicClient.GetCommands(setup.Ctx)
	require.NoError(t, err)

	if len(commands) == 0 {
		t.Skip("No commands available to test")
	}

	// Test getting details for first command
	command := commands[0]

	// Get command with parameters
	cmdWithParams, err := setup.MythicClient.GetCommandWithParameters(setup.Ctx, command.PayloadTypeID, command.Cmd)
	require.NoError(t, err)
	require.NotNil(t, cmdWithParams)

	t.Logf("Command '%s' details:", command.Cmd)
	t.Logf("  - Version: %d", command.Version)
	t.Logf("  - Description: %s", command.Description)
	t.Logf("  - Help: %s", command.Help)
	t.Logf("  - Parameters: %d", len(cmdWithParams.Parameters))

	// Check if it's a raw string command
	if cmdWithParams.IsRawStringCommand() {
		t.Logf("  - Type: Raw string command")
	}

	// Check if it has required parameters
	if cmdWithParams.HasRequiredParameters() {
		t.Logf("  - Has required parameters: yes")
	} else {
		t.Logf("  - Has required parameters: no")
	}
}
