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
