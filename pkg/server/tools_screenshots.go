package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerScreenshotsTools registers screenshot management MCP tools
func (s *Server) registerScreenshotsTools() {
	// mythic_get_screenshots - List screenshots for a callback
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_screenshots",
		Description: "Get screenshots captured by a specific callback",
	}, s.handleGetScreenshots)

	// mythic_get_screenshot_by_id - Get specific screenshot
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_screenshot_by_id",
		Description: "Get detailed information about a specific screenshot by its ID",
	}, s.handleGetScreenshotByID)

	// mythic_get_screenshot_timeline - Get screenshots in time range
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_screenshot_timeline",
		Description: "Get screenshots from a callback within a specific time range",
	}, s.handleGetScreenshotTimeline)

	// mythic_get_screenshot_thumbnail - Download screenshot thumbnail
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_screenshot_thumbnail",
		Description: "Download thumbnail image of a screenshot",
	}, s.handleGetScreenshotThumbnail)

	// mythic_download_screenshot - Download full screenshot
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_download_screenshot",
		Description: "Download the full resolution screenshot image",
	}, s.handleDownloadScreenshot)

	// mythic_delete_screenshot - Delete a screenshot
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_delete_screenshot",
		Description: "Delete a screenshot from Mythic (destructive operation)",
	}, s.handleDeleteScreenshot)
}

// Tool argument types for screenshot tools

type getScreenshotsArgs struct {
	CallbackID int `json:"callback_id" jsonschema:"required,description=Display ID of the callback"`
	Limit      int `json:"limit" jsonschema:"required,description=Maximum number of screenshots to retrieve"`
}

type getScreenshotByIDArgs struct {
	ScreenshotID int `json:"screenshot_id" jsonschema:"required,description=ID of the screenshot"`
}

type getScreenshotTimelineArgs struct {
	CallbackID int    `json:"callback_id" jsonschema:"required,description=Display ID of the callback"`
	StartTime  string `json:"start_time" jsonschema:"required,description=Start time in RFC3339 format"`
	EndTime    string `json:"end_time" jsonschema:"required,description=End time in RFC3339 format"`
}

type getScreenshotThumbnailArgs struct {
	AgentFileID string `json:"agent_file_id" jsonschema:"required,description=Agent file ID of the screenshot"`
}

type downloadScreenshotArgs struct {
	AgentFileID string `json:"agent_file_id" jsonschema:"required,description=Agent file ID of the screenshot"`
}

type deleteScreenshotArgs struct {
	AgentFileID string `json:"agent_file_id" jsonschema:"required,description=Agent file ID of the screenshot to delete"`
}

// Tool handlers

// handleGetScreenshots retrieves screenshots for a callback
func (s *Server) handleGetScreenshots(ctx context.Context, req *mcp.CallToolRequest, args getScreenshotsArgs) (*mcp.CallToolResult, any, error) {
	screenshots, err := s.mythicClient.GetScreenshots(ctx, args.CallbackID, args.Limit)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(screenshots, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	// Calculate total size
	var totalSize int64
	for _, screenshot := range screenshots {
		totalSize += screenshot.Size
	}

	summary := fmt.Sprintf("Screenshots for callback %d (%d total, %s):\n\n",
		args.CallbackID, len(screenshots), formatBytes(totalSize))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("%s%s", summary, string(data)),
			},
		},
	}, screenshots, nil
}

// handleGetScreenshotByID retrieves a specific screenshot
func (s *Server) handleGetScreenshotByID(ctx context.Context, req *mcp.CallToolRequest, args getScreenshotByIDArgs) (*mcp.CallToolResult, any, error) {
	screenshot, err := s.mythicClient.GetScreenshotByID(ctx, args.ScreenshotID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(screenshot, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Screenshot %d:\nFilename: %s\nSize: %s\nAgent File ID: %s\n\n%s",
					screenshot.ID, screenshot.Filename, formatBytes(screenshot.Size),
					screenshot.AgentFileID, string(data)),
			},
		},
	}, screenshot, nil
}

// handleGetScreenshotTimeline retrieves screenshots in a time range
func (s *Server) handleGetScreenshotTimeline(ctx context.Context, req *mcp.CallToolRequest, args getScreenshotTimelineArgs) (*mcp.CallToolResult, any, error) {
	// Parse times
	startTime, err := time.Parse(time.RFC3339, args.StartTime)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid start_time format: %w", err)
	}

	endTime, err := time.Parse(time.RFC3339, args.EndTime)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid end_time format: %w", err)
	}

	screenshots, err := s.mythicClient.GetScreenshotTimeline(ctx, args.CallbackID, &startTime, &endTime)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(screenshots, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	summary := fmt.Sprintf("Screenshots for callback %d from %s to %s (%d total):\n\n",
		args.CallbackID, startTime.Format("2006-01-02 15:04"), endTime.Format("2006-01-02 15:04"), len(screenshots))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("%s%s", summary, string(data)),
			},
		},
	}, screenshots, nil
}

// handleGetScreenshotThumbnail downloads a screenshot thumbnail
func (s *Server) handleGetScreenshotThumbnail(ctx context.Context, req *mcp.CallToolRequest, args getScreenshotThumbnailArgs) (*mcp.CallToolResult, any, error) {
	thumbnailData, err := s.mythicClient.GetScreenshotThumbnail(ctx, args.AgentFileID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	// Encode as base64 for transmission
	encodedData := base64.StdEncoding.EncodeToString(thumbnailData)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Screenshot thumbnail for %s (%s):\n\nBase64 encoded data:\n%s",
					args.AgentFileID, formatBytes(int64(len(thumbnailData))), encodedData),
			},
		},
	}, map[string]interface{}{
		"agent_file_id": args.AgentFileID,
		"size":          len(thumbnailData),
		"base64_data":   encodedData,
	}, nil
}

// handleDownloadScreenshot downloads a full screenshot
func (s *Server) handleDownloadScreenshot(ctx context.Context, req *mcp.CallToolRequest, args downloadScreenshotArgs) (*mcp.CallToolResult, any, error) {
	screenshotData, err := s.mythicClient.DownloadScreenshot(ctx, args.AgentFileID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	// Encode as base64 for transmission
	encodedData := base64.StdEncoding.EncodeToString(screenshotData)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Downloaded screenshot %s (%s):\n\nBase64 encoded data:\n%s",
					args.AgentFileID, formatBytes(int64(len(screenshotData))), encodedData),
			},
		},
	}, map[string]interface{}{
		"agent_file_id": args.AgentFileID,
		"size":          len(screenshotData),
		"base64_data":   encodedData,
	}, nil
}

// handleDeleteScreenshot deletes a screenshot
func (s *Server) handleDeleteScreenshot(ctx context.Context, req *mcp.CallToolRequest, args deleteScreenshotArgs) (*mcp.CallToolResult, any, error) {
	err := s.mythicClient.DeleteScreenshot(ctx, args.AgentFileID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully deleted screenshot %s", args.AgentFileID),
			},
		},
	}, map[string]interface{}{
		"agent_file_id": args.AgentFileID,
		"deleted":       true,
	}, nil
}

// formatBytes formats bytes into human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
