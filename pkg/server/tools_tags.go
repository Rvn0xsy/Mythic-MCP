package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nbaertsch/mythic-sdk-go/pkg/mythic/types"
)

// registerTagsTools registers tag management MCP tools
func (s *Server) registerTagsTools() {
	// mythic_get_tag_types - List all tag types
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_tag_types",
		Description: "Get a list of all tag types (categories for tags)",
	}, s.handleGetTagTypes)

	// mythic_get_tag_types_by_operation - Get tag types for operation
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_tag_types_by_operation",
		Description: "Get tag types filtered by operation",
	}, s.handleGetTagTypesByOperation)

	// mythic_get_tag_type - Get specific tag type
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_tag_type",
		Description: "Get details of a specific tag type by ID",
	}, s.handleGetTagType)

	// mythic_create_tag_type - Create new tag type
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_create_tag_type",
		Description: "Create a new tag type (category)",
	}, s.handleCreateTagType)

	// mythic_update_tag_type - Update existing tag type
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_update_tag_type",
		Description: "Update an existing tag type's properties",
	}, s.handleUpdateTagType)

	// mythic_delete_tag_type - Delete tag type
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_delete_tag_type",
		Description: "Delete a tag type",
	}, s.handleDeleteTagType)

	// mythic_create_tag - Create tag on object
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_create_tag",
		Description: "Create a tag and apply it to an object (task, callback, file, etc.)",
	}, s.handleCreateTag)

	// mythic_get_tag - Get specific tag
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_tag",
		Description: "Get details of a specific tag by ID",
	}, s.handleGetTag)

	// mythic_get_tags - Get tags for object
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_tags",
		Description: "Get all tags for a specific object",
	}, s.handleGetTags)

	// mythic_get_tags_by_operation - Get tags in operation
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_get_tags_by_operation",
		Description: "Get all tags in an operation",
	}, s.handleGetTagsByOperation)

	// mythic_delete_tag - Delete tag
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mythic_delete_tag",
		Description: "Delete a tag from an object",
	}, s.handleDeleteTag)
}

// Tool argument types for tags tools

type getTagTypesArgs struct{}

type getTagTypesByOperationArgs struct {
	OperationID int `json:"operation_id" jsonschema:"required,description=Operation ID to filter tag types"`
}

type getTagTypeArgs struct {
	TagTypeID int `json:"tag_type_id" jsonschema:"required,description=ID of the tag type to retrieve"`
}

type createTagTypeArgs struct {
	Name        string  `json:"name" jsonschema:"required,description=Name of the tag type"`
	Description *string `json:"description,omitempty" jsonschema:"description=Description of the tag type"`
	Color       *string `json:"color,omitempty" jsonschema:"description=Hex color code (e.g. #FF5733)"`
}

type updateTagTypeArgs struct {
	TagTypeID   int     `json:"tag_type_id" jsonschema:"required,description=ID of the tag type to update"`
	Name        *string `json:"name,omitempty" jsonschema:"description=New name for the tag type"`
	Description *string `json:"description,omitempty" jsonschema:"description=New description"`
	Color       *string `json:"color,omitempty" jsonschema:"description=New hex color code"`
	Deleted     *bool   `json:"deleted,omitempty" jsonschema:"description=Mark as deleted"`
}

type deleteTagTypeArgs struct {
	TagTypeID int `json:"tag_type_id" jsonschema:"required,description=ID of the tag type to delete"`
}

type createTagArgs struct {
	TagTypeID  int    `json:"tag_type_id" jsonschema:"required,description=ID of the tag type to use"`
	SourceType string `json:"source_type" jsonschema:"required,description=Type of object to tag (task/callback/filemeta/payload/artifact/process/keylog)"`
	SourceID   int    `json:"source_id" jsonschema:"required,description=ID of the object to tag"`
}

type getTagArgs struct {
	TagID int `json:"tag_id" jsonschema:"required,description=ID of the tag to retrieve"`
}

type getTagsArgs struct {
	SourceType string `json:"source_type" jsonschema:"required,description=Type of object (task/callback/filemeta/etc.)"`
	SourceID   int    `json:"source_id" jsonschema:"required,description=ID of the object"`
}

type getTagsByOperationArgs struct {
	OperationID int `json:"operation_id" jsonschema:"required,description=Operation ID to get tags for"`
}

type deleteTagArgs struct {
	TagID int `json:"tag_id" jsonschema:"required,description=ID of the tag to delete"`
}

// Tool handlers

// handleGetTagTypes retrieves all tag types
func (s *Server) handleGetTagTypes(ctx context.Context, req *mcp.CallToolRequest, args getTagTypesArgs) (*mcp.CallToolResult, any, error) {
	tagTypes, err := s.mythicClient.GetTagTypes(ctx)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(tagTypes, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Tag types (%d total):\n\n%s", len(tagTypes), string(data)),
			},
		},
	}, tagTypes, nil
}

// handleGetTagTypesByOperation retrieves tag types for an operation
func (s *Server) handleGetTagTypesByOperation(ctx context.Context, req *mcp.CallToolRequest, args getTagTypesByOperationArgs) (*mcp.CallToolResult, any, error) {
	tagTypes, err := s.mythicClient.GetTagTypesByOperation(ctx, args.OperationID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(tagTypes, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Tag types for operation %d (%d total):\n\n%s", args.OperationID, len(tagTypes), string(data)),
			},
		},
	}, tagTypes, nil
}

// handleGetTagType retrieves a specific tag type by ID
func (s *Server) handleGetTagType(ctx context.Context, req *mcp.CallToolRequest, args getTagTypeArgs) (*mcp.CallToolResult, any, error) {
	tagType, err := s.mythicClient.GetTagTypeByID(ctx, args.TagTypeID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(tagType, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Tag type details:\n\n%s", string(data)),
			},
		},
	}, tagType, nil
}

// handleCreateTagType creates a new tag type
func (s *Server) handleCreateTagType(ctx context.Context, req *mcp.CallToolRequest, args createTagTypeArgs) (*mcp.CallToolResult, any, error) {
	createReq := &types.CreateTagTypeRequest{
		Name:        args.Name,
		Description: args.Description,
		Color:       args.Color,
	}

	tagType, err := s.mythicClient.CreateTagType(ctx, createReq)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(tagType, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully created tag type '%s' (ID: %d)\n\n%s", tagType.Name, tagType.ID, string(data)),
			},
		},
	}, tagType, nil
}

// handleUpdateTagType updates an existing tag type
func (s *Server) handleUpdateTagType(ctx context.Context, req *mcp.CallToolRequest, args updateTagTypeArgs) (*mcp.CallToolResult, any, error) {
	updateReq := &types.UpdateTagTypeRequest{
		ID:          args.TagTypeID,
		Name:        args.Name,
		Description: args.Description,
		Color:       args.Color,
		Deleted:     args.Deleted,
	}

	tagType, err := s.mythicClient.UpdateTagType(ctx, updateReq)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(tagType, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully updated tag type '%s'\n\n%s", tagType.Name, string(data)),
			},
		},
	}, tagType, nil
}

// handleDeleteTagType deletes a tag type
func (s *Server) handleDeleteTagType(ctx context.Context, req *mcp.CallToolRequest, args deleteTagTypeArgs) (*mcp.CallToolResult, any, error) {
	err := s.mythicClient.DeleteTagType(ctx, args.TagTypeID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully deleted tag type %d", args.TagTypeID),
			},
		},
	}, map[string]interface{}{
		"tag_type_id": args.TagTypeID,
		"success":     true,
	}, nil
}

// handleCreateTag creates a tag on an object
func (s *Server) handleCreateTag(ctx context.Context, req *mcp.CallToolRequest, args createTagArgs) (*mcp.CallToolResult, any, error) {
	createReq := &types.CreateTagRequest{
		TagTypeID:  args.TagTypeID,
		SourceType: args.SourceType,
		SourceID:   args.SourceID,
	}

	tag, err := s.mythicClient.CreateTag(ctx, createReq)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(tag, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully created tag on %s %d (Tag ID: %d)\n\n%s", args.SourceType, args.SourceID, tag.ID, string(data)),
			},
		},
	}, tag, nil
}

// handleGetTag retrieves a specific tag by ID
func (s *Server) handleGetTag(ctx context.Context, req *mcp.CallToolRequest, args getTagArgs) (*mcp.CallToolResult, any, error) {
	tag, err := s.mythicClient.GetTagByID(ctx, args.TagID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(tag, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Tag details:\n\n%s", string(data)),
			},
		},
	}, tag, nil
}

// handleGetTags retrieves all tags for a specific object
func (s *Server) handleGetTags(ctx context.Context, req *mcp.CallToolRequest, args getTagsArgs) (*mcp.CallToolResult, any, error) {
	tags, err := s.mythicClient.GetTags(ctx, args.SourceType, args.SourceID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(tags, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Tags for %s %d (%d total):\n\n%s", args.SourceType, args.SourceID, len(tags), string(data)),
			},
		},
	}, tags, nil
}

// handleGetTagsByOperation retrieves all tags in an operation
func (s *Server) handleGetTagsByOperation(ctx context.Context, req *mcp.CallToolRequest, args getTagsByOperationArgs) (*mcp.CallToolResult, any, error) {
	tags, err := s.mythicClient.GetTagsByOperation(ctx, args.OperationID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	data, err := json.MarshalIndent(tags, "", "  ")
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Tags in operation %d (%d total):\n\n%s", args.OperationID, len(tags), string(data)),
			},
		},
	}, tags, nil
}

// handleDeleteTag deletes a tag
func (s *Server) handleDeleteTag(ctx context.Context, req *mcp.CallToolRequest, args deleteTagArgs) (*mcp.CallToolResult, any, error) {
	err := s.mythicClient.DeleteTag(ctx, args.TagID)
	if err != nil {
		return nil, nil, translateError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Successfully deleted tag %d", args.TagID),
			},
		},
	}, map[string]interface{}{
		"tag_id":  args.TagID,
		"success": true,
	}, nil
}
