//go:build integration && e2e
// +build integration,e2e

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nbaertsch/Mythic-MCP/pkg/config"
	"github.com/nbaertsch/Mythic-MCP/pkg/server"
	"github.com/nbaertsch/mythic-sdk-go/pkg/mythic"
	"github.com/nbaertsch/mythic-sdk-go/pkg/mythic/types"
	"github.com/stretchr/testify/require"
)

// Global request ID counter
var requestCounter int64

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

	// E2E tests require real Mythic credentials.
	require.NotEmpty(t, mythicPassword, "MYTHIC_PASSWORD not set - E2E tests require a running Mythic instance")

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

	// Perform MCP initialization handshake
	if err := setup.initializeMCPSession(); err != nil {
		require.NoError(t, err, "Failed to initialize MCP session")
	}

	return setup
}

func e2eStrictMode() bool {
	return os.Getenv("E2E_STRICT") == "1"
}

func requireCurrentOperationIDOrReturn(t *testing.T, setup *MCPTestSetup) (int, bool) {
	me, err := setup.MythicClient.GetMe(setup.Ctx)
	require.NoError(t, err)
	if me.CurrentOperation == nil {
		if e2eStrictMode() {
			require.FailNow(t, "No current operation set")
		}
		t.Logf("No current operation set; exercising negative/empty-path behavior")
		return 0, false
	}
	return me.CurrentOperation.ID, true
}

func requireCallbacksOrReturn(t *testing.T, setup *MCPTestSetup, min int) ([]*types.Callback, bool) {
	cbs, err := setup.MythicClient.GetAllCallbacks(setup.Ctx)
	require.NoError(t, err)
	if len(cbs) < min {
		if e2eStrictMode() {
			require.FailNow(t, "Not enough callbacks", "need=%d have=%d", min, len(cbs))
		}
		t.Logf("Not enough callbacks available (need=%d have=%d); exercising negative/empty-path behavior", min, len(cbs))
		return nil, false
	}
	return cbs, true
}

func requireActiveCallbacksOrReturn(t *testing.T, setup *MCPTestSetup, min int) ([]*types.Callback, bool) {
	cbs, err := setup.MythicClient.GetAllActiveCallbacks(setup.Ctx)
	require.NoError(t, err)
	if len(cbs) < min {
		if e2eStrictMode() {
			require.FailNow(t, "Not enough active callbacks", "need=%d have=%d", min, len(cbs))
		}
		t.Logf("Not enough active callbacks available (need=%d have=%d); exercising negative/empty-path behavior", min, len(cbs))
		return nil, false
	}
	return cbs, true
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
	// Create unique request ID using atomic counter
	requestID := atomic.AddInt64(&requestCounter, 1)

	// Create tool call request
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

	// Expected ID is what we sent
	expectedID := requestID

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
				fmt.Printf("DEBUG: Resp ID: %v (type %T), Expected ID: %d\\n", resp.ID.Raw(), resp.ID.Raw(), expectedID)
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

			if responseID != expectedID {
				if os.Getenv("E2E_DEBUG") == "1" {
					fmt.Printf("DEBUG: Skipping response with mismatched ID: %v != %v\\n", responseID, expectedID)
				}
				continue // Not our response, keep waiting
			}

			// This is our response!
			if os.Getenv("E2E_DEBUG") == "1" {
				fmt.Printf("DEBUG: Found matching response for expected ID %d\\n", expectedID)
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

			// Copy structuredContent as "metadata" and normalize "content" for test compatibility.
			// For list envelopes produced by wrapList ("items" + "count"), expose content
			// as the underlying items array because many tests expect []interface{}.
			if structuredContent, ok := result["structuredContent"]; ok {
				normalizedResult["metadata"] = structuredContent

				if scMap, isMap := structuredContent.(map[string]interface{}); isMap {
					if items, hasItems := scMap["items"]; hasItems {
						normalizedResult["content"] = items
					} else {
						normalizedResult["content"] = structuredContent
					}
				} else {
					normalizedResult["content"] = structuredContent
				}
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

			// Surface MCP tool-level errors as Go errors so tests can assert on err.
			if isErrorRaw, ok := normalizedResult["isError"]; ok {
				if isError, ok := isErrorRaw.(bool); ok && isError {
					if content, ok := normalizedResult["content"].([]interface{}); ok && len(content) > 0 {
						if first, ok := content[0].(map[string]interface{}); ok {
							if text, ok := first["text"].(string); ok && text != "" {
								return nil, fmt.Errorf("%s", text)
							}
						}
					}

				if mcpContent, ok := normalizedResult["mcp_content"].([]interface{}); ok && len(mcpContent) > 0 {
					if first, ok := mcpContent[0].(map[string]interface{}); ok {
						if text, ok := first["text"].(string); ok && text != "" {
							return nil, fmt.Errorf("%s", text)
						}
					}
				}

				return nil, fmt.Errorf("MCP tool returned isError=true")
			}
			if isError, ok := isErrorRaw.(float64); ok && isError != 0 {
				return nil, fmt.Errorf("MCP tool returned isError=true")
			}
			if isError, ok := isErrorRaw.(string); ok && isError == "true" {
				return nil, fmt.Errorf("MCP tool returned isError=true")
			}
			if isError, ok := isErrorRaw.(string); ok && isError == "1" {
				return nil, fmt.Errorf("MCP tool returned isError=true")
			}
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

// initializeMCPSession performs the MCP initialization handshake
func (s *MCPTestSetup) initializeMCPSession() error {
	// Send initialize request
	requestID := atomic.AddInt64(&requestCounter, 1)
	initRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      requestID,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "E2ETestClient",
				"version": "1.0.0",
			},
		},
	}

	requestData, err := json.Marshal(initRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal initialize request: %w", err)
	}

	requestMsg, err := jsonrpc.DecodeMessage(requestData)
	if err != nil {
		return fmt.Errorf("failed to decode initialize request: %w", err)
	}

	s.MCPTransport.SendMessage(requestMsg)

	// Wait for initialize response
	timeout := time.After(5 * time.Second)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for initialize response")
		default:
			responseMsg, ok := s.MCPTransport.ReceiveMessage()
			if !ok {
				return fmt.Errorf("transport closed during initialization")
			}

			// Skip notifications
			if req, isRequest := responseMsg.(*jsonrpc.Request); isRequest {
				if req.Method == "notifications/tools/list_changed" {
					continue
				}
			}

			// Check for response
			resp, isResponse := responseMsg.(*jsonrpc.Response)
			if !isResponse {
				continue
			}

			// Check if it's our initialize response
			if resp.ID.IsValid() {
				respID, ok := resp.ID.Raw().(int64)
				if !ok {
					if f, isFloat := resp.ID.Raw().(float64); isFloat {
						respID = int64(f)
					} else {
						continue
					}
				}

				if respID == requestID {
					// Got initialize response
					if resp.Error != nil {
						return fmt.Errorf("initialize failed: %v", resp.Error)
					}
					goto initializeDone
				}
			}
		}
	}

initializeDone:
	// Send initialized notification
	notificationRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}

	notifData, err := json.Marshal(notificationRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal initialized notification: %w", err)
	}

	notifMsg, err := jsonrpc.DecodeMessage(notifData)
	if err != nil {
		return fmt.Errorf("failed to decode initialized notification: %w", err)
	}

	s.MCPTransport.SendMessage(notifMsg)

	// Wait a moment for notification to be processed
	time.Sleep(50 * time.Millisecond)

	return nil
}
