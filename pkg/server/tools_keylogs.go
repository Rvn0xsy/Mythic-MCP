package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerKeylogsTools registers keylogger data MCP tools
func (s *Server) registerKeylogsTools() {
	// mythic_get_keylogs - List all keylogs
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_keylogs",
		Description: "Get all keylogger data captured across all operations",
	}, s.handleGetKeylogs)

	// mythic_get_keylogs_by_operation - List keylogs by operation
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_keylogs_by_operation",
		Description: "Get all keylogger data captured in a specific operation",
	}, s.handleGetKeylogsByOperation)

	// mythic_get_keylogs_by_callback - List keylogs by callback
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_keylogs_by_callback",
		Description: "Get all keylogger data captured by a specific callback",
	}, s.handleGetKeylogsByCallback)
}

// Tool argument types for keylog tools

type getKeylogsArgs struct{}

type getKeylogsByOperationArgs struct {
	OperationID int `json:"operation_id" jsonschema:"ID of the operation"`
}

type getKeylogsByCallbackArgs struct {
	CallbackID int `json:"callback_id" jsonschema:"Display ID of the callback"`
}

// Tool handlers

// handleGetKeylogs retrieves all keylogs
func (s *Server) handleGetKeylogs(ctx context.Context, req *mcp.CallToolRequest, args getKeylogsArgs) (*mcp.CallToolResult, any, error) {
	keylogs, err := s.mythicClient.GetKeylogs(ctx)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(keylogs, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	// Calculate total keystrokes
	totalKeystrokes := 0
	for _, keylog := range keylogs {
		totalKeystrokes += len(keylog.Keystrokes)
	}

	summary := fmt.Sprintf("All keylogs (%d entries, ~%d keystrokes total):\n\n",
		len(keylogs), totalKeystrokes)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("%s%s", summary, string(data)),
			},
		},
	}, keylogs, nil
}

// handleGetKeylogsByOperation retrieves keylogs for an operation
func (s *Server) handleGetKeylogsByOperation(ctx context.Context, req *mcp.CallToolRequest, args getKeylogsByOperationArgs) (*mcp.CallToolResult, any, error) {
	keylogs, err := s.mythicClient.GetKeylogsByOperation(ctx, args.OperationID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(keylogs, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	// Build summary by user
	userMap := make(map[string]int)
	totalKeystrokes := 0
	for _, keylog := range keylogs {
		userMap[keylog.User]++
		totalKeystrokes += len(keylog.Keystrokes)
	}

	summary := fmt.Sprintf("Keylogs in operation %d (%d entries, ~%d keystrokes):\n\n",
		args.OperationID, len(keylogs), totalKeystrokes)

	if len(userMap) > 0 {
		summary += "Breakdown by user:\n"
		for user, count := range userMap {
			summary += fmt.Sprintf("  - %s: %d keylog(s)\n", user, count)
		}
		summary += "\n"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("%sFull details:\n\n%s", summary, string(data)),
			},
		},
	}, keylogs, nil
}

// handleGetKeylogsByCallback retrieves keylogs for a callback
func (s *Server) handleGetKeylogsByCallback(ctx context.Context, req *mcp.CallToolRequest, args getKeylogsByCallbackArgs) (*mcp.CallToolResult, any, error) {
	keylogs, err := s.mythicClient.GetKeylogsByCallback(ctx, args.CallbackID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(keylogs, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	// Build summary by window
	windowMap := make(map[string]int)
	totalKeystrokes := 0
	for _, keylog := range keylogs {
		windowMap[keylog.Window]++
		totalKeystrokes += len(keylog.Keystrokes)
	}

	summary := fmt.Sprintf("Keylogs for callback %d (%d entries, ~%d keystrokes):\n\n",
		args.CallbackID, len(keylogs), totalKeystrokes)

	if len(windowMap) > 0 {
		summary += "Breakdown by window:\n"
		for window, count := range windowMap {
			if window == "" {
				window = "(unknown)"
			}
			summary += fmt.Sprintf("  - %s: %d keylog(s)\n", window, count)
		}
		summary += "\n"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("%sFull details:\n\n%s", summary, string(data)),
			},
		},
	}, keylogs, nil
}
