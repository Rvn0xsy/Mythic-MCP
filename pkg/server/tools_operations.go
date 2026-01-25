package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nbaertsch/mythic-sdk-go/pkg/mythic/types"
)

// registerOperationsTools registers operation management MCP tools
func (s *Server) registerOperationsTools() {
	// mythic_get_operations - List all operations
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_operations",
		Description: "Get a list of all operations in the Mythic instance",
	}, s.handleGetOperations)

	// mythic_get_operation - Get specific operation by ID
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_operation",
		Description: "Get details of a specific operation by ID",
	}, s.handleGetOperation)

	// mythic_create_operation - Create new operation
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_create_operation",
		Description: "Create a new operation (campaign/engagement)",
	}, s.handleCreateOperation)

	// mythic_update_operation - Update existing operation
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_update_operation",
		Description: "Update an existing operation's properties",
	}, s.handleUpdateOperation)

	// mythic_set_current_operation - Set current operation context
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_set_current_operation",
		Description: "Set the current operation context for the client",
	}, s.handleSetCurrentOperation)

	// mythic_get_current_operation - Get current operation
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_current_operation",
		Description: "Get the currently active operation context",
	}, s.handleGetCurrentOperation)

	// mythic_get_operation_operators - List operators in operation
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_operation_operators",
		Description: "Get list of operators (users) in a specific operation",
	}, s.handleGetOperationOperators)

	// mythic_create_event_log - Create event log entry
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_create_event_log",
		Description: "Create an event log entry for an operation",
	}, s.handleCreateEventLog)

	// mythic_get_event_log - Get event logs
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_event_log",
		Description: "Get event log entries for an operation",
	}, s.handleGetEventLog)

	// mythic_get_global_settings - Get global Mythic settings
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_global_settings",
		Description: "Get global Mythic server settings",
	}, s.handleGetGlobalSettings)

	// mythic_update_global_settings - Update global settings
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_update_global_settings",
		Description: "Update global Mythic server settings",
	}, s.handleUpdateGlobalSettings)
}

// Tool argument types for operations tools

type getOperationsArgs struct{}

type getOperationArgs struct {
	OperationID int `json:"operation_id" jsonschema:"ID of the operation to retrieve"`
}

type createOperationArgs struct {
	Name    string  `json:"name" jsonschema:"Name of the new operation"`
	Webhook *string `json:"webhook,omitempty" jsonschema:"Webhook URL for notifications"`
	Channel *string `json:"channel,omitempty" jsonschema:"Slack/Discord channel for notifications"`
	AdminID *int    `json:"admin_id,omitempty" jsonschema:"Operator ID to set as admin"`
}

type updateOperationArgs struct {
	OperationID int     `json:"operation_id" jsonschema:"ID of the operation to update"`
	Name        *string `json:"name,omitempty" jsonschema:"New name for the operation"`
	Webhook     *string `json:"webhook,omitempty" jsonschema:"Webhook URL for notifications"`
	Channel     *string `json:"channel,omitempty" jsonschema:"Slack/Discord channel"`
	Complete    *bool   `json:"complete,omitempty" jsonschema:"Mark operation as complete"`
	AdminID     *int    `json:"admin_id,omitempty" jsonschema:"New admin operator ID"`
	BannerText  *string `json:"banner_text,omitempty" jsonschema:"Banner text for operation"`
	BannerColor *string `json:"banner_color,omitempty" jsonschema:"Banner color (hex code)"`
}

type setCurrentOperationArgs struct {
	OperationID int `json:"operation_id" jsonschema:"ID of the operation to set as current"`
}

type getCurrentOperationArgs struct{}

type getOperationOperatorsArgs struct {
	OperationID int `json:"operation_id" jsonschema:"ID of the operation"`
}

type createEventLogArgs struct {
	OperationID int     `json:"operation_id" jsonschema:"ID of the operation"`
	Message     string  `json:"message" jsonschema:"Event log message"`
	Level       *string `json:"level,omitempty" jsonschema:"Log level (info/warning/error)"`
	Source      *string `json:"source,omitempty" jsonschema:"Source of the event"`
}

type getEventLogArgs struct {
	OperationID int `json:"operation_id" jsonschema:"ID of the operation"`
	Limit       int `json:"limit,omitempty" jsonschema:"Maximum number of log entries to return (default 100)"`
}

type getGlobalSettingsArgs struct{}

type updateGlobalSettingsArgs struct {
	Settings map[string]interface{} `json:"settings" jsonschema:"Settings to update (key-value pairs)"`
}

// Tool handlers

// handleGetOperations retrieves all operations
func (s *Server) handleGetOperations(ctx context.Context, req *mcp.CallToolRequest, args getOperationsArgs) (*mcp.CallToolResult, any, error) {
	operations, err := s.mythicClient.GetOperations(ctx)
	if err != nil {
		return nil, nil, translateError(err)
	}

	// Marshal operations to JSON for display
	data, err := json.MarshalIndent(operations, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(data),
			},
		},
	}, operations, nil
}

// handleGetOperation retrieves a specific operation by ID
func (s *Server) handleGetOperation(ctx context.Context, req *mcp.CallToolRequest, args getOperationArgs) (*mcp.CallToolResult, any, error) {
	operation, err := s.mythicClient.GetOperationByID(ctx, args.OperationID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(operation, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(data),
			},
		},
	}, operation, nil
}

// handleCreateOperation creates a new operation
func (s *Server) handleCreateOperation(ctx context.Context, req *mcp.CallToolRequest, args createOperationArgs) (*mcp.CallToolResult, any, error) {
	createReq := &types.CreateOperationRequest{
		Name: args.Name,
	}

	if args.Webhook != nil {
		createReq.Webhook = *args.Webhook
	}
	if args.Channel != nil {
		createReq.Channel = *args.Channel
	}
	if args.AdminID != nil {
		createReq.AdminID = args.AdminID
	}

	operation, err := s.mythicClient.CreateOperation(ctx, createReq)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(operation, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully created operation '%s' (ID: %d)\n\n%s", operation.Name, operation.ID, string(data)),
			},
		},
	}, operation, nil
}

// handleUpdateOperation updates an existing operation
func (s *Server) handleUpdateOperation(ctx context.Context, req *mcp.CallToolRequest, args updateOperationArgs) (*mcp.CallToolResult, any, error) {
	updateReq := &types.UpdateOperationRequest{
		OperationID: args.OperationID,
		Name:        args.Name,
		Webhook:     args.Webhook,
		Channel:     args.Channel,
		Complete:    args.Complete,
		AdminID:     args.AdminID,
		BannerText:  args.BannerText,
		BannerColor: args.BannerColor,
	}

	operation, err := s.mythicClient.UpdateOperation(ctx, updateReq)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(operation, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully updated operation '%s'\n\n%s", operation.Name, string(data)),
			},
		},
	}, operation, nil
}

// handleSetCurrentOperation sets the current operation context
func (s *Server) handleSetCurrentOperation(ctx context.Context, req *mcp.CallToolRequest, args setCurrentOperationArgs) (*mcp.CallToolResult, any, error) {
	s.mythicClient.SetCurrentOperation(args.OperationID)

	return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Successfully set current operation to ID %d", args.OperationID),
				},
			},
		}, map[string]interface{}{
			"operation_id": args.OperationID,
			"success":      true,
		}, nil
}

// handleGetCurrentOperation gets the current operation context
func (s *Server) handleGetCurrentOperation(ctx context.Context, req *mcp.CallToolRequest, args getCurrentOperationArgs) (*mcp.CallToolResult, any, error) {
	currentOpID := s.mythicClient.GetCurrentOperation()

	if currentOpID == nil {
		return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "No current operation set",
					},
				},
			}, map[string]interface{}{
				"operation_id": nil,
			}, nil
	}

	return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Current operation ID: %d", *currentOpID),
				},
			},
		}, map[string]interface{}{
			"operation_id": *currentOpID,
		}, nil
}

// handleGetOperationOperators gets operators in an operation
func (s *Server) handleGetOperationOperators(ctx context.Context, req *mcp.CallToolRequest, args getOperationOperatorsArgs) (*mcp.CallToolResult, any, error) {
	operators, err := s.mythicClient.GetOperatorsByOperation(ctx, args.OperationID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(operators, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Operators in operation %d:\n\n%s", args.OperationID, string(data)),
			},
		},
	}, operators, nil
}

// handleCreateEventLog creates an event log entry
func (s *Server) handleCreateEventLog(ctx context.Context, req *mcp.CallToolRequest, args createEventLogArgs) (*mcp.CallToolResult, any, error) {
	createReq := &types.CreateOperationEventLogRequest{
		OperationID: args.OperationID,
		Message:     args.Message,
	}

	if args.Level != nil {
		createReq.Level = *args.Level
	} else {
		// Default to "info" level
		createReq.Level = "info"
	}

	if args.Source != nil {
		createReq.Source = *args.Source
	}

	eventLog, err := s.mythicClient.CreateOperationEventLog(ctx, createReq)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(eventLog, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully created event log entry\n\n%s", string(data)),
			},
		},
	}, eventLog, nil
}

// handleGetEventLog retrieves event logs for an operation
func (s *Server) handleGetEventLog(ctx context.Context, req *mcp.CallToolRequest, args getEventLogArgs) (*mcp.CallToolResult, any, error) {
	limit := args.Limit
	if limit <= 0 {
		limit = 100 // Default limit
	}

	eventLogs, err := s.mythicClient.GetOperationEventLog(ctx, args.OperationID, limit)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(eventLogs, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Event logs for operation %d (limit %d):\n\n%s", args.OperationID, limit, string(data)),
			},
		},
	}, eventLogs, nil
}

// handleGetGlobalSettings retrieves global Mythic settings
func (s *Server) handleGetGlobalSettings(ctx context.Context, req *mcp.CallToolRequest, args getGlobalSettingsArgs) (*mcp.CallToolResult, any, error) {
	settings, err := s.mythicClient.GetGlobalSettings(ctx)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Global Mythic settings:\n\n%s", string(data)),
			},
		},
	}, settings, nil
}

// handleUpdateGlobalSettings updates global Mythic settings
func (s *Server) handleUpdateGlobalSettings(ctx context.Context, req *mcp.CallToolRequest, args updateGlobalSettingsArgs) (*mcp.CallToolResult, any, error) {
	err := s.mythicClient.UpdateGlobalSettings(ctx, args.Settings)
	if err != nil {
		return nil, nil, translateError(err)
	}

	return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Successfully updated global settings",
				},
			},
		}, map[string]interface{}{
			"success": true,
		}, nil
}
