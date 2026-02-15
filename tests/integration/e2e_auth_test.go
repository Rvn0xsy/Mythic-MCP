//go:build integration && e2e
// +build integration,e2e

package integration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

const e2eFindAPITokenIDQuery = `
query FindAPITokenID($token_value: String!) {
  apitokens(where: {token_value: {_eq: $token_value}, deleted: {_eq: false}}, limit: 1) {
    id
  }
}
`

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
	var tokenValue string

	t.Run("CreateAPIToken", func(t *testing.T) {
		result, err := setup.CallMCPTool("mythic_create_api_token", map[string]interface{}{
			"token_type": "User",
		})
		require.NoError(t, err)
		assert.NotNil(t, result)

		meta, ok := result["metadata"].(map[string]interface{})
		require.True(t, ok, "Expected metadata in create api token result")
		val, ok := meta["token_value"].(string)
		require.True(t, ok && val != "", "Expected metadata.token_value to be a non-empty string")
		tokenValue = val

		// Resolve the token's database ID so we can delete it.
		resp, err := setup.MythicClient.ExecuteRawGraphQL(setup.Ctx, e2eFindAPITokenIDQuery, map[string]interface{}{
			"token_value": tokenValue,
		})
		require.NoError(t, err)
		if errs, ok := resp["errors"]; ok {
			t.Fatalf("GraphQL errors while resolving apitoken id: %v", errs)
		}

		// ExecuteRawGraphQL may return either the raw GraphQL envelope ({data:{...}})
		// or, depending on upstream behavior, a flattened object. Handle both.
		dataAny := resp["data"]
		if dataAny == nil {
			dataAny = resp
		}
		data, ok := dataAny.(map[string]interface{})
		require.True(t, ok, "Expected data object in GraphQL response")

		rows, ok := data["apitokens"].([]interface{})
		require.True(t, ok, "Expected apitokens array in GraphQL response")
		require.NotEmpty(t, rows, "Expected to find created api token by token_value")
		row, ok := rows[0].(map[string]interface{})
		require.True(t, ok)
		idFloat, ok := row["id"].(float64)
		require.True(t, ok)
		tokenID = int(idFloat)
		require.NotZero(t, tokenID, "Expected to resolve token_id for created token")
	})

	t.Run("DeleteAPIToken", func(t *testing.T) {
		require.NotZero(t, tokenID, "Expected token_id to be set by CreateAPIToken")
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
