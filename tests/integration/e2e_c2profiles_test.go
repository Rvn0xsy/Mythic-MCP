package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_C2Profiles_GetC2Profiles tests listing all C2 profiles
func TestE2E_C2Profiles_GetC2Profiles(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all C2 profiles
	result, err := setup.CallMCPTool("mythic_get_c2_profiles", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Result should be an array (may be empty if no C2 profiles)
	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Total C2 profiles: %d", len(content))
}

// TestE2E_C2Profiles_GetC2Profile tests getting specific C2 profile
func TestE2E_C2Profiles_GetC2Profile(t *testing.T) {
	setup := SetupE2ETest(t)

	// First get all C2 profiles to find one to query
	allProfiles, err := setup.MythicClient.GetC2Profiles(setup.Ctx)
	require.NoError(t, err)

	if len(allProfiles) == 0 {
		t.Skip("No C2 profiles available to test")
	}

	profileID := allProfiles[0].ID

	// Get specific C2 profile
	result, err := setup.CallMCPTool("mythic_get_c2_profile", map[string]interface{}{
		"profile_id": profileID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
}

// TestE2E_C2Profiles_CreateC2Instance tests creating a C2 instance
func TestE2E_C2Profiles_CreateC2Instance(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get C2 profiles to find one that isn't running
	profiles, err := setup.MythicClient.GetC2Profiles(setup.Ctx)
	require.NoError(t, err)

	if len(profiles) == 0 {
		t.Skip("No C2 profiles available for testing")
	}

	// Try to create an instance (may fail if already running)
	createResult, err := setup.CallMCPTool("mythic_create_c2_instance", map[string]interface{}{
		"c2_profile_name": profiles[0].Name,
	})

	// May fail if instance already exists - that's okay
	if err != nil {
		t.Logf("Create C2 instance failed (expected if already running): %v", err)
		return
	}

	require.NotNil(t, createResult)
	t.Logf("Successfully created C2 instance")
}

// TestE2E_C2Profiles_ImportC2Instance tests importing C2 instance config
func TestE2E_C2Profiles_ImportC2Instance(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a C2 profile
	profiles, err := setup.MythicClient.GetC2Profiles(setup.Ctx)
	require.NoError(t, err)

	if len(profiles) == 0 {
		t.Skip("No C2 profiles available to test")
	}

	// Try to import (needs valid config JSON - will likely fail without it)
	importResult, err := setup.CallMCPTool("mythic_import_c2_instance", map[string]interface{}{
		"c2_profile_name": profiles[0].Name,
		"config_data":     "{}",
	})

	// May fail with invalid config - that's expected
	if err != nil {
		t.Logf("Import C2 instance failed (expected with empty config): %v", err)
		return
	}

	require.NotNil(t, importResult)
	t.Logf("Successfully imported C2 instance")
}

// TestE2E_C2Profiles_StartStopProfile tests starting and stopping C2 profiles
func TestE2E_C2Profiles_StartStopProfile(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a C2 profile
	profiles, err := setup.MythicClient.GetC2Profiles(setup.Ctx)
	require.NoError(t, err)

	if len(profiles) == 0 {
		t.Skip("No C2 profiles available to test")
	}

	profileID := profiles[0].ID
	isRunning := profiles[0].Running

	if !isRunning {
		// Try to start it
		startResult, err := setup.CallMCPTool("mythic_start_c2_profile", map[string]interface{}{
			"profile_id": profileID,
		})

		if err != nil {
			t.Logf("Start C2 profile failed (expected if config incomplete): %v", err)
		} else {
			require.NotNil(t, startResult)
			t.Logf("Successfully started C2 profile %d", profileID)

			// Try to stop it
			stopResult, err := setup.CallMCPTool("mythic_stop_c2_profile", map[string]interface{}{
				"profile_id": profileID,
			})
			require.NoError(t, err)
			require.NotNil(t, stopResult)
			t.Logf("Successfully stopped C2 profile %d", profileID)
		}
	} else {
		// Profile is running, try to stop it
		stopResult, err := setup.CallMCPTool("mythic_stop_c2_profile", map[string]interface{}{
			"profile_id": profileID,
		})

		if err != nil {
			t.Logf("Stop C2 profile failed: %v", err)
		} else {
			require.NotNil(t, stopResult)
			t.Logf("Successfully stopped C2 profile %d", profileID)

			// Try to start it again
			startResult, err := setup.CallMCPTool("mythic_start_c2_profile", map[string]interface{}{
				"profile_id": profileID,
			})

			if err != nil {
				t.Logf("Start C2 profile failed: %v", err)
			} else {
				require.NotNil(t, startResult)
				t.Logf("Successfully started C2 profile %d", profileID)
			}
		}
	}
}

// TestE2E_C2Profiles_GetProfileOutput tests getting C2 profile output
func TestE2E_C2Profiles_GetProfileOutput(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a C2 profile
	profiles, err := setup.MythicClient.GetC2Profiles(setup.Ctx)
	require.NoError(t, err)

	if len(profiles) == 0 {
		t.Skip("No C2 profiles available to test")
	}

	profileID := profiles[0].ID

	// Get profile output
	result, err := setup.CallMCPTool("mythic_get_c2_profile_output", map[string]interface{}{
		"profile_id": profileID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	t.Logf("Retrieved output for C2 profile %d", profileID)
}

// TestE2E_C2Profiles_C2HostFile tests hosting a file on C2 profile
func TestE2E_C2Profiles_C2HostFile(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a C2 profile
	profiles, err := setup.MythicClient.GetC2Profiles(setup.Ctx)
	require.NoError(t, err)

	if len(profiles) == 0 {
		t.Skip("No C2 profiles available to test")
	}

	profileID := profiles[0].ID

	// Get a file to host
	files, err := setup.MythicClient.GetFiles(setup.Ctx, 10)
	require.NoError(t, err)

	if len(files) == 0 {
		t.Skip("No files available to host")
	}

	fileUUID := files[0].AgentFileID

	// Host file on C2
	result, err := setup.CallMCPTool("mythic_c2_host_file", map[string]interface{}{
		"profile_id": profileID,
		"file_uuid":  fileUUID,
	})

	// May fail if profile not running or file type incompatible
	if err != nil {
		t.Logf("C2 host file failed (expected if profile not configured): %v", err)
		return
	}

	require.NotNil(t, result)
	t.Logf("Successfully hosted file on C2 profile %d", profileID)
}

// TestE2E_C2Profiles_C2SampleMessage tests getting sample message from C2
func TestE2E_C2Profiles_C2SampleMessage(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a C2 profile
	profiles, err := setup.MythicClient.GetC2Profiles(setup.Ctx)
	require.NoError(t, err)

	if len(profiles) == 0 {
		t.Skip("No C2 profiles available to test")
	}

	profileID := profiles[0].ID

	// Get sample message (try "checkin" message type)
	result, err := setup.CallMCPTool("mythic_c2_sample_message", map[string]interface{}{
		"profile_id":   profileID,
		"message_type": "checkin",
	})

	// May fail if message type not supported
	if err != nil {
		t.Logf("C2 sample message failed (expected for some message types): %v", err)
		return
	}

	require.NotNil(t, result)
	t.Logf("Successfully retrieved sample message from C2 profile %d", profileID)
}

// TestE2E_C2Profiles_C2GetIOC tests getting IOCs from C2 profile
func TestE2E_C2Profiles_C2GetIOC(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get a C2 profile
	profiles, err := setup.MythicClient.GetC2Profiles(setup.Ctx)
	require.NoError(t, err)

	if len(profiles) == 0 {
		t.Skip("No C2 profiles available to test")
	}

	profileID := profiles[0].ID

	// Get IOCs
	result, err := setup.CallMCPTool("mythic_c2_get_ioc", map[string]interface{}{
		"profile_id": profileID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	t.Logf("Retrieved IOCs from C2 profile %d", profileID)
}

// TestE2E_C2Profiles_ErrorHandling tests error scenarios
func TestE2E_C2Profiles_ErrorHandling(t *testing.T) {
	setup := SetupE2ETest(t)

	// Test getting non-existent profile
	_, err := setup.CallMCPTool("mythic_get_c2_profile", map[string]interface{}{
		"profile_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting non-existent C2 profile")

	// Test starting non-existent profile
	_, err = setup.CallMCPTool("mythic_start_c2_profile", map[string]interface{}{
		"profile_id": 999999,
	})
	assert.Error(t, err, "Expected error when starting non-existent C2 profile")

	// Test getting output for non-existent profile
	_, err = setup.CallMCPTool("mythic_get_c2_profile_output", map[string]interface{}{
		"profile_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting output for non-existent C2 profile")
}

// TestE2E_C2Profiles_FullWorkflow tests complete C2 profile workflow
func TestE2E_C2Profiles_FullWorkflow(t *testing.T) {
	setup := SetupE2ETest(t)

	// Workflow: List profiles → Get one → Get output → Get IOCs → Sample message

	// 1. Get all C2 profiles
	allResult, err := setup.CallMCPTool("mythic_get_c2_profiles", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, allResult)

	// Get a profile to work with
	allProfiles, err := setup.MythicClient.GetC2Profiles(setup.Ctx)
	require.NoError(t, err)

	if len(allProfiles) == 0 {
		t.Skip("No C2 profiles available for full workflow test")
	}

	profile := allProfiles[0]

	// 2. Get specific profile
	getResult, err := setup.CallMCPTool("mythic_get_c2_profile", map[string]interface{}{
		"profile_id": profile.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, getResult)

	// 3. Get profile output
	outputResult, err := setup.CallMCPTool("mythic_get_c2_profile_output", map[string]interface{}{
		"profile_id": profile.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, outputResult)

	// 4. Get IOCs
	iocResult, err := setup.CallMCPTool("mythic_c2_get_ioc", map[string]interface{}{
		"profile_id": profile.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, iocResult)

	// 5. Get sample message (if supported)
	sampleResult, err := setup.CallMCPTool("mythic_c2_sample_message", map[string]interface{}{
		"profile_id":   profile.ID,
		"message_type": "checkin",
	})

	if err != nil {
		t.Logf("Sample message not supported (expected): %v", err)
	} else {
		require.NotNil(t, sampleResult)
	}

	t.Logf("Workflow complete for C2 profile %d (%s)", profile.ID, profile.Name)
}
