//go:build integration && e2e
// +build integration,e2e

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_Hosts_GetHosts tests listing all hosts in an operation
func TestE2E_Hosts_GetHosts(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get current operation
	me, err := setup.MythicClient.GetMe(setup.Ctx)
	require.NoError(t, err)

	if me.CurrentOperation == nil {
		t.Skip("No current operation set")
	}

	operationID := me.CurrentOperation.ID

	// Get hosts for operation
	result, err := setup.CallMCPTool("mythic_get_hosts", map[string]interface{}{
		"operation_id": operationID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Hosts in operation %d: %d", operationID, len(content))
}

// TestE2E_Hosts_GetHostByID tests getting a specific host by ID
func TestE2E_Hosts_GetHostByID(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get current operation
	me, err := setup.MythicClient.GetMe(setup.Ctx)
	require.NoError(t, err)

	if me.CurrentOperation == nil {
		t.Skip("No current operation set")
	}

	operationID := me.CurrentOperation.ID

	// Get hosts to find a valid ID
	hosts, err := setup.MythicClient.GetHosts(setup.Ctx, operationID)
	require.NoError(t, err)

	if len(hosts) == 0 {
		t.Skip("No hosts available to test")
	}

	hostID := hosts[0].ID

	// Get specific host by ID
	result, err := setup.CallMCPTool("mythic_get_host_by_id", map[string]interface{}{
		"host_id": hostID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	t.Logf("Retrieved host ID %d", hostID)
}

// TestE2E_Hosts_GetHostByHostname tests getting a host by hostname
func TestE2E_Hosts_GetHostByHostname(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get current operation
	me, err := setup.MythicClient.GetMe(setup.Ctx)
	require.NoError(t, err)

	if me.CurrentOperation == nil {
		t.Skip("No current operation set")
	}

	operationID := me.CurrentOperation.ID

	// Get hosts to find a valid hostname
	hosts, err := setup.MythicClient.GetHosts(setup.Ctx, operationID)
	require.NoError(t, err)

	if len(hosts) == 0 {
		t.Skip("No hosts available to test")
	}

	hostname := hosts[0].Hostname

	// Get host by hostname
	result, err := setup.CallMCPTool("mythic_get_host_by_hostname", map[string]interface{}{
		"hostname": hostname,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	t.Logf("Retrieved host by hostname: %s", hostname)
}

// TestE2E_Hosts_GetHostNetworkMap tests getting network map for operation
func TestE2E_Hosts_GetHostNetworkMap(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get current operation
	me, err := setup.MythicClient.GetMe(setup.Ctx)
	require.NoError(t, err)

	if me.CurrentOperation == nil {
		t.Skip("No current operation set")
	}

	operationID := me.CurrentOperation.ID

	// Get network map
	result, err := setup.CallMCPTool("mythic_get_host_network_map", map[string]interface{}{
		"operation_id": operationID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	t.Logf("Network map retrieved for operation %d", operationID)
}

// TestE2E_Hosts_GetCallbacksForHost tests listing callbacks on a host
func TestE2E_Hosts_GetCallbacksForHost(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get current operation
	me, err := setup.MythicClient.GetMe(setup.Ctx)
	require.NoError(t, err)

	if me.CurrentOperation == nil {
		t.Skip("No current operation set")
	}

	operationID := me.CurrentOperation.ID

	// Get hosts to find a valid host ID
	hosts, err := setup.MythicClient.GetHosts(setup.Ctx, operationID)
	require.NoError(t, err)

	if len(hosts) == 0 {
		t.Skip("No hosts available to test")
	}

	hostID := hosts[0].ID

	// Get callbacks for host
	result, err := setup.CallMCPTool("mythic_get_callbacks_for_host", map[string]interface{}{
		"host_id": hostID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	t.Logf("Callbacks on host %d: %d", hostID, len(content))
}

// TestE2E_Hosts_ErrorHandling tests error scenarios
func TestE2E_Hosts_ErrorHandling(t *testing.T) {
	setup := SetupE2ETest(t)

	// Test getting non-existent host by ID
	_, err := setup.CallMCPTool("mythic_get_host_by_id", map[string]interface{}{
		"host_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting non-existent host")

	// Test getting host by invalid hostname
	_, err = setup.CallMCPTool("mythic_get_host_by_hostname", map[string]interface{}{
		"hostname": "nonexistent-host-12345",
	})
	assert.Error(t, err, "Expected error when getting non-existent hostname")

	// Test getting callbacks for non-existent host
	_, err = setup.CallMCPTool("mythic_get_callbacks_for_host", map[string]interface{}{
		"host_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting callbacks for non-existent host")

	// Test getting hosts for non-existent operation
	_, err = setup.CallMCPTool("mythic_get_hosts", map[string]interface{}{
		"operation_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting hosts for non-existent operation")
}

// TestE2E_Hosts_FullWorkflow tests complete host workflow
func TestE2E_Hosts_FullWorkflow(t *testing.T) {
	setup := SetupE2ETest(t)

	// Workflow: Get hosts → Get specific host → Get callbacks → Get network map

	// Get current operation
	me, err := setup.MythicClient.GetMe(setup.Ctx)
	require.NoError(t, err)

	if me.CurrentOperation == nil {
		t.Skip("No current operation set for full workflow test")
	}

	operationID := me.CurrentOperation.ID

	// 1. Get all hosts in operation
	hostsResult, err := setup.CallMCPTool("mythic_get_hosts", map[string]interface{}{
		"operation_id": operationID,
	})
	require.NoError(t, err)
	require.NotNil(t, hostsResult)

	// Get hosts to work with
	hosts, err := setup.MythicClient.GetHosts(setup.Ctx, operationID)
	require.NoError(t, err)

	if len(hosts) == 0 {
		t.Skip("No hosts available for full workflow test")
	}

	host := hosts[0]

	// 2. Get specific host by ID
	hostByIDResult, err := setup.CallMCPTool("mythic_get_host_by_id", map[string]interface{}{
		"host_id": host.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, hostByIDResult)

	// 3. Get specific host by hostname
	hostByNameResult, err := setup.CallMCPTool("mythic_get_host_by_hostname", map[string]interface{}{
		"hostname": host.Hostname,
	})
	require.NoError(t, err)
	require.NotNil(t, hostByNameResult)

	// 4. Get callbacks for host
	callbacksResult, err := setup.CallMCPTool("mythic_get_callbacks_for_host", map[string]interface{}{
		"host_id": host.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, callbacksResult)

	// 5. Get network map
	networkMapResult, err := setup.CallMCPTool("mythic_get_host_network_map", map[string]interface{}{
		"operation_id": operationID,
	})
	require.NoError(t, err)
	require.NotNil(t, networkMapResult)

	t.Logf("Workflow complete for host %s (ID: %d)", host.Hostname, host.ID)
}

// TestE2E_Hosts_HostDetails tests detailed host information
func TestE2E_Hosts_HostDetails(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get current operation
	me, err := setup.MythicClient.GetMe(setup.Ctx)
	require.NoError(t, err)

	if me.CurrentOperation == nil {
		t.Skip("No current operation set")
	}

	operationID := me.CurrentOperation.ID

	// Get all hosts
	hosts, err := setup.MythicClient.GetHosts(setup.Ctx, operationID)
	require.NoError(t, err)

	if len(hosts) == 0 {
		t.Skip("No hosts available to test")
	}

	// Log details for first few hosts
	for i, host := range hosts {
		if i >= 3 {
			break
		}

		t.Logf("Host %d:", host.ID)
		t.Logf("  - Hostname: %s", host.Hostname)
		t.Logf("  - IP: %s", host.IP)
		t.Logf("  - OS: %s", host.OS)
		t.Logf("  - Architecture: %s", host.Architecture)
	}
}
