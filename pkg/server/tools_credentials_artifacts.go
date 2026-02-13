package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nbaertsch/mythic-sdk-go/pkg/mythic/types"
)

// registerCredentialsTools registers credential management MCP tools
func (s *Server) registerCredentialsTools() {
	// mythic_get_credentials - List all credentials
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_credentials",
		Description: "Get a list of all credentials stored in Mythic",
	}, s.handleGetCredentials)

	// mythic_get_credential - Get specific credential
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_credential",
		Description: "Get details of a specific credential by ID",
	}, s.handleGetCredential)

	// mythic_get_operation_credentials - Get credentials for operation
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_operation_credentials",
		Description: "Get credentials filtered by operation",
	}, s.handleGetOperationCredentials)

	// mythic_create_credential - Create new credential
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_create_credential",
		Description: "Create a new credential entry",
	}, s.handleCreateCredential)

	// mythic_update_credential - Update existing credential
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_update_credential",
		Description: "Update an existing credential's properties",
	}, s.handleUpdateCredential)

	// mythic_delete_credential - Delete credential
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_delete_credential",
		Description: "Delete a credential from Mythic",
	}, s.handleDeleteCredential)
}

// registerArtifactsTools registers artifact management MCP tools
func (s *Server) registerArtifactsTools() {
	// mythic_get_artifacts - List all artifacts
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_artifacts",
		Description: "Get a list of all artifacts (IOCs, forensic evidence)",
	}, s.handleGetArtifacts)

	// mythic_get_artifact - Get specific artifact
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_artifact",
		Description: "Get details of a specific artifact by ID",
	}, s.handleGetArtifact)

	// mythic_get_operation_artifacts - Get artifacts for operation
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_operation_artifacts",
		Description: "Get artifacts filtered by operation",
	}, s.handleGetOperationArtifacts)

	// mythic_get_host_artifacts - Get artifacts for host
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_host_artifacts",
		Description: "Get artifacts filtered by host",
	}, s.handleGetHostArtifacts)

	// mythic_get_artifacts_by_type - Get artifacts by type
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_artifacts_by_type",
		Description: "Get artifacts filtered by artifact type",
	}, s.handleGetArtifactsByType)

	// mythic_create_artifact - Create new artifact
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_create_artifact",
		Description: "Create a new artifact entry (IOC, forensic evidence)",
	}, s.handleCreateArtifact)

	// mythic_update_artifact - Update existing artifact
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_update_artifact",
		Description: "Update an existing artifact's properties",
	}, s.handleUpdateArtifact)

	// mythic_delete_artifact - Delete artifact
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_delete_artifact",
		Description: "Delete an artifact from Mythic",
	}, s.handleDeleteArtifact)
}

// Tool argument types for credentials tools

type getCredentialsArgs struct{}

type getCredentialArgs struct {
	CredentialID int `json:"credential_id" jsonschema:"ID of the credential to retrieve"`
}

type getOperationCredentialsArgs struct {
	OperationID int `json:"operation_id" jsonschema:"Operation ID to filter credentials"`
}

type createCredentialArgs struct {
	Type       string  `json:"type" jsonschema:"Credential type (plaintext/hash/key/ticket/etc.)"`
	Account    string  `json:"account" jsonschema:"Account/username"`
	Realm      *string `json:"realm,omitempty" jsonschema:"Domain/realm"`
	Credential string  `json:"credential" jsonschema:"The actual credential (password/hash/key)"`
	Comment    *string `json:"comment,omitempty" jsonschema:"Additional notes about the credential"`
	TaskID     *int    `json:"task_id,omitempty" jsonschema:"Task ID that discovered this credential"`
}

type updateCredentialArgs struct {
	CredentialID int     `json:"credential_id" jsonschema:"ID of the credential to update"`
	Type         *string `json:"type,omitempty" jsonschema:"New credential type"`
	Account      *string `json:"account,omitempty" jsonschema:"New account/username"`
	Realm        *string `json:"realm,omitempty" jsonschema:"New domain/realm"`
	Credential   *string `json:"credential,omitempty" jsonschema:"New credential value"`
	Comment      *string `json:"comment,omitempty" jsonschema:"New comment"`
}

type deleteCredentialArgs struct {
	CredentialID int `json:"credential_id" jsonschema:"ID of the credential to delete"`
}

// Tool argument types for artifacts tools

type getArtifactsArgs struct{}

type getArtifactArgs struct {
	ArtifactID int `json:"artifact_id" jsonschema:"ID of the artifact to retrieve"`
}

type getOperationArtifactsArgs struct {
	OperationID int `json:"operation_id" jsonschema:"Operation ID to filter artifacts"`
}

type getHostArtifactsArgs struct {
	Host string `json:"host" jsonschema:"Hostname to filter artifacts"`
}

type getArtifactsByTypeArgs struct {
	ArtifactType string `json:"artifact_type" jsonschema:"Artifact type to filter (File Write/Registry Write/etc.)"`
}

type createArtifactArgs struct {
	Artifact     string  `json:"artifact" jsonschema:"The artifact (file path/registry key/etc.)"`
	BaseArtifact *string `json:"base_artifact,omitempty" jsonschema:"Base artifact for pattern matching"`
	Host         *string `json:"host,omitempty" jsonschema:"Hostname where artifact was observed"`
	TaskID       *int    `json:"task_id,omitempty" jsonschema:"Task ID that created this artifact"`
}

type updateArtifactArgs struct {
	ArtifactID int     `json:"artifact_id" jsonschema:"ID of the artifact to update"`
	Host       *string `json:"host,omitempty" jsonschema:"New hostname"`
}

type deleteArtifactArgs struct {
	ArtifactID int `json:"artifact_id" jsonschema:"ID of the artifact to delete"`
}

// Tool handlers for credentials

// handleGetCredentials retrieves all credentials
func (s *Server) handleGetCredentials(ctx context.Context, req *mcp.CallToolRequest, args getCredentialsArgs) (*mcp.CallToolResult, any, error) {
	credentials, err := s.mythicClient.GetCredentials(ctx)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Credentials (%d total):\n\n%s", len(credentials), string(data)),
			},
		},
	}, wrapList(credentials), nil
}

// handleGetCredential retrieves a specific credential by ID
func (s *Server) handleGetCredential(ctx context.Context, req *mcp.CallToolRequest, args getCredentialArgs) (*mcp.CallToolResult, any, error) {
	credential, err := s.mythicClient.GetCredentialByID(ctx, args.CredentialID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(credential, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Credential details:\n\n%s", string(data)),
			},
		},
	}, credential, nil
}

// handleGetOperationCredentials retrieves credentials for an operation
func (s *Server) handleGetOperationCredentials(ctx context.Context, req *mcp.CallToolRequest, args getOperationCredentialsArgs) (*mcp.CallToolResult, any, error) {
	credentials, err := s.mythicClient.GetCredentialsByOperation(ctx, args.OperationID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Credentials for operation %d (%d total):\n\n%s", args.OperationID, len(credentials), string(data)),
			},
		},
	}, wrapList(credentials), nil
}

// handleCreateCredential creates a new credential
func (s *Server) handleCreateCredential(ctx context.Context, req *mcp.CallToolRequest, args createCredentialArgs) (*mcp.CallToolResult, any, error) {
	createReq := &types.CreateCredentialRequest{
		Type:       args.Type,
		Account:    args.Account,
		Credential: args.Credential,
		TaskID:     args.TaskID,
	}

	if args.Realm != nil {
		createReq.Realm = *args.Realm
	}
	if args.Comment != nil {
		createReq.Comment = *args.Comment
	}

	credential, err := s.mythicClient.CreateCredential(ctx, createReq)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(credential, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully created credential for account '%s' (ID: %d)\n\n%s", credential.Account, credential.ID, string(data)),
			},
		},
	}, credential, nil
}

// handleUpdateCredential updates an existing credential
func (s *Server) handleUpdateCredential(ctx context.Context, req *mcp.CallToolRequest, args updateCredentialArgs) (*mcp.CallToolResult, any, error) {
	updateReq := &types.UpdateCredentialRequest{
		ID:         args.CredentialID,
		Type:       args.Type,
		Account:    args.Account,
		Realm:      args.Realm,
		Credential: args.Credential,
		Comment:    args.Comment,
	}

	credential, err := s.mythicClient.UpdateCredential(ctx, updateReq)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(credential, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully updated credential\n\n%s", string(data)),
			},
		},
	}, credential, nil
}

// handleDeleteCredential deletes a credential
func (s *Server) handleDeleteCredential(ctx context.Context, req *mcp.CallToolRequest, args deleteCredentialArgs) (*mcp.CallToolResult, any, error) {
	err := s.mythicClient.DeleteCredential(ctx, args.CredentialID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Successfully deleted credential %d", args.CredentialID),
				},
			},
		}, map[string]interface{}{
			"credential_id": args.CredentialID,
			"success":       true,
		}, nil
}

// Tool handlers for artifacts

// handleGetArtifacts retrieves all artifacts
func (s *Server) handleGetArtifacts(ctx context.Context, req *mcp.CallToolRequest, args getArtifactsArgs) (*mcp.CallToolResult, any, error) {
	artifacts, err := s.mythicClient.GetArtifacts(ctx)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(artifacts, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Artifacts (%d total):\n\n%s", len(artifacts), string(data)),
			},
		},
	}, wrapList(artifacts), nil
}

// handleGetArtifact retrieves a specific artifact by ID
func (s *Server) handleGetArtifact(ctx context.Context, req *mcp.CallToolRequest, args getArtifactArgs) (*mcp.CallToolResult, any, error) {
	artifact, err := s.mythicClient.GetArtifactByID(ctx, args.ArtifactID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Artifact details:\n\n%s", string(data)),
			},
		},
	}, artifact, nil
}

// handleGetOperationArtifacts retrieves artifacts for an operation
func (s *Server) handleGetOperationArtifacts(ctx context.Context, req *mcp.CallToolRequest, args getOperationArtifactsArgs) (*mcp.CallToolResult, any, error) {
	artifacts, err := s.mythicClient.GetArtifactsByOperation(ctx, args.OperationID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(artifacts, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Artifacts for operation %d (%d total):\n\n%s", args.OperationID, len(artifacts), string(data)),
			},
		},
	}, wrapList(artifacts), nil
}

// handleGetHostArtifacts retrieves artifacts for a host
func (s *Server) handleGetHostArtifacts(ctx context.Context, req *mcp.CallToolRequest, args getHostArtifactsArgs) (*mcp.CallToolResult, any, error) {
	artifacts, err := s.mythicClient.GetArtifactsByHost(ctx, args.Host)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(artifacts, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Artifacts for host '%s' (%d total):\n\n%s", args.Host, len(artifacts), string(data)),
			},
		},
	}, wrapList(artifacts), nil
}

// handleGetArtifactsByType retrieves artifacts by type
func (s *Server) handleGetArtifactsByType(ctx context.Context, req *mcp.CallToolRequest, args getArtifactsByTypeArgs) (*mcp.CallToolResult, any, error) {
	artifacts, err := s.mythicClient.GetArtifactsByType(ctx, args.ArtifactType)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(artifacts, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Artifacts of type '%s' (%d total):\n\n%s", args.ArtifactType, len(artifacts), string(data)),
			},
		},
	}, wrapList(artifacts), nil
}

// handleCreateArtifact creates a new artifact
func (s *Server) handleCreateArtifact(ctx context.Context, req *mcp.CallToolRequest, args createArtifactArgs) (*mcp.CallToolResult, any, error) {
	createReq := &types.CreateArtifactRequest{
		Artifact:     args.Artifact,
		BaseArtifact: args.BaseArtifact,
		Host:         args.Host,
		TaskID:       args.TaskID,
	}

	artifact, err := s.mythicClient.CreateArtifact(ctx, createReq)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully created artifact '%s' (ID: %d)\n\n%s", artifact.Artifact, artifact.ID, string(data)),
			},
		},
	}, artifact, nil
}

// handleUpdateArtifact updates an existing artifact
func (s *Server) handleUpdateArtifact(ctx context.Context, req *mcp.CallToolRequest, args updateArtifactArgs) (*mcp.CallToolResult, any, error) {
	updateReq := &types.UpdateArtifactRequest{
		ID:   args.ArtifactID,
		Host: args.Host,
	}

	artifact, err := s.mythicClient.UpdateArtifact(ctx, updateReq)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully updated artifact\n\n%s", string(data)),
			},
		},
	}, artifact, nil
}

// handleDeleteArtifact deletes an artifact
func (s *Server) handleDeleteArtifact(ctx context.Context, req *mcp.CallToolRequest, args deleteArtifactArgs) (*mcp.CallToolResult, any, error) {
	err := s.mythicClient.DeleteArtifact(ctx, args.ArtifactID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Successfully deleted artifact %d", args.ArtifactID),
				},
			},
		}, map[string]interface{}{
			"artifact_id": args.ArtifactID,
			"success":     true,
		}, nil
}
