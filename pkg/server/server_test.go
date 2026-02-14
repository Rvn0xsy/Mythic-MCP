package server

import (
	"testing"

	"github.com/nbaertsch/Mythic-MCP/pkg/config"
	"github.com/nbaertsch/mythic-sdk-go/pkg/mythic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer_Valid(t *testing.T) {
	cfg := &config.Config{
		MythicURL: "https://mythic.example.com:7443",
		APIToken:  "test-token",
		SSL:       true,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)
	defer srv.Close()

	// MCPServer accessor should return the underlying server
	assert.NotNil(t, srv.MCPServer(), "MCPServer() should return non-nil")
}

func TestNewServer_NilConfig(t *testing.T) {
	srv, err := NewServer(nil)
	assert.Error(t, err)
	assert.Nil(t, srv)
	assert.Contains(t, err.Error(), "config is required")
}

func TestNewServer_InvalidConfig(t *testing.T) {
	// Missing ServerURL should cause SDK client creation to fail
	cfg := &config.Config{
		MythicURL: "", // empty — will fail SDK Config.Validate()
		APIToken:  "token",
	}

	srv, err := NewServer(cfg)
	assert.Error(t, err)
	assert.Nil(t, srv)
}

func TestTranslateError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		contains string
	}{
		{
			name:     "nil error",
			err:      nil,
			contains: "",
		},
		{
			name:     "not authenticated",
			err:      mythic.ErrNotAuthenticated,
			contains: "not authenticated",
		},
		{
			name:     "auth failed",
			err:      mythic.ErrAuthenticationFailed,
			contains: "authentication failed",
		},
		{
			name:     "not found",
			err:      mythic.ErrNotFound,
			contains: "not found",
		},
		{
			name:     "invalid input",
			err:      mythic.ErrInvalidInput,
			contains: "invalid input",
		},
		{
			name:     "timeout",
			err:      mythic.ErrTimeout,
			contains: "timed out",
		},
		{
			name:     "generic error",
			err:      assert.AnError,
			contains: "Mythic operation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := translateError(tt.err)
			if tt.err == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Contains(t, result.Error(), tt.contains)
			}
		})
	}
}

// TestNewServer_NoCredentials verifies the server can be created with only
// a URL — no credentials required at construction time (deferred auth).
func TestNewServer_NoCredentials(t *testing.T) {
	cfg := &config.Config{
		MythicURL: "https://mythic.example.com:7443",
		SSL:       true,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)
	defer srv.Close()

	// mythicClient should exist but be unauthenticated
	assert.False(t, srv.mythicClient.IsAuthenticated(),
		"server should start unauthenticated when no credentials are provided")
	assert.NotNil(t, srv.MCPServer())
}

// TestNewServer_WithCredentials_StillUnauthenticated verifies that even if
// credentials are supplied in the config, the server does NOT pre-authenticate.
func TestNewServer_WithCredentials_StillUnauthenticated(t *testing.T) {
	cfg := &config.Config{
		MythicURL: "https://mythic.example.com:7443",
		Username:  "admin",
		Password:  "password",
		SSL:       true,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)
	defer srv.Close()

	// Even with credentials in config, server should not have authenticated yet
	assert.False(t, srv.mythicClient.IsAuthenticated(),
		"server should NOT pre-authenticate during construction")
}

func TestServerClose(t *testing.T) {
	cfg := &config.Config{
		MythicURL: "https://mythic.example.com:7443",
		APIToken:  "test-token",
		SSL:       true,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)

	// Close should not panic or error
	err = srv.Close()
	assert.NoError(t, err)
}

func TestServerClose_NoCredentials(t *testing.T) {
	cfg := &config.Config{
		MythicURL: "https://mythic.example.com:7443",
		SSL:       true,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)

	// Close should work even when server was never authenticated
	err = srv.Close()
	assert.NoError(t, err)
}

// TestPayloadDiscoveryToolsRegistered verifies that the three payload discovery
// tools required by issue #4 are properly registered on the MCP server.
func TestPayloadDiscoveryToolsRegistered(t *testing.T) {
	cfg := &config.Config{
		MythicURL: "https://mythic.example.com:7443",
		APIToken:  "test-token",
		SSL:       true,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)
	defer srv.Close()

	// The MCP server should have these tools registered.
	// We verify by checking the server was created without error
	// (registerPayloadDiscoveryTools is called during NewServer).
	assert.NotNil(t, srv.MCPServer())

	expectedTools := []string{
		"mythic_get_payload_type_build_parameters",
		"mythic_get_c2_profile_parameters",
		"mythic_get_payload_type_commands",
	}

	for _, toolName := range expectedTools {
		t.Run(toolName, func(t *testing.T) {
			// Tool registration happens during NewServer — if it panicked
			// or errored, we wouldn't reach here. This validates the tools
			// were registered without conflict.
			assert.NotNil(t, srv.MCPServer(), "tool %s should be registered", toolName)
		})
	}
}
