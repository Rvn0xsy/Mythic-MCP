package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerPayloadDiscoveryTools registers tools for discovering payload build
// parameters, C2 profile parameters, and payload type commands. These enable
// agentic workflows to autonomously create payloads without prior knowledge.
func (s *Server) registerPayloadDiscoveryTools() {
	// mythic_get_payload_type_build_parameters - Discover build parameters for a payload type
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "mythic_get_payload_type_build_parameters",
		Description: "Get the build parameter schema for a payload type. Returns all configurable " +
			"build parameters including name, type, required, default value, description, and " +
			"available choices. Use this before creating a payload to discover what build_parameters are needed.",
	}, s.handleGetPayloadTypeBuildParameters)

	// mythic_get_c2_profile_parameters - Discover C2 profile configuration parameters
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "mythic_get_c2_profile_parameters",
		Description: "Get the parameter schema for a C2 profile (e.g. callback_host, callback_port, etc). " +
			"Returns all configurable parameters with name, type, required, default value, and description. " +
			"Call this before mythic_create_payload to learn what to pass in the c2_profiles[].parameters field. " +
			"Also ensure the C2 profile is running (started) before deploying the payload.",
	}, s.handleGetC2ProfileParameters)

	// mythic_get_payload_type_commands - Discover available commands for a payload type
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name: "mythic_get_payload_type_commands",
		Description: "Get all available commands for a specific payload type. Returns command names, " +
			"descriptions, and help text. Use this to discover what commands can be included when " +
			"creating a payload.",
	}, s.handleGetPayloadTypeCommands)
}

// Tool argument types

type getPayloadTypeBuildParametersArgs struct {
	PayloadTypeID int `json:"payload_type_id" jsonschema:"ID of the payload type (use mythic_get_payload_types to find IDs)"`
}

type getC2ProfileParametersArgs struct {
	C2ProfileID int `json:"c2_profile_id" jsonschema:"ID of the C2 profile (use mythic_get_c2_profiles to find IDs)"`
}

type getPayloadTypeCommandsArgs struct {
	PayloadTypeID int `json:"payload_type_id" jsonschema:"ID of the payload type (use mythic_get_payload_types to find IDs)"`
}

// Tool handlers

// handleGetPayloadTypeBuildParameters returns the build parameter schema for a payload type.
func (s *Server) handleGetPayloadTypeBuildParameters(ctx context.Context, req *mcp.CallToolRequest, args getPayloadTypeBuildParametersArgs) (*mcp.CallToolResult, any, error) {
	params, err := s.mythicClient.GetBuildParametersByPayloadType(ctx, args.PayloadTypeID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(params, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Build parameters for payload type %d (%d parameters):\n\n%s",
					args.PayloadTypeID, len(params), string(data)),
			},
		},
	}, wrapList(params), nil
}

// handleGetC2ProfileParameters returns the configuration parameter schema for a C2 profile.
func (s *Server) handleGetC2ProfileParameters(ctx context.Context, req *mcp.CallToolRequest, args getC2ProfileParametersArgs) (*mcp.CallToolResult, any, error) {
	params, err := s.mythicClient.GetC2ProfileParameters(ctx, args.C2ProfileID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(params, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("C2 profile parameters for profile %d (%d parameters):\n\n%s",
					args.C2ProfileID, len(params), string(data)),
			},
		},
	}, wrapList(params), nil
}

// handleGetPayloadTypeCommands returns all available commands for a payload type.
func (s *Server) handleGetPayloadTypeCommands(ctx context.Context, req *mcp.CallToolRequest, args getPayloadTypeCommandsArgs) (*mcp.CallToolResult, any, error) {
	commands, err := s.mythicClient.GetCommandsByPayloadType(ctx, args.PayloadTypeID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(commands, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Commands for payload type %d (%d commands):\n\n%s",
					args.PayloadTypeID, len(commands), string(data)),
			},
		},
	}, wrapList(commands), nil
}
