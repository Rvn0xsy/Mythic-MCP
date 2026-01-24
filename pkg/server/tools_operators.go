package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nbaertsch/mythic-sdk-go/pkg/mythic/types"
)

// registerOperatorsTools registers operator management MCP tools
func (s *Server) registerOperatorsTools() {
	// mythic_get_operators - List all operators
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_operators",
		Description: "Get a list of all operators (users) in the Mythic instance",
	}, s.handleGetOperators)

	// mythic_get_operator - Get specific operator by ID
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_operator",
		Description: "Get details of a specific operator by ID",
	}, s.handleGetOperator)

	// mythic_create_operator - Create new operator
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_create_operator",
		Description: "Create a new operator (user) account",
	}, s.handleCreateOperator)

	// mythic_update_operator_status - Update operator status
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_update_operator_status",
		Description: "Update operator status (active/inactive, admin privileges, deleted)",
	}, s.handleUpdateOperatorStatus)

	// mythic_update_password_email - Update password and email
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_update_password_email",
		Description: "Update operator password and/or email address",
	}, s.handleUpdatePasswordEmail)

	// mythic_get_operator_preferences - Get operator preferences
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_operator_preferences",
		Description: "Get UI preferences for an operator",
	}, s.handleGetOperatorPreferences)

	// mythic_update_operator_preferences - Update operator preferences
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_update_operator_preferences",
		Description: "Update UI preferences for an operator",
	}, s.handleUpdateOperatorPreferences)

	// mythic_get_operator_secrets - Get operator secrets
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_operator_secrets",
		Description: "Get secrets/keys associated with an operator",
	}, s.handleGetOperatorSecrets)

	// mythic_update_operator_secrets - Update operator secrets
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_update_operator_secrets",
		Description: "Update secrets/keys for an operator",
	}, s.handleUpdateOperatorSecrets)

	// mythic_get_invite_links - Get invite links
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_invite_links",
		Description: "Get all invitation links for new operators",
	}, s.handleGetInviteLinks)

	// mythic_create_invite_link - Create invite link
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_create_invite_link",
		Description: "Create a new invitation link for operator registration",
	}, s.handleCreateInviteLink)

	// mythic_update_operator_operation - Update operator in operation
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_update_operator_operation",
		Description: "Add or remove operators from an operation",
	}, s.handleUpdateOperatorOperation)
}

// Tool argument types for operators tools

type getOperatorsArgs struct{}

type getOperatorArgs struct {
	OperatorID int `json:"operator_id" jsonschema:"required,description=ID of the operator to retrieve"`
}

type createOperatorArgs struct {
	Username string  `json:"username" jsonschema:"required,description=Username for the new operator"`
	Password string  `json:"password" jsonschema:"required,description=Password (minimum 12 characters)"`
	Email    *string `json:"email,omitempty" jsonschema:"description=Email address"`
	Bot      *bool   `json:"bot,omitempty" jsonschema:"description=Create as bot account"`
}

type updateOperatorStatusArgs struct {
	OperatorID int   `json:"operator_id" jsonschema:"required,description=ID of the operator to update"`
	Active     *bool `json:"active,omitempty" jsonschema:"description=Set operator active/inactive"`
	Admin      *bool `json:"admin,omitempty" jsonschema:"description=Grant/revoke admin privileges"`
	Deleted    *bool `json:"deleted,omitempty" jsonschema:"description=Mark operator as deleted"`
}

type updatePasswordEmailArgs struct {
	OperatorID  int     `json:"operator_id" jsonschema:"required,description=ID of the operator"`
	OldPassword string  `json:"old_password" jsonschema:"required,description=Current password"`
	NewPassword *string `json:"new_password,omitempty" jsonschema:"description=New password (min 12 chars)"`
	Email       *string `json:"email,omitempty" jsonschema:"description=New email address"`
}

type getOperatorPreferencesArgs struct {
	OperatorID int `json:"operator_id" jsonschema:"required,description=ID of the operator"`
}

type updateOperatorPreferencesArgs struct {
	OperatorID  int                    `json:"operator_id" jsonschema:"required,description=ID of the operator"`
	Preferences map[string]interface{} `json:"preferences" jsonschema:"required,description=Preferences to update (key-value pairs)"`
}

type getOperatorSecretsArgs struct {
	OperatorID int `json:"operator_id" jsonschema:"required,description=ID of the operator"`
}

type updateOperatorSecretsArgs struct {
	OperatorID int                    `json:"operator_id" jsonschema:"required,description=ID of the operator"`
	Secrets    map[string]interface{} `json:"secrets" jsonschema:"required,description=Secrets to update (key-value pairs)"`
}

type getInviteLinksArgs struct{}

type createInviteLinkArgs struct {
	OperationID   *int    `json:"operation_id,omitempty" jsonschema:"description=Operation to associate link with"`
	OperationRole *string `json:"operation_role,omitempty" jsonschema:"description=Role for new users (operator/spectator)"`
	MaxUses       *int    `json:"max_uses,omitempty" jsonschema:"description=Maximum number of uses"`
	Name          *string `json:"name,omitempty" jsonschema:"description=Human-readable name for the link"`
	ShortCode     *string `json:"short_code,omitempty" jsonschema:"description=Custom short code"`
}

type updateOperatorOperationArgs struct {
	OperationID        int    `json:"operation_id" jsonschema:"required,description=Operation to modify"`
	AddUsers           *[]int `json:"add_users,omitempty" jsonschema:"description=Operator IDs to add with full access"`
	RemoveUsers        *[]int `json:"remove_users,omitempty" jsonschema:"description=Operator IDs to remove"`
	ViewModeOperators  *[]int `json:"view_mode_operators,omitempty" jsonschema:"description=Operator IDs to set as view-only"`
	ViewModeSpectators *[]int `json:"view_mode_spectators,omitempty" jsonschema:"description=Operator IDs to set as spectators"`
}

// Tool handlers

// handleGetOperators retrieves all operators
func (s *Server) handleGetOperators(ctx context.Context, req *mcp.CallToolRequest, args getOperatorsArgs) (*mcp.CallToolResult, any, error) {
	operators, err := s.mythicClient.GetOperators(ctx)
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
				Text: fmt.Sprintf("Operators (%d total):\n\n%s", len(operators), string(data)),
			},
		},
	}, operators, nil
}

// handleGetOperator retrieves a specific operator by ID
func (s *Server) handleGetOperator(ctx context.Context, req *mcp.CallToolRequest, args getOperatorArgs) (*mcp.CallToolResult, any, error) {
	operator, err := s.mythicClient.GetOperatorByID(ctx, args.OperatorID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(operator, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Operator details:\n\n%s", string(data)),
			},
		},
	}, operator, nil
}

// handleCreateOperator creates a new operator
func (s *Server) handleCreateOperator(ctx context.Context, req *mcp.CallToolRequest, args createOperatorArgs) (*mcp.CallToolResult, any, error) {
	createReq := &types.CreateOperatorRequest{
		Username: args.Username,
		Password: args.Password,
	}

	if args.Email != nil {
		createReq.Email = *args.Email
	}
	if args.Bot != nil {
		createReq.Bot = *args.Bot
	}

	operator, err := s.mythicClient.CreateOperator(ctx, createReq)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(operator, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully created operator '%s' (ID: %d)\n\n%s", operator.Username, operator.ID, string(data)),
			},
		},
	}, operator, nil
}

// handleUpdateOperatorStatus updates operator status
func (s *Server) handleUpdateOperatorStatus(ctx context.Context, req *mcp.CallToolRequest, args updateOperatorStatusArgs) (*mcp.CallToolResult, any, error) {
	updateReq := &types.UpdateOperatorStatusRequest{
		OperatorID: args.OperatorID,
		Active:     args.Active,
		Admin:      args.Admin,
		Deleted:    args.Deleted,
	}

	err := s.mythicClient.UpdateOperatorStatus(ctx, updateReq)
	if err != nil {
		return nil, nil, translateError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully updated operator %d status", args.OperatorID),
			},
		},
	}, map[string]interface{}{
		"operator_id": args.OperatorID,
		"success":     true,
	}, nil
}

// handleUpdatePasswordEmail updates password and email
func (s *Server) handleUpdatePasswordEmail(ctx context.Context, req *mcp.CallToolRequest, args updatePasswordEmailArgs) (*mcp.CallToolResult, any, error) {
	updateReq := &types.UpdatePasswordAndEmailRequest{
		OperatorID:  args.OperatorID,
		OldPassword: args.OldPassword,
		NewPassword: args.NewPassword,
		Email:       args.Email,
	}

	err := s.mythicClient.UpdatePasswordAndEmail(ctx, updateReq)
	if err != nil {
		return nil, nil, translateError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: "Successfully updated operator password and/or email",
			},
		},
	}, map[string]interface{}{
		"success": true,
	}, nil
}

// handleGetOperatorPreferences retrieves operator preferences
func (s *Server) handleGetOperatorPreferences(ctx context.Context, req *mcp.CallToolRequest, args getOperatorPreferencesArgs) (*mcp.CallToolResult, any, error) {
	prefs, err := s.mythicClient.GetOperatorPreferences(ctx, args.OperatorID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Operator %d preferences:\n\n%s", args.OperatorID, string(data)),
			},
		},
	}, prefs, nil
}

// handleUpdateOperatorPreferences updates operator preferences
func (s *Server) handleUpdateOperatorPreferences(ctx context.Context, req *mcp.CallToolRequest, args updateOperatorPreferencesArgs) (*mcp.CallToolResult, any, error) {
	updateReq := &types.UpdateOperatorPreferencesRequest{
		OperatorID:  args.OperatorID,
		Preferences: args.Preferences,
	}

	err := s.mythicClient.UpdateOperatorPreferences(ctx, updateReq)
	if err != nil {
		return nil, nil, translateError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully updated preferences for operator %d", args.OperatorID),
			},
		},
	}, map[string]interface{}{
		"operator_id": args.OperatorID,
		"success":     true,
	}, nil
}

// handleGetOperatorSecrets retrieves operator secrets
func (s *Server) handleGetOperatorSecrets(ctx context.Context, req *mcp.CallToolRequest, args getOperatorSecretsArgs) (*mcp.CallToolResult, any, error) {
	secrets, err := s.mythicClient.GetOperatorSecrets(ctx, args.OperatorID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Operator %d secrets:\n\n%s", args.OperatorID, string(data)),
			},
		},
	}, secrets, nil
}

// handleUpdateOperatorSecrets updates operator secrets
func (s *Server) handleUpdateOperatorSecrets(ctx context.Context, req *mcp.CallToolRequest, args updateOperatorSecretsArgs) (*mcp.CallToolResult, any, error) {
	updateReq := &types.UpdateOperatorSecretsRequest{
		OperatorID: args.OperatorID,
		Secrets:    args.Secrets,
	}

	err := s.mythicClient.UpdateOperatorSecrets(ctx, updateReq)
	if err != nil {
		return nil, nil, translateError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully updated secrets for operator %d", args.OperatorID),
			},
		},
	}, map[string]interface{}{
		"operator_id": args.OperatorID,
		"success":     true,
	}, nil
}

// handleGetInviteLinks retrieves all invite links
func (s *Server) handleGetInviteLinks(ctx context.Context, req *mcp.CallToolRequest, args getInviteLinksArgs) (*mcp.CallToolResult, any, error) {
	links, err := s.mythicClient.GetInviteLinks(ctx)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(links, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Invite links (%d total):\n\n%s", len(links), string(data)),
			},
		},
	}, links, nil
}

// handleCreateInviteLink creates a new invite link
func (s *Server) handleCreateInviteLink(ctx context.Context, req *mcp.CallToolRequest, args createInviteLinkArgs) (*mcp.CallToolResult, any, error) {
	createReq := &types.CreateInviteLinkRequest{}

	if args.OperationID != nil {
		createReq.OperationID = args.OperationID
	}
	if args.OperationRole != nil {
		createReq.OperationRole = *args.OperationRole
	}
	if args.MaxUses != nil {
		createReq.MaxUses = *args.MaxUses
	}
	if args.Name != nil {
		createReq.Name = *args.Name
	}
	if args.ShortCode != nil {
		createReq.ShortCode = *args.ShortCode
	}

	link, err := s.mythicClient.CreateInviteLink(ctx, createReq)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(link, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully created invite link\n\n%s", string(data)),
			},
		},
	}, link, nil
}

// handleUpdateOperatorOperation updates operators in an operation
func (s *Server) handleUpdateOperatorOperation(ctx context.Context, req *mcp.CallToolRequest, args updateOperatorOperationArgs) (*mcp.CallToolResult, any, error) {
	updateReq := &types.UpdateOperatorOperationRequest{
		OperationID: args.OperationID,
	}

	if args.AddUsers != nil {
		updateReq.AddUsers = *args.AddUsers
	}
	if args.RemoveUsers != nil {
		updateReq.RemoveUsers = *args.RemoveUsers
	}
	if args.ViewModeOperators != nil {
		updateReq.ViewModeOperators = *args.ViewModeOperators
	}
	if args.ViewModeSpectators != nil {
		updateReq.ViewModeSpectators = *args.ViewModeSpectators
	}

	err := s.mythicClient.UpdateOperatorOperation(ctx, updateReq)
	if err != nil {
		return nil, nil, translateError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully updated operators for operation %d", args.OperationID),
			},
		},
	}, map[string]interface{}{
		"operation_id": args.OperationID,
		"success":      true,
	}, nil
}
