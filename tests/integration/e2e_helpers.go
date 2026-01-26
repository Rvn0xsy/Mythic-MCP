//go:build integration && e2e
// +build integration,e2e

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nbaertsch/Mythic-MCP/pkg/config"
	"github.com/nbaertsch/Mythic-MCP/pkg/server"
	"github.com/nbaertsch/mythic-sdk-go/pkg/mythic"
	"github.com/stretchr/testify/require"
)

// MCPTestSetup provides E2E test infrastructure
type MCPTestSetup struct {
	T      *testing.T
	Ctx    context.Context
	Cancel context.CancelFunc

	// MCP server components
	MCPServer    *server.Server
	MCPTransport *testTransport

	// Mythic SDK client (for verification)
	MythicClient *mythic.Client

	// Cleanup functions
	cleanupFuncs []func()
}

// testTransport is an in-memory transport for testing (MCP SDK v1.2.0 compatible)
type testTransport struct {
	conn *testConnection
}

// testConnection implements mcp.Connection for testing
type testConnection struct {
	serverToClient chan jsonrpc.Message
	clientToServer chan jsonrpc.Message
	closed         bool
}

// newTestTransport creates a new test transport
func newTestTransport() *testTransport {
	return &testTransport{
		conn: &testConnection{
			serverToClient: make(chan jsonrpc.Message, 10),
			clientToServer: make(chan jsonrpc.Message, 10),
		},
	}
}

// Connect implements mcp.Transport interface (v1.2.0)
func (t *testTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	return t.conn, nil
}

// Read implements mcp.Connection interface
func (c *testConnection) Read(ctx context.Context) (jsonrpc.Message, error) {
	if c.closed {
		return nil, fmt.Errorf("connection closed")
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg, ok := <-c.clientToServer:
		if !ok {
			return nil, fmt.Errorf("connection closed")
		}
		return msg, nil
	}
}

// Write implements mcp.Connection interface
func (c *testConnection) Write(ctx context.Context, msg jsonrpc.Message) error {
	if c.closed {
		return fmt.Errorf("connection closed")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.serverToClient <- msg:
		return nil
	}
}

// Close implements mcp.Connection interface
func (c *testConnection) Close() error {
	if !c.closed {
		c.closed = true
		close(c.serverToClient)
		close(c.clientToServer)
	}
	return nil
}

// SessionID implements mcp.Connection interface
func (c *testConnection) SessionID() string {
	return "test-session"
}

// Helper methods for testing

// SendMessage sends a message from client to server (for testing)
func (t *testTransport) SendMessage(msg jsonrpc.Message) {
	t.conn.clientToServer <- msg
}

// ReceiveMessage receives a message from server to client (for testing)
func (t *testTransport) ReceiveMessage() (jsonrpc.Message, bool) {
	msg, ok := <-t.conn.serverToClient
	return msg, ok
}

// Close closes the transport
func (t *testTransport) Close() error {
	return t.conn.Close()
}

// SetupE2ETest creates complete E2E test environment
func SetupE2ETest(t *testing.T) *MCPTestSetup {
	ctx, cancel := context.WithCancel(context.Background())

	// Get Mythic credentials from environment
	mythicURL := getEnvOrDefault("MYTHIC_URL", "https://127.0.0.1:7443")
	mythicPassword := os.Getenv("MYTHIC_PASSWORD")
	mythicUsername := getEnvOrDefault("MYTHIC_USERNAME", "mythic_admin")

	// Skip if credentials not available
	if mythicPassword == "" {
		t.Skip("MYTHIC_PASSWORD not set - skipping E2E test")
	}

	// Create configuration
	cfg := &config.Config{
		MythicURL:     mythicURL,
		Username:      mythicUsername,
		Password:      mythicPassword,
		SSL:           true,
		SkipTLSVerify: true,
		LogLevel:      "error", // Quiet during tests
		Timeout:       30 * time.Second,
	}

	// Create Mythic SDK client for verification
	mythicClient, err := mythic.NewClient(&mythic.Config{
		ServerURL:     cfg.MythicURL,
		Username:      cfg.Username,
		Password:      cfg.Password,
		SSL:           cfg.SSL,
		SkipTLSVerify: cfg.SkipTLSVerify,
		Timeout:       cfg.Timeout,
	})
	require.NoError(t, err, "Failed to create Mythic client")

	// Authenticate Mythic client for verification
	err = mythicClient.Login(ctx)
	require.NoError(t, err, "Failed to authenticate with Mythic")

	// Create MCP server
	mcpServer, err := server.NewServer(cfg)
	require.NoError(t, err, "Failed to create MCP server")

	// Create test transport
	transport := newTestTransport()

	// Start MCP server in background
	go func() {
		_ = mcpServer.Run(ctx, transport)
	}()

	// Wait a moment for server to initialize
	time.Sleep(100 * time.Millisecond)

	setup := &MCPTestSetup{
		T:            t,
		Ctx:          ctx,
		Cancel:       cancel,
		MCPServer:    mcpServer,
		MCPTransport: transport,
		MythicClient: mythicClient,
		cleanupFuncs: []func(){},
	}

	// Register cleanup
	t.Cleanup(setup.Cleanup)

	return setup
}

// Cleanup runs all registered cleanup functions
func (s *MCPTestSetup) Cleanup() {
	// Cancel context
	if s.Cancel != nil {
		s.Cancel()
	}

	// Run cleanup functions in reverse order
	for i := len(s.cleanupFuncs) - 1; i >= 0; i-- {
		s.cleanupFuncs[i]()
	}

	// Close clients
	if s.MythicClient != nil {
		s.MythicClient.Close()
	}

	if s.MCPServer != nil {
		s.MCPServer.Close()
	}

	if s.MCPTransport != nil {
		s.MCPTransport.Close()
	}
}

// CallMCPTool executes an MCP tool and returns result
func (s *MCPTestSetup) CallMCPTool(toolName string, args map[string]interface{}) (map[string]interface{}, error) {
	// Create tool call request with unique ID
	requestID := time.Now().UnixNano() // Use timestamp for unique ID
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      requestID,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      toolName,
			"arguments": args,
		},
	}

	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Decode to JSON-RPC message
	requestMsg, err := jsonrpc.DecodeMessage(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode request message: %w", err)
	}

	// Send request
	s.MCPTransport.SendMessage(requestMsg)

	// Wait for response with matching ID (loop to skip notifications)
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for MCP response")
		default:
			responseMsg, ok := s.MCPTransport.ReceiveMessage()
			if !ok {
				return nil, fmt.Errorf("transport closed")
			}

			// Check message type - skip requests (notifications)
			if req, isRequest := responseMsg.(*jsonrpc.Request); isRequest {
				if os.Getenv("E2E_DEBUG") == "1" {
					fmt.Printf("DEBUG: Skipping notification: %s\n", req.Method)
				}
				continue // Skip notifications, wait for actual response
			}

			// Must be a response at this point
			resp, isResponse := responseMsg.(*jsonrpc.Response)
			if !isResponse {
				if os.Getenv("E2E_DEBUG") == "1" {
					fmt.Printf("DEBUG: Unknown message type: %T\n", responseMsg)
				}
				continue
			}

			if os.Getenv("E2E_DEBUG") == "1" {
				fmt.Printf("DEBUG: Resp ID: %v (type %T), Request ID: %d\n", resp.ID.Raw(), resp.ID.Raw(), requestID)
			}
			// Check if this response matches our request ID using ID.Raw()
			if !resp.ID.IsValid() {
				if os.Getenv("E2E_DEBUG") == "1" {
					fmt.Printf("DEBUG: Response has invalid/nil ID\n")
				}
				continue
			}

			responseID, ok := resp.ID.Raw().(int64)
			if !ok {
				// Try float64 (from JSON unmarshaling)
				if f, isFloat := resp.ID.Raw().(float64); isFloat {
					responseID = int64(f)
				} else {
					if os.Getenv("E2E_DEBUG") == "1" {
						fmt.Printf("DEBUG: Response ID is not int64: %T = %v\n", resp.ID.Raw(), resp.ID.Raw())
					}
					continue
				}
			}

			if responseID != requestID {
				if os.Getenv("E2E_DEBUG") == "1" {
					fmt.Printf("DEBUG: Skipping response with mismatched ID: %v != %v\n", responseID, requestID)
				}
				continue // Not our response, keep waiting
			}

			// This is our response!
			if os.Getenv("E2E_DEBUG") == "1" {
				fmt.Printf("DEBUG: Found matching response for request ID %d\n", requestID)
			}

			// Check for error in response
			if resp.Error != nil {
				return nil, fmt.Errorf("MCP error: %v", resp.Error)
			}

			// Extract result from Response
			if resp.Result == nil {
				// No result - return empty map
				return map[string]interface{}{}, nil
			}

			// Unmarshal the Result (json.RawMessage) into a map
			var result map[string]interface{}
			if err := json.Unmarshal(resp.Result, &result); err != nil {
				// Result is not a map - try to handle as raw value
				if os.Getenv("E2E_DEBUG") == "1" {
					fmt.Printf("DEBUG: Result is not a JSON object: %s\n", string(resp.Result))
				}
				return map[string]interface{}{"result": string(resp.Result)}, nil
			}

			// Extract structured content (domain objects) from MCP response
			// The MCP SDK puts the second return value in "structuredContent" field
			var normalizedResult = make(map[string]interface{})

			// Copy structuredContent as both "metadata" and "content" for test compatibility
			// This field contains the actual domain objects from our tool handlers
			if structuredContent, ok := result["structuredContent"]; ok {
				normalizedResult["metadata"] = structuredContent
				normalizedResult["content"] = structuredContent
			}

			// Also copy MCP Content array (text content) as mcp_content for reference
			if content, ok := result["content"]; ok {
				normalizedResult["mcp_content"] = content

				// If we don't have structured content, use content field
				if _, hasStructured := normalizedResult["content"]; !hasStructured {
					normalizedResult["content"] = content
				}
			}

			// Copy isError field
			if isError, ok := result["isError"]; ok {
				normalizedResult["isError"] = isError
			}

			// Copy _meta if present (for backward compatibility)
			if meta, ok := result["_meta"]; ok {
				normalizedResult["_meta"] = meta
			}

			// If normalized result is empty, return original result
			if len(normalizedResult) == 0 {
				if os.Getenv("E2E_DEBUG") == "1" {
					fmt.Printf("DEBUG: Normalized result is empty, returning original result\n\n")
				}
				return result, nil
			}

			// Debug: Print normalized result
			if os.Getenv("E2E_DEBUG") == "1" {
				normalizedData, _ := json.MarshalIndent(normalizedResult, "", "  ")
				fmt.Printf("DEBUG: Normalized result for %s:\n%s\n", toolName, string(normalizedData))
				if content, ok := normalizedResult["content"]; ok {
					fmt.Printf("DEBUG: content field type: %T\n", content)
				}
				fmt.Printf("=================================================\n\n")
			}

			return normalizedResult, nil
		}
	}
}

// getEnvOrDefault returns environment variable or default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper functions for parsing responses

// parseOperatorResult parses an operator from MCP response
func parseOperatorResult(response map[string]interface{}) *mythic.Operator {
	// This is a placeholder - actual implementation will parse the MCP result
	// For now, return nil
	return nil
}

// parseTokenResult parses a token from MCP response
func parseTokenResult(response map[string]interface{}) *APIToken {
	// Placeholder
	return nil
}

// APIToken represents an API token (simplified for tests)
type APIToken struct {
	ID    int
	Value string
}
