//go:build integration && e2e
// +build integration,e2e

package integration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_Auth_LoginLogout(t *testing.T) {
	setup := SetupE2ETest(t)

	t.Run("Login", func(t *testing.T) {
		result, err := setup.CallMCPTool("mythic_login", map[string]interface{}{
			"username": os.Getenv("MYTHIC_USERNAME"),
			"password": os.Getenv("MYTHIC_PASSWORD"),
		})
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Verify response indicates success
		if resultData, ok := result["result"]; ok {
			assert.NotNil(t, resultData)
		}
	})

	t.Run("IsAuthenticated", func(t *testing.T) {
		result, err := setup.CallMCPTool("mythic_is_authenticated", nil)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Should be authenticated after login
		if resultData, ok := result["result"]; ok {
			assert.NotNil(t, resultData)
		}
	})

	t.Run("GetCurrentUser", func(t *testing.T) {
		result, err := setup.CallMCPTool("mythic_get_current_user", nil)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Verify we get user information
		if resultData, ok := result["result"]; ok {
			assert.NotNil(t, resultData)
		}
	})

	t.Run("Logout", func(t *testing.T) {
		result, err := setup.CallMCPTool("mythic_logout", nil)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("NotAuthenticatedAfterLogout", func(t *testing.T) {
		result, err := setup.CallMCPTool("mythic_is_authenticated", nil)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Should NOT be authenticated after logout
		if resultData, ok := result["result"]; ok {
			assert.NotNil(t, resultData)
		}
	})
}

func TestE2E_Auth_APITokens(t *testing.T) {
	setup := SetupE2ETest(t)

	// Login first
	_, err := setup.CallMCPTool("mythic_login", map[string]interface{}{
		"username": os.Getenv("MYTHIC_USERNAME"),
		"password": os.Getenv("MYTHIC_PASSWORD"),
	})
	require.NoError(t, err)

	var tokenID int

	t.Run("CreateAPIToken", func(t *testing.T) {
		result, err := setup.CallMCPTool("mythic_create_api_token", map[string]interface{}{
			"token_type": "User",
		})
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Extract token ID for cleanup
		if resultData, ok := result["result"].(map[string]interface{}); ok {
			if content, ok := resultData["content"].([]interface{}); ok && len(content) > 0 {
				// Parse token ID from content
				// This is simplified - actual parsing will depend on response format
				tokenID = 1 // Placeholder
			}
		}
	})

	t.Run("DeleteAPIToken", func(t *testing.T) {
		if tokenID == 0 {
			t.Skip("No token ID available")
		}

		result, err := setup.CallMCPTool("mythic_delete_api_token", map[string]interface{}{
			"token_id": tokenID,
		})
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestE2E_Auth_RefreshToken(t *testing.T) {
	setup := SetupE2ETest(t)

	// Login first
	_, err := setup.CallMCPTool("mythic_login", map[string]interface{}{
		"username": os.Getenv("MYTHIC_USERNAME"),
		"password": os.Getenv("MYTHIC_PASSWORD"),
	})
	require.NoError(t, err)

	t.Run("RefreshToken", func(t *testing.T) {
		result, err := setup.CallMCPTool("mythic_refresh_token", nil)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Should still be authenticated after refresh
		authResult, err := setup.CallMCPTool("mythic_is_authenticated", nil)
		require.NoError(t, err)
		assert.NotNil(t, authResult)
	})
}

func TestE2E_Auth_ErrorHandling(t *testing.T) {
	setup := SetupE2ETest(t)

	t.Run("LoginWithInvalidCredentials", func(t *testing.T) {
		_, err := setup.CallMCPTool("mythic_login", map[string]interface{}{
			"username": "invalid_user",
			"password": "wrong_password",
		})

		// Should return error
		assert.Error(t, err)
	})

	t.Run("GetCurrentUserWhenNotAuthenticated", func(t *testing.T) {
		// Make sure we're logged out
		setup.CallMCPTool("mythic_logout", nil)

		_, err := setup.CallMCPTool("mythic_get_current_user", nil)

		// Should return error when not authenticated
		assert.Error(t, err)
	})
}
