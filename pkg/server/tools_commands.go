package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerCommandsTools registers command query MCP tools
func (s *Server) registerCommandsTools() {
	// mythic_get_commands - List all commands
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_commands",
		Description: "Get a list of all commands available in Mythic",
	}, s.handleGetCommands)

	// mythic_get_command_parameters - Get all command parameters
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_command_parameters",
		Description: "Get a list of all command parameters across all commands",
	}, s.handleGetCommandParameters)

	// mythic_get_command_with_parameters - Get command with parameters
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_command_with_parameters",
		Description: "Get details of a specific command including its parameters and helper methods",
	}, s.handleGetCommandWithParameters)

	// Note: mythic_get_loaded_commands is already implemented in tools_callbacks.go
	// as it's specific to callback operations (GetLoadedCommands(ctx, callbackID))
}

// Tool argument types for command tools

type getCommandsArgs struct{}

type getCommandParametersArgs struct{}

type getCommandWithParametersArgs struct {
	PayloadTypeID int    `json:"payload_type_id" jsonschema:"ID of the payload type"`
	CommandName   string `json:"command_name" jsonschema:"Name of the command"`
}

// Tool handlers

// handleGetCommands retrieves all commands
func (s *Server) handleGetCommands(ctx context.Context, req *mcp.CallToolRequest, args getCommandsArgs) (*mcp.CallToolResult, any, error) {
	commands, err := s.mythicClient.GetCommands(ctx)
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
				Text: fmt.Sprintf("All commands (%d total):\n\n%s", len(commands), string(data)),
			},
		},
	}, wrapList(commands), nil
}

// handleGetCommandParameters retrieves all command parameters
func (s *Server) handleGetCommandParameters(ctx context.Context, req *mcp.CallToolRequest, args getCommandParametersArgs) (*mcp.CallToolResult, any, error) {
	parameters, err := s.mythicClient.GetCommandParameters(ctx)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(parameters, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("All command parameters (%d total):\n\n%s", len(parameters), string(data)),
			},
		},
	}, wrapList(parameters), nil
}

// handleGetCommandWithParameters retrieves a command with its parameters
func (s *Server) handleGetCommandWithParameters(ctx context.Context, req *mcp.CallToolRequest, args getCommandWithParametersArgs) (*mcp.CallToolResult, any, error) {
	cmdWithParams, err := s.mythicClient.GetCommandWithParameters(ctx, args.PayloadTypeID, args.CommandName)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(cmdWithParams, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	// Build summary information
	summary := fmt.Sprintf("Command: %s\n", cmdWithParams.Command.Cmd)
	summary += fmt.Sprintf("Version: %d\n", cmdWithParams.Command.Version)
	summary += fmt.Sprintf("Description: %s\n", cmdWithParams.Command.Description)
	summary += fmt.Sprintf("Parameters: %d\n", len(cmdWithParams.Parameters))

	if cmdWithParams.IsRawStringCommand() {
		summary += "Type: Raw string command\n"
	}

	if cmdWithParams.HasRequiredParameters() {
		summary += "Has required parameters: yes\n"
	} else {
		summary += "Has required parameters: no\n"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("%s\nFull details:\n\n%s", summary, string(data)),
			},
		},
	}, cmdWithParams, nil
}
