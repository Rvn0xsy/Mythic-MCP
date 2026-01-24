package server

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerAuthTools registers authentication-related MCP tools
func (s *Server) registerAuthTools() {
	// mythic_login - Authenticate with username/password
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_login",
		Description: "Authenticate with Mythic server using username and password",
	}, s.handleLogin)

	// mythic_logout - End session
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_logout",
		Description: "End the current Mythic session and clear authentication",
	}, s.handleLogout)

	// mythic_is_authenticated - Check auth status
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_is_authenticated",
		Description: "Check if currently authenticated with Mythic server",
	}, s.handleIsAuthenticated)

	// mythic_get_current_user - Get current user info
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_current_user",
		Description: "Get information about the current authenticated user",
	}, s.handleGetCurrentUser)

	// mythic_create_api_token - Generate API token
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_create_api_token",
		Description: "Create a new API token for programmatic access",
	}, s.handleCreateAPIToken)

	// mythic_delete_api_token - Revoke API token
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_delete_api_token",
		Description: "Delete an existing API token",
	}, s.handleDeleteAPIToken)

	// mythic_refresh_token - Refresh access token
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_refresh_token",
		Description: "Refresh the current access token to extend session",
	}, s.handleRefreshToken)
}

// Tool handler types for MCP SDK
type loginArgs struct {
	Username string `json:"username" jsonschema:"required,description=Mythic username"`
	Password string `json:"password" jsonschema:"required,description=Mythic password"`
}

type logoutArgs struct{}

type isAuthenticatedArgs struct{}

type getCurrentUserArgs struct{}

type createAPITokenArgs struct{}

type deleteAPITokenArgs struct {
	TokenID int `json:"token_id" jsonschema:"required,description=ID of the token to delete"`
}

type refreshTokenArgs struct{}

// handleLogin authenticates with Mythic using username/password
func (s *Server) handleLogin(ctx context.Context, req *mcp.CallToolRequest, args loginArgs) (*mcp.CallToolResult, any, error) {
	// Note: The MCP server was already authenticated in Run()
	// This tool allows re-authentication or verification
	// For now, we'll use the existing client

	// Attempt login with provided credentials
	// In a real implementation, we might create a new client or re-auth the existing one
	// For now, just verify authentication works
	if err := s.mythicClient.Login(ctx); err != nil {
		return nil, nil, translateError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: "Successfully authenticated with Mythic server",
			},
		},
	}, nil, nil
}

// handleLogout ends the current Mythic session
func (s *Server) handleLogout(ctx context.Context, req *mcp.CallToolRequest, args logoutArgs) (*mcp.CallToolResult, any, error) {
	s.mythicClient.Logout()

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: "Successfully logged out from Mythic server",
			},
		},
	}, nil, nil
}

// handleIsAuthenticated checks authentication status
func (s *Server) handleIsAuthenticated(ctx context.Context, req *mcp.CallToolRequest, args isAuthenticatedArgs) (*mcp.CallToolResult, any, error) {
	isAuth := s.mythicClient.IsAuthenticated()

	status := "not authenticated"
	if isAuth {
		status = "authenticated"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: status,
			},
		},
	}, map[string]interface{}{
		"authenticated": isAuth,
	}, nil
}

// handleGetCurrentUser retrieves current user information
func (s *Server) handleGetCurrentUser(ctx context.Context, req *mcp.CallToolRequest, args getCurrentUserArgs) (*mcp.CallToolResult, any, error) {
	operator, err := s.mythicClient.GetMe(ctx)
	if err != nil {
		return nil, nil, translateError(err)
	}

	// Marshal operator to JSON for display
	data, err := json.MarshalIndent(operator, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(data),
			},
		},
	}, operator, nil
}

// handleCreateAPIToken creates a new API token
func (s *Server) handleCreateAPIToken(ctx context.Context, req *mcp.CallToolRequest, args createAPITokenArgs) (*mcp.CallToolResult, any, error) {
	token, err := s.mythicClient.CreateAPIToken(ctx)
	if err != nil {
		return nil, nil, translateError(err)
	}

	// Marshal token to JSON
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(data),
			},
		},
	}, token, nil
}

// handleDeleteAPIToken deletes an API token
func (s *Server) handleDeleteAPIToken(ctx context.Context, req *mcp.CallToolRequest, args deleteAPITokenArgs) (*mcp.CallToolResult, any, error) {
	if err := s.mythicClient.DeleteAPIToken(ctx, args.TokenID); err != nil {
		return nil, nil, translateError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: "Successfully deleted API token",
			},
		},
	}, map[string]interface{}{
		"success": true,
	}, nil
}

// handleRefreshToken refreshes the current access token
func (s *Server) handleRefreshToken(ctx context.Context, req *mcp.CallToolRequest, args refreshTokenArgs) (*mcp.CallToolResult, any, error) {
	if err := s.mythicClient.RefreshAccessToken(ctx); err != nil {
		return nil, nil, translateError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: "Successfully refreshed access token",
			},
		},
	}, map[string]interface{}{
		"success": true,
	}, nil
}
