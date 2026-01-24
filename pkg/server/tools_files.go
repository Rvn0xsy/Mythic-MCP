package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerFilesTools registers file operation MCP tools
func (s *Server) registerFilesTools() {
	// mythic_get_files - List all files
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_files",
		Description: "Get a list of all files in Mythic (uploaded and downloaded)",
	}, s.handleGetFiles)

	// mythic_get_file - Get specific file info
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_file",
		Description: "Get metadata and information about a specific file by its UUID",
	}, s.handleGetFile)

	// mythic_get_downloaded_files - List downloaded files
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_downloaded_files",
		Description: "Get a list of files downloaded from agents",
	}, s.handleGetDownloadedFiles)

	// mythic_upload_file - Upload a file
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_upload_file",
		Description: "Upload a file to Mythic for later use (tasking, payload building, etc.)",
	}, s.handleUploadFile)

	// mythic_download_file - Download a file
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_download_file",
		Description: "Download a file's content from Mythic by its UUID",
	}, s.handleDownloadFile)

	// mythic_delete_file - Delete a file
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_delete_file",
		Description: "Delete a file from Mythic by its UUID",
	}, s.handleDeleteFile)

	// mythic_bulk_download_files - Download multiple files as ZIP
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_bulk_download_files",
		Description: "Download multiple files as a single ZIP archive",
	}, s.handleBulkDownloadFiles)

	// mythic_preview_file - Preview file content
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_preview_file",
		Description: "Preview a file's content (for text files, limited size)",
	}, s.handlePreviewFile)
}

// Tool argument types for file tools

type getFilesArgs struct {
	Limit int `json:"limit,omitempty" jsonschema:"description=Maximum number of files to return (default 100)"`
}

type getFileArgs struct {
	FileID string `json:"file_id" jsonschema:"required,description=UUID of the file to retrieve"`
}

type getDownloadedFilesArgs struct {
	Limit int `json:"limit,omitempty" jsonschema:"description=Maximum number of files to return (default 100)"`
}

type uploadFileArgs struct {
	Filename string `json:"filename" jsonschema:"required,description=Name of the file"`
	FileData string `json:"file_data" jsonschema:"required,description=Base64-encoded file content"`
}

type downloadFileArgs struct {
	FileUUID string `json:"file_uuid" jsonschema:"required,description=UUID of the file to download"`
}

type deleteFileArgs struct {
	FileID string `json:"file_id" jsonschema:"required,description=UUID of the file to delete"`
}

type bulkDownloadFilesArgs struct {
	FileUUIDs []string `json:"file_uuids" jsonschema:"required,description=Array of file UUIDs to download"`
}

type previewFileArgs struct {
	FileID string `json:"file_id" jsonschema:"required,description=UUID of the file to preview"`
}

// Tool handlers

// handleGetFiles retrieves all files
func (s *Server) handleGetFiles(ctx context.Context, req *mcp.CallToolRequest, args getFilesArgs) (*mcp.CallToolResult, any, error) {
	limit := args.Limit
	if limit <= 0 {
		limit = 100 // Default limit
	}

	files, err := s.mythicClient.GetFiles(ctx, limit)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Files (limit %d):\n\n%s", limit, string(data)),
			},
		},
	}, files, nil
}

// handleGetFile retrieves a specific file's metadata
func (s *Server) handleGetFile(ctx context.Context, req *mcp.CallToolRequest, args getFileArgs) (*mcp.CallToolResult, any, error) {
	file, err := s.mythicClient.GetFileByID(ctx, args.FileID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("File info for %s:\n\n%s", args.FileID, string(data)),
			},
		},
	}, file, nil
}

// handleGetDownloadedFiles retrieves downloaded files
func (s *Server) handleGetDownloadedFiles(ctx context.Context, req *mcp.CallToolRequest, args getDownloadedFilesArgs) (*mcp.CallToolResult, any, error) {
	limit := args.Limit
	if limit <= 0 {
		limit = 100 // Default limit
	}

	files, err := s.mythicClient.GetDownloadedFiles(ctx, limit)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Downloaded files (limit %d):\n\n%s", limit, string(data)),
			},
		},
	}, files, nil
}

// handleUploadFile uploads a file to Mythic
func (s *Server) handleUploadFile(ctx context.Context, req *mcp.CallToolRequest, args uploadFileArgs) (*mcp.CallToolResult, any, error) {
	// Decode base64 file data
	fileData, err := base64.StdEncoding.DecodeString(args.FileData)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid base64 file data: %w", err)
	}

	// Upload file
	agentFileID, err := s.mythicClient.UploadFile(ctx, args.Filename, fileData)
	if err != nil {
		return nil, nil, translateError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully uploaded file '%s'\nFile UUID: %s\nSize: %d bytes", args.Filename, agentFileID, len(fileData)),
			},
		},
	}, map[string]interface{}{
		"agent_file_id": agentFileID,
		"filename":      args.Filename,
		"size":          len(fileData),
	}, nil
}

// handleDownloadFile downloads a file from Mythic
func (s *Server) handleDownloadFile(ctx context.Context, req *mcp.CallToolRequest, args downloadFileArgs) (*mcp.CallToolResult, any, error) {
	fileData, err := s.mythicClient.DownloadFile(ctx, args.FileUUID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	// Encode file data as base64 for transport
	encodedData := base64.StdEncoding.EncodeToString(fileData)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully downloaded file %s\nSize: %d bytes\n\nFile data is base64-encoded in the metadata.", args.FileUUID, len(fileData)),
			},
		},
	}, map[string]interface{}{
		"file_uuid": args.FileUUID,
		"size":      len(fileData),
		"file_data": encodedData,
	}, nil
}

// handleDeleteFile deletes a file from Mythic
func (s *Server) handleDeleteFile(ctx context.Context, req *mcp.CallToolRequest, args deleteFileArgs) (*mcp.CallToolResult, any, error) {
	err := s.mythicClient.DeleteFile(ctx, args.FileID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully deleted file %s", args.FileID),
			},
		},
	}, map[string]interface{}{
		"file_id": args.FileID,
		"success": true,
	}, nil
}

// handleBulkDownloadFiles downloads multiple files as a ZIP
func (s *Server) handleBulkDownloadFiles(ctx context.Context, req *mcp.CallToolRequest, args bulkDownloadFilesArgs) (*mcp.CallToolResult, any, error) {
	if len(args.FileUUIDs) == 0 {
		return nil, nil, fmt.Errorf("at least one file UUID is required")
	}

	// BulkDownloadFiles returns a download URL or path
	zipURL, err := s.mythicClient.BulkDownloadFiles(ctx, args.FileUUIDs)
	if err != nil {
		return nil, nil, translateError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully created bulk download for %d files\nDownload URL: %s", len(args.FileUUIDs), zipURL),
			},
		},
	}, map[string]interface{}{
		"zip_url":    zipURL,
		"file_count": len(args.FileUUIDs),
	}, nil
}

// handlePreviewFile previews a file's content
func (s *Server) handlePreviewFile(ctx context.Context, req *mcp.CallToolRequest, args previewFileArgs) (*mcp.CallToolResult, any, error) {
	preview, err := s.mythicClient.PreviewFile(ctx, args.FileID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(preview, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("File preview for %s:\n\n%s", args.FileID, string(data)),
			},
		},
	}, preview, nil
}
