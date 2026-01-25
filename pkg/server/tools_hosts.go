package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerHostsTools registers host management MCP tools
func (s *Server) registerHostsTools() {
	// mythic_get_hosts - List all hosts in an operation
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_hosts",
		Description: "Get all hosts tracked in a specific operation",
	}, s.handleGetHosts)

	// mythic_get_host_by_id - Get host by ID
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_host_by_id",
		Description: "Get detailed information about a specific host by its ID",
	}, s.handleGetHostByID)

	// mythic_get_host_by_hostname - Get host by hostname
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_host_by_hostname",
		Description: "Get detailed information about a specific host by its hostname",
	}, s.handleGetHostByHostname)

	// mythic_get_host_network_map - Get network map
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_host_network_map",
		Description: "Get network topology map showing host relationships in an operation",
	}, s.handleGetHostNetworkMap)

	// mythic_get_callbacks_for_host - Get callbacks on a host
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_callbacks_for_host",
		Description: "Get all callbacks (agents) running on a specific host",
	}, s.handleGetCallbacksForHost)
}

// Tool argument types for host tools

type getHostsArgs struct {
	OperationID int `json:"operation_id" jsonschema:"required,description=ID of the operation"`
}

type getHostByIDArgs struct {
	HostID int `json:"host_id" jsonschema:"required,description=ID of the host"`
}

type getHostByHostnameArgs struct {
	Hostname string `json:"hostname" jsonschema:"required,description=Hostname to search for"`
}

type getHostNetworkMapArgs struct {
	OperationID int `json:"operation_id" jsonschema:"required,description=ID of the operation"`
}

type getCallbacksForHostArgs struct {
	HostID int `json:"host_id" jsonschema:"required,description=ID of the host"`
}

// Tool handlers

// handleGetHosts retrieves all hosts in an operation
func (s *Server) handleGetHosts(ctx context.Context, req *mcp.CallToolRequest, args getHostsArgs) (*mcp.CallToolResult, any, error) {
	hosts, err := s.mythicClient.GetHosts(ctx, args.OperationID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(hosts, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	// Build summary by OS
	osMap := make(map[string]int)
	for _, host := range hosts {
		osMap[host.OS]++
	}

	summary := fmt.Sprintf("Hosts in operation %d (%d total):\n\n", args.OperationID, len(hosts))
	summary += "Breakdown by OS:\n"
	for os, count := range osMap {
		summary += fmt.Sprintf("  - %s: %d host(s)\n", os, count)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("%s\nFull details:\n\n%s", summary, string(data)),
			},
		},
	}, hosts, nil
}

// handleGetHostByID retrieves a specific host by ID
func (s *Server) handleGetHostByID(ctx context.Context, req *mcp.CallToolRequest, args getHostByIDArgs) (*mcp.CallToolResult, any, error) {
	host, err := s.mythicClient.GetHostByID(ctx, args.HostID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(host, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Host %s (ID: %d):\nIP: %s\nOS: %s\nArchitecture: %s\n\n%s",
					host.Hostname, host.ID, host.IP, host.OS, host.Architecture, string(data)),
			},
		},
	}, host, nil
}

// handleGetHostByHostname retrieves a host by hostname
func (s *Server) handleGetHostByHostname(ctx context.Context, req *mcp.CallToolRequest, args getHostByHostnameArgs) (*mcp.CallToolResult, any, error) {
	host, err := s.mythicClient.GetHostByHostname(ctx, args.Hostname)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(host, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Host %s (ID: %d):\nIP: %s\nOS: %s\nArchitecture: %s\n\n%s",
					host.Hostname, host.ID, host.IP, host.OS, host.Architecture, string(data)),
			},
		},
	}, host, nil
}

// handleGetHostNetworkMap retrieves network topology for operation
func (s *Server) handleGetHostNetworkMap(ctx context.Context, req *mcp.CallToolRequest, args getHostNetworkMapArgs) (*mcp.CallToolResult, any, error) {
	networkMap, err := s.mythicClient.GetHostNetworkMap(ctx, args.OperationID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(networkMap, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Network map for operation %d:\n\n%s",
					args.OperationID, string(data)),
			},
		},
	}, networkMap, nil
}

// handleGetCallbacksForHost retrieves callbacks running on a host
func (s *Server) handleGetCallbacksForHost(ctx context.Context, req *mcp.CallToolRequest, args getCallbacksForHostArgs) (*mcp.CallToolResult, any, error) {
	callbacks, err := s.mythicClient.GetCallbacksForHost(ctx, args.HostID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(callbacks, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	// Build summary by status
	statusMap := make(map[string]int)
	for _, callback := range callbacks {
		status := "inactive"
		if callback.Active {
			status = "active"
		}
		statusMap[status]++
	}

	summary := fmt.Sprintf("Callbacks on host %d (%d total):\n\n", args.HostID, len(callbacks))
	summary += "Breakdown by status:\n"
	for status, count := range statusMap {
		summary += fmt.Sprintf("  - %s: %d callback(s)\n", status, count)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("%s\nFull details:\n\n%s", summary, string(data)),
			},
		},
	}, callbacks, nil
}
