package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nbaertsch/mythic-sdk-go/pkg/mythic/types"
)

// registerPayloadsTools registers payload management MCP tools
func (s *Server) registerPayloadsTools() {
	// mythic_get_payloads - List all payloads
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_payloads",
		Description: "Get a list of all payloads in Mythic",
	}, s.handleGetPayloads)

	// mythic_get_payload - Get specific payload
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_payload",
		Description: "Get details of a specific payload by UUID",
	}, s.handleGetPayload)

	// mythic_get_payload_types - List payload types
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_payload_types",
		Description: "Get list of available payload types (agent types)",
	}, s.handleGetPayloadTypes)

	// mythic_create_payload - Create a new payload
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_create_payload",
		Description: "Create/build a new payload",
	}, s.handleCreatePayload)

	// mythic_update_payload - Update payload properties
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_update_payload",
		Description: "Update a payload's properties (description, tag, etc.)",
	}, s.handleUpdatePayload)

	// mythic_delete_payload - Delete a payload
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_delete_payload",
		Description: "Delete a payload from Mythic",
	}, s.handleDeletePayload)

	// mythic_rebuild_payload - Rebuild a payload
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_rebuild_payload",
		Description: "Rebuild/regenerate an existing payload",
	}, s.handleRebuildPayload)

	// mythic_export_payload_config - Export payload configuration
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_export_payload_config",
		Description: "Export a payload's configuration as JSON",
	}, s.handleExportPayloadConfig)

	// mythic_get_payload_commands - Get commands for a payload
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_payload_commands",
		Description: "Get list of commands available in a payload",
	}, s.handleGetPayloadCommands)

	// mythic_get_payload_on_host - Get payloads on hosts
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_payload_on_host",
		Description: "Get list of payloads deployed on hosts in an operation",
	}, s.handleGetPayloadOnHost)

	// mythic_wait_for_payload - Wait for payload build completion
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_wait_for_payload",
		Description: "Wait for a payload build to complete with timeout",
	}, s.handleWaitForPayload)

	// mythic_download_payload - Download a built payload
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_download_payload",
		Description: "Download a built payload's binary file (base64-encoded)",
	}, s.handleDownloadPayload)
}

// Tool argument types for payload tools

type getPayloadsArgs struct{}

type getPayloadArgs struct {
	PayloadUUID string `json:"payload_uuid" jsonschema:"required,description=UUID of the payload to retrieve"`
}

type getPayloadTypesArgs struct{}

type createPayloadArgs struct {
	PayloadType     string                   `json:"payload_type" jsonschema:"required,description=Payload type name (agent type)"`
	Description     string                   `json:"description,omitempty" jsonschema:"description=Description of the payload"`
	Tag             string                   `json:"tag,omitempty" jsonschema:"description=Tag for the payload"`
	Filename        string                   `json:"filename,omitempty" jsonschema:"description=Filename for the payload"`
	OS              string                   `json:"os,omitempty" jsonschema:"description=Operating system for the payload"`
	SelectedOS      string                   `json:"selected_os,omitempty" jsonschema:"description=Selected OS variant"`
	Commands        []string                 `json:"commands,omitempty" jsonschema:"description=List of command names to include"`
	C2Profiles      []map[string]interface{} `json:"c2_profiles,omitempty" jsonschema:"description=C2 profile configurations (array of {name, parameters})"`
	BuildParameters map[string]interface{}   `json:"build_parameters,omitempty" jsonschema:"description=Build parameter key-value pairs"`
	WrapperPayload  string                   `json:"wrapper_payload,omitempty" jsonschema:"description=UUID of payload to wrap"`
}

type updatePayloadArgs struct {
	PayloadUUID   string `json:"payload_uuid" jsonschema:"required,description=UUID of the payload to update"`
	Description   *string `json:"description,omitempty" jsonschema:"description=Update payload description"`
	CallbackAlert *bool   `json:"callback_alert,omitempty" jsonschema:"description=Update callback alert setting"`
	Deleted       *bool   `json:"deleted,omitempty" jsonschema:"description=Mark payload as deleted"`
}

type deletePayloadArgs struct {
	PayloadUUID string `json:"payload_uuid" jsonschema:"required,description=UUID of the payload to delete"`
}

type rebuildPayloadArgs struct {
	PayloadUUID string `json:"payload_uuid" jsonschema:"required,description=UUID of the payload to rebuild"`
}

type exportPayloadConfigArgs struct {
	PayloadUUID string `json:"payload_uuid" jsonschema:"required,description=UUID of the payload to export"`
}

type getPayloadCommandsArgs struct {
	PayloadID int `json:"payload_id" jsonschema:"required,description=ID of the payload"`
}

type getPayloadOnHostArgs struct {
	OperationID int `json:"operation_id" jsonschema:"required,description=ID of the operation"`
}

type waitForPayloadArgs struct {
	PayloadUUID string `json:"payload_uuid" jsonschema:"required,description=UUID of the payload to wait for"`
	Timeout     int    `json:"timeout,omitempty" jsonschema:"description=Timeout in seconds (default 60)"`
}

type downloadPayloadArgs struct {
	PayloadUUID string `json:"payload_uuid" jsonschema:"required,description=UUID of the payload to download"`
}

// Tool handlers

// handleGetPayloads retrieves all payloads
func (s *Server) handleGetPayloads(ctx context.Context, req *mcp.CallToolRequest, args getPayloadsArgs) (*mcp.CallToolResult, any, error) {
	payloads, err := s.mythicClient.GetPayloads(ctx)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(payloads, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("All payloads (%d total):\n\n%s", len(payloads), string(data)),
			},
		},
	}, payloads, nil
}

// handleGetPayload retrieves a specific payload by UUID
func (s *Server) handleGetPayload(ctx context.Context, req *mcp.CallToolRequest, args getPayloadArgs) (*mcp.CallToolResult, any, error) {
	payload, err := s.mythicClient.GetPayloadByUUID(ctx, args.PayloadUUID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	status := payload.BuildPhase
	if status == "" {
		status = "unknown"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Payload %s (%s): %s\nType: %s\nDescription: %s\n\n%s",
					payload.UUID, status, payload.BuildMessage, payload.PayloadType, payload.Description, string(data)),
			},
		},
	}, payload, nil
}

// handleGetPayloadTypes retrieves all payload types
func (s *Server) handleGetPayloadTypes(ctx context.Context, req *mcp.CallToolRequest, args getPayloadTypesArgs) (*mcp.CallToolResult, any, error) {
	payloadTypes, err := s.mythicClient.GetPayloadTypes(ctx)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(payloadTypes, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Payload types (%d total):\n\n%s", len(payloadTypes), string(data)),
			},
		},
	}, payloadTypes, nil
}

// handleCreatePayload creates a new payload
func (s *Server) handleCreatePayload(ctx context.Context, req *mcp.CallToolRequest, args createPayloadArgs) (*mcp.CallToolResult, any, error) {
	createReq := &types.CreatePayloadRequest{
		PayloadType: args.PayloadType,
	}

	// Optional fields
	if args.Description != "" {
		createReq.Description = args.Description
	}
	if args.Tag != "" {
		createReq.Tag = args.Tag
	}
	if args.Filename != "" {
		createReq.Filename = args.Filename
	}
	if args.OS != "" {
		createReq.OS = args.OS
	}
	if args.SelectedOS != "" {
		createReq.SelectedOS = args.SelectedOS
	}
	if args.Commands != nil {
		createReq.Commands = args.Commands
	}
	if args.BuildParameters != nil {
		createReq.BuildParameters = args.BuildParameters
	}
	if args.WrapperPayload != "" {
		createReq.WrapperPayload = args.WrapperPayload
	}

	// Convert C2 profiles from map to C2ProfileConfig
	if args.C2Profiles != nil {
		c2Profiles := make([]types.C2ProfileConfig, 0, len(args.C2Profiles))
		for _, profile := range args.C2Profiles {
			name, ok := profile["name"].(string)
			if !ok {
				continue
			}
			params, _ := profile["parameters"].(map[string]interface{})
			c2Profiles = append(c2Profiles, types.C2ProfileConfig{
				Name:       name,
				Parameters: params,
			})
		}
		createReq.C2Profiles = c2Profiles
	}

	payload, err := s.mythicClient.CreatePayload(ctx, createReq)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully created payload %s\nType: %s\nStatus: %s\n\n%s",
					payload.UUID, payload.PayloadType, payload.BuildPhase, string(data)),
			},
		},
	}, payload, nil
}

// handleUpdatePayload updates a payload's properties
func (s *Server) handleUpdatePayload(ctx context.Context, req *mcp.CallToolRequest, args updatePayloadArgs) (*mcp.CallToolResult, any, error) {
	updateReq := &types.UpdatePayloadRequest{
		UUID: args.PayloadUUID,
	}

	if args.Description != nil {
		updateReq.Description = args.Description
	}
	if args.CallbackAlert != nil {
		updateReq.CallbackAlert = args.CallbackAlert
	}
	if args.Deleted != nil {
		updateReq.Deleted = args.Deleted
	}

	payload, err := s.mythicClient.UpdatePayload(ctx, updateReq)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully updated payload %s\n\n%s", payload.UUID, string(data)),
			},
		},
	}, payload, nil
}

// handleDeletePayload deletes a payload
func (s *Server) handleDeletePayload(ctx context.Context, req *mcp.CallToolRequest, args deletePayloadArgs) (*mcp.CallToolResult, any, error) {
	err := s.mythicClient.DeletePayload(ctx, args.PayloadUUID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully deleted payload %s", args.PayloadUUID),
			},
		},
	}, map[string]interface{}{
		"payload_uuid": args.PayloadUUID,
		"success":      true,
	}, nil
}

// handleRebuildPayload rebuilds a payload
func (s *Server) handleRebuildPayload(ctx context.Context, req *mcp.CallToolRequest, args rebuildPayloadArgs) (*mcp.CallToolResult, any, error) {
	payload, err := s.mythicClient.RebuildPayload(ctx, args.PayloadUUID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully initiated rebuild of payload %s\nStatus: %s\n\n%s",
					payload.UUID, payload.BuildPhase, string(data)),
			},
		},
	}, payload, nil
}

// handleExportPayloadConfig exports a payload's configuration
func (s *Server) handleExportPayloadConfig(ctx context.Context, req *mcp.CallToolRequest, args exportPayloadConfigArgs) (*mcp.CallToolResult, any, error) {
	config, err := s.mythicClient.ExportPayloadConfig(ctx, args.PayloadUUID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully exported payload configuration\n\nConfig (length: %d bytes):\n%s",
					len(config), config),
			},
		},
	}, map[string]interface{}{
		"payload_uuid": args.PayloadUUID,
		"config":       config,
		"size":         len(config),
	}, nil
}

// handleGetPayloadCommands retrieves commands for a payload
func (s *Server) handleGetPayloadCommands(ctx context.Context, req *mcp.CallToolRequest, args getPayloadCommandsArgs) (*mcp.CallToolResult, any, error) {
	commands, err := s.mythicClient.GetPayloadCommands(ctx, args.PayloadID)
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
				Text: fmt.Sprintf("Commands for payload %d (%d total):\n\n%s",
					args.PayloadID, len(commands), string(data)),
			},
		},
	}, commands, nil
}

// handleGetPayloadOnHost retrieves payloads on hosts
func (s *Server) handleGetPayloadOnHost(ctx context.Context, req *mcp.CallToolRequest, args getPayloadOnHostArgs) (*mcp.CallToolResult, any, error) {
	payloadsOnHost, err := s.mythicClient.GetPayloadOnHost(ctx, args.OperationID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(payloadsOnHost, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Payloads on host for operation %d (%d total):\n\n%s",
					args.OperationID, len(payloadsOnHost), string(data)),
			},
		},
	}, payloadsOnHost, nil
}

// handleWaitForPayload waits for a payload build to complete
func (s *Server) handleWaitForPayload(ctx context.Context, req *mcp.CallToolRequest, args waitForPayloadArgs) (*mcp.CallToolResult, any, error) {
	timeout := args.Timeout
	if timeout == 0 {
		timeout = 60
	}

	err := s.mythicClient.WaitForPayloadComplete(ctx, args.PayloadUUID, timeout)
	if err != nil {
		return nil, nil, translateError(err)
	}

	// Get the payload details after completion
	payload, err := s.mythicClient.GetPayloadByUUID(ctx, args.PayloadUUID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Payload %s build completed\nStatus: %s\nMessage: %s\n\n%s",
					payload.UUID, payload.BuildPhase, payload.BuildMessage, string(data)),
			},
		},
	}, payload, nil
}

// handleDownloadPayload downloads a built payload
func (s *Server) handleDownloadPayload(ctx context.Context, req *mcp.CallToolRequest, args downloadPayloadArgs) (*mcp.CallToolResult, any, error) {
	payloadData, err := s.mythicClient.DownloadPayload(ctx, args.PayloadUUID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	// Encode to base64 for transfer
	encodedData := base64.StdEncoding.EncodeToString(payloadData)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully downloaded payload %s\nSize: %d bytes (raw), %d bytes (base64)",
					args.PayloadUUID, len(payloadData), len(encodedData)),
			},
		},
	}, map[string]interface{}{
		"payload_uuid": args.PayloadUUID,
		"payload_data": encodedData,
		"size_raw":     len(payloadData),
		"size_base64":  len(encodedData),
	}, nil
}
