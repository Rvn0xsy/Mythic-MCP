//go:build integration && e2e
// +build integration,e2e

package integration

import (
	"testing"

	"github.com/nbaertsch/mythic-sdk-go/pkg/mythic/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_Tags_TagTypes tests tag type management
func TestE2E_Tags_TagTypes(t *testing.T) {
	setup := SetupE2ETest(t)

	// Step 1: Create a new tag type
	createResult, err := setup.CallMCPTool("mythic_create_tag_type", map[string]interface{}{
		"name":        "E2E Test Tag",
		"description": "Tag type for E2E testing",
		"color":       "#FF5733",
	})
	require.NoError(t, err)
	require.NotNil(t, createResult)

	// Extract tag type ID
	metadata, ok := createResult["metadata"].(map[string]interface{})
	require.True(t, ok, "Expected metadata in result")
	tagTypeIDFloat, ok := metadata["id"].(float64)
	require.True(t, ok, "Expected tag type ID in metadata")
	tagTypeID := int(tagTypeIDFloat)

	// Step 2: Get the tag type by ID
	getResult, err := setup.CallMCPTool("mythic_get_tag_type", map[string]interface{}{
		"tag_type_id": tagTypeID,
	})
	require.NoError(t, err)
	require.NotNil(t, getResult)

	// Step 3: Update the tag type
	updateResult, err := setup.CallMCPTool("mythic_update_tag_type", map[string]interface{}{
		"tag_type_id": tagTypeID,
		"name":        "Updated E2E Tag",
		"color":       "#00FF00",
	})
	require.NoError(t, err)
	require.NotNil(t, updateResult)

	// Verify update
	tagType, err := setup.MythicClient.GetTagTypeByID(setup.Ctx, tagTypeID)
	require.NoError(t, err)
	assert.Equal(t, "Updated E2E Tag", tagType.Name)
	assert.Equal(t, "#00FF00", tagType.Color)

	// Step 4: Delete the tag type
	deleteResult, err := setup.CallMCPTool("mythic_delete_tag_type", map[string]interface{}{
		"tag_type_id": tagTypeID,
	})
	require.NoError(t, err)
	require.NotNil(t, deleteResult)
}

// TestE2E_Tags_GetTagTypes tests listing tag types
func TestE2E_Tags_GetTagTypes(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all tag types
	result, err := setup.CallMCPTool("mythic_get_tag_types", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Result should contain tag types array
	_, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	// May be empty if no tag types exist
}

// TestE2E_Tags_GetTagTypesByOperation tests filtering tag types by operation
func TestE2E_Tags_GetTagTypesByOperation(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get an operation
	operations, err := setup.MythicClient.GetOperations(setup.Ctx)
	require.NoError(t, err)
	require.NotEmpty(t, operations)
	operationID := operations[0].ID

	// Get tag types for this operation
	result, err := setup.CallMCPTool("mythic_get_tag_types_by_operation", map[string]interface{}{
		"operation_id": operationID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	_, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
}

// TestE2E_Tags_CreateAndApplyTags tests creating and applying tags to objects
func TestE2E_Tags_CreateAndApplyTags(t *testing.T) {
	setup := SetupE2ETest(t)

	// Step 1: Create a tag type
	tagType, err := setup.MythicClient.CreateTagType(setup.Ctx, &types.CreateTagTypeRequest{
		Name: "Test Apply Tag",
	})
	require.NoError(t, err)

	// Step 2: Upload a file to tag
	fileUUID, err := setup.MythicClient.UploadFile(setup.Ctx, "test-tag-file.txt", []byte("test content"))
	require.NoError(t, err)

	// Get file info to get its ID
	fileMeta, err := setup.MythicClient.GetFileByID(setup.Ctx, fileUUID)
	require.NoError(t, err)

	// Step 3: Create a tag on the file
	createTagResult, err := setup.CallMCPTool("mythic_create_tag", map[string]interface{}{
		"tag_type_id": tagType.ID,
		"source_type": "filemeta",
		"source_id":   fileMeta.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, createTagResult)

	// Extract tag ID
	tagMetadata, ok := createTagResult["metadata"].(map[string]interface{})
	require.True(t, ok)
	tagIDFloat, ok := tagMetadata["id"].(float64)
	require.True(t, ok)
	tagID := int(tagIDFloat)

	// Step 4: Get tags for the file
	getTagsResult, err := setup.CallMCPTool("mythic_get_tags", map[string]interface{}{
		"source_type": "filemeta",
		"source_id":   fileMeta.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, getTagsResult)

	// Step 5: Get tag by ID
	getTagResult, err := setup.CallMCPTool("mythic_get_tag", map[string]interface{}{
		"tag_id": tagID,
	})
	require.NoError(t, err)
	require.NotNil(t, getTagResult)

	// Step 6: Delete the tag
	deleteTagResult, err := setup.CallMCPTool("mythic_delete_tag", map[string]interface{}{
		"tag_id": tagID,
	})
	require.NoError(t, err)
	require.NotNil(t, deleteTagResult)

	// Cleanup
	setup.MythicClient.DeleteFile(setup.Ctx, fileUUID)
	setup.MythicClient.DeleteTagType(setup.Ctx, tagType.ID)
}

// TestE2E_Tags_GetTagsByOperation tests getting all tags in an operation
func TestE2E_Tags_GetTagsByOperation(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get an operation
	operations, err := setup.MythicClient.GetOperations(setup.Ctx)
	require.NoError(t, err)
	require.NotEmpty(t, operations)
	operationID := operations[0].ID

	// Get tags for this operation
	result, err := setup.CallMCPTool("mythic_get_tags_by_operation", map[string]interface{}{
		"operation_id": operationID,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	_, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
}

// TestE2E_Tags_MultipleTagTypes tests creating multiple tag types
func TestE2E_Tags_MultipleTagTypes(t *testing.T) {
	setup := SetupE2ETest(t)

	tagTypeIDs := make([]int, 0, 3)

	// Create 3 tag types
	colors := []string{"#FF0000", "#00FF00", "#0000FF"}
	for i, color := range colors {
		createResult, err := setup.CallMCPTool("mythic_create_tag_type", map[string]interface{}{
			"name":  "Multi Tag Type " + string(rune('A'+i)),
			"color": color,
		})
		require.NoError(t, err)

		metadata, ok := createResult["metadata"].(map[string]interface{})
		require.True(t, ok)
		tagTypeID := int(metadata["id"].(float64))
		tagTypeIDs = append(tagTypeIDs, tagTypeID)
	}

	// Verify all tag types exist
	allTagTypes, err := setup.MythicClient.GetTagTypes(setup.Ctx)
	require.NoError(t, err)

	for _, ttID := range tagTypeIDs {
		found := false
		for _, tt := range allTagTypes {
			if tt.ID == ttID {
				found = true
				break
			}
		}
		assert.True(t, found, "Tag type %d should exist", ttID)
	}

	// Clean up
	for _, ttID := range tagTypeIDs {
		setup.MythicClient.DeleteTagType(setup.Ctx, ttID)
	}
}

// TestE2E_Tags_ErrorHandling tests error scenarios
func TestE2E_Tags_ErrorHandling(t *testing.T) {
	setup := SetupE2ETest(t)

	// Test getting non-existent tag type
	_, err := setup.CallMCPTool("mythic_get_tag_type", map[string]interface{}{
		"tag_type_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting non-existent tag type")

	// Test getting non-existent tag
	_, err = setup.CallMCPTool("mythic_get_tag", map[string]interface{}{
		"tag_id": 999999,
	})
	assert.Error(t, err, "Expected error when getting non-existent tag")

	// Test creating tag with invalid source type
	_, err = setup.CallMCPTool("mythic_create_tag", map[string]interface{}{
		"tag_type_id": 1,
		"source_type": "invalid_source",
		"source_id":   1,
	})
	assert.Error(t, err, "Expected error with invalid source type")

	// Test creating tag type with empty name
	_, err = setup.CallMCPTool("mythic_create_tag_type", map[string]interface{}{
		"name": "",
	})
	assert.Error(t, err, "Expected error with empty tag type name")

	// Test updating non-existent tag type
	_, err = setup.CallMCPTool("mythic_update_tag_type", map[string]interface{}{
		"tag_type_id": 999999,
		"name":        "Should Fail",
	})
	assert.Error(t, err, "Expected error when updating non-existent tag type")

	// Test deleting non-existent tag type
	_, err = setup.CallMCPTool("mythic_delete_tag_type", map[string]interface{}{
		"tag_type_id": 999999,
	})
	assert.Error(t, err, "Expected error when deleting non-existent tag type")

	// Test deleting non-existent tag
	_, err = setup.CallMCPTool("mythic_delete_tag", map[string]interface{}{
		"tag_id": 999999,
	})
	assert.Error(t, err, "Expected error when deleting non-existent tag")
}

// TestE2E_Tags_FullWorkflow tests a complete tagging workflow
func TestE2E_Tags_FullWorkflow(t *testing.T) {
	setup := SetupE2ETest(t)

	// Workflow: Create tag type → Upload file → Tag file → Query tags → Clean up

	// 1. Create tag type
	createTTResult, err := setup.CallMCPTool("mythic_create_tag_type", map[string]interface{}{
		"name":        "Workflow Tag",
		"description": "Full workflow test tag",
		"color":       "#FFAA00",
	})
	require.NoError(t, err)
	tagTypeMeta, ok := createTTResult["metadata"].(map[string]interface{})
	require.True(t, ok)
	tagTypeID := int(tagTypeMeta["id"].(float64))

	// 2. Upload a file
	fileUUID, err := setup.MythicClient.UploadFile(setup.Ctx, "workflow-test.txt", []byte("workflow test"))
	require.NoError(t, err)
	fileMeta, err := setup.MythicClient.GetFileByID(setup.Ctx, fileUUID)
	require.NoError(t, err)

	// 3. Apply tag to file
	createTagResult, err := setup.CallMCPTool("mythic_create_tag", map[string]interface{}{
		"tag_type_id": tagTypeID,
		"source_type": "filemeta",
		"source_id":   fileMeta.ID,
	})
	require.NoError(t, err)
	tagMeta, ok := createTagResult["metadata"].(map[string]interface{})
	require.True(t, ok)
	tagID := int(tagMeta["id"].(float64))

	// 4. Get tags for the file
	getTagsResult, err := setup.CallMCPTool("mythic_get_tags", map[string]interface{}{
		"source_type": "filemeta",
		"source_id":   fileMeta.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, getTagsResult)

	// 5. Get tag by ID
	getTagResult, err := setup.CallMCPTool("mythic_get_tag", map[string]interface{}{
		"tag_id": tagID,
	})
	require.NoError(t, err)
	require.NotNil(t, getTagResult)

	// 6. Get current operation's tags
	currentOp, err := setup.MythicClient.GetMe(setup.Ctx)
	require.NoError(t, err)
	if currentOp.CurrentOperation != nil {
		opTagsResult, err := setup.CallMCPTool("mythic_get_tags_by_operation", map[string]interface{}{
			"operation_id": currentOp.CurrentOperation.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, opTagsResult)
	}

	// 7. Update tag type
	_, err = setup.CallMCPTool("mythic_update_tag_type", map[string]interface{}{
		"tag_type_id": tagTypeID,
		"name":        "Updated Workflow Tag",
	})
	require.NoError(t, err)

	// 8. Verify update
	updatedTT, err := setup.MythicClient.GetTagTypeByID(setup.Ctx, tagTypeID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Workflow Tag", updatedTT.Name)

	// 9. Clean up: delete tag, tag type, and file
	_, err = setup.CallMCPTool("mythic_delete_tag", map[string]interface{}{
		"tag_id": tagID,
	})
	require.NoError(t, err)

	_, err = setup.CallMCPTool("mythic_delete_tag_type", map[string]interface{}{
		"tag_type_id": tagTypeID,
	})
	require.NoError(t, err)

	setup.MythicClient.DeleteFile(setup.Ctx, fileUUID)
}
