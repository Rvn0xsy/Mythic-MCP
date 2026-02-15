package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerProcessesTools registers process enumeration MCP tools
func (s *Server) registerProcessesTools() {
	// mythic_get_processes - List all processes
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_processes",
		Description: "Get a list of all processes enumerated in Mythic",
	}, s.handleGetProcesses)

	// mythic_get_processes_by_operation - List processes by operation
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_processes_by_operation",
		Description: "Get all processes enumerated in a specific operation",
	}, s.handleGetProcessesByOperation)

	// mythic_get_processes_by_callback - List processes by callback
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_processes_by_callback",
		Description: "Get all processes enumerated by a specific callback",
	}, s.handleGetProcessesByCallback)

	// mythic_get_process_tree - Get process tree
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_process_tree",
		Description: "Get process tree structure for a callback showing parent-child relationships",
	}, s.handleGetProcessTree)

	// mythic_get_processes_by_host - List processes by host
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_processes_by_host",
		Description: "Get all processes enumerated on a specific host",
	}, s.handleGetProcessesByHost)
}

// Tool argument types for process tools

type getProcessesArgs struct{}

type getProcessesByOperationArgs struct {
	OperationID int `json:"operation_id" jsonschema:"ID of the operation"`
}

type getProcessesByCallbackArgs struct {
	CallbackID int `json:"callback_id" jsonschema:"Callback display_id (the number shown in the Mythic UI, not the internal database id)"`
}

type getProcessTreeArgs struct {
	CallbackID int `json:"callback_id" jsonschema:"Callback display_id (the number shown in the Mythic UI, not the internal database id)"`
}

type getProcessesByHostArgs struct {
	HostID int `json:"host_id" jsonschema:"ID of the host"`
}

// Tool handlers

// handleGetProcesses retrieves all processes
func (s *Server) handleGetProcesses(ctx context.Context, req *mcp.CallToolRequest, args getProcessesArgs) (*mcp.CallToolResult, any, error) {
	processes, err := s.mythicClient.GetProcesses(ctx)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(processes, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("All processes (%d total):\n\n%s", len(processes), string(data)),
			},
		},
	}, wrapList(processes), nil
}

// handleGetProcessesByOperation retrieves processes for an operation
func (s *Server) handleGetProcessesByOperation(ctx context.Context, req *mcp.CallToolRequest, args getProcessesByOperationArgs) (*mcp.CallToolResult, any, error) {
	processes, err := s.mythicClient.GetProcessesByOperation(ctx, args.OperationID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(processes, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	// Build summary by host
	hostMap := make(map[string]int)
	for _, process := range processes {
		if process.Host != nil {
			hostMap[process.Host.Host]++
		}
	}

	summary := fmt.Sprintf("Processes in operation %d (%d total):\n\n", args.OperationID, len(processes))
	summary += "Breakdown by host:\n"
	for host, count := range hostMap {
		summary += fmt.Sprintf("  - %s: %d process(es)\n", host, count)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("%s\nFull details:\n\n%s", summary, string(data)),
			},
		},
	}, wrapList(processes), nil
}

// handleGetProcessesByCallback retrieves processes for a callback
func (s *Server) handleGetProcessesByCallback(ctx context.Context, req *mcp.CallToolRequest, args getProcessesByCallbackArgs) (*mcp.CallToolResult, any, error) {
	processes, err := s.mythicClient.GetProcessesByCallback(ctx, args.CallbackID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(processes, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Processes enumerated by callback %d (%d total):\n\n%s",
					args.CallbackID, len(processes), string(data)),
			},
		},
	}, wrapList(processes), nil
}

// handleGetProcessTree retrieves process tree for a callback
func (s *Server) handleGetProcessTree(ctx context.Context, req *mcp.CallToolRequest, args getProcessTreeArgs) (*mcp.CallToolResult, any, error) {
	processTree, err := s.mythicClient.GetProcessTree(ctx, args.CallbackID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(processTree, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Process tree for callback %d (%d nodes):\n\n%s",
					args.CallbackID, len(processTree), string(data)),
			},
		},
	}, wrapList(processTree), nil
}

// handleGetProcessesByHost retrieves processes for a host
func (s *Server) handleGetProcessesByHost(ctx context.Context, req *mcp.CallToolRequest, args getProcessesByHostArgs) (*mcp.CallToolResult, any, error) {
	processes, err := s.mythicClient.GetProcessesByHost(ctx, args.HostID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(processes, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Processes on host %d (%d total):\n\n%s",
					args.HostID, len(processes), string(data)),
			},
		},
	}, wrapList(processes), nil
}
