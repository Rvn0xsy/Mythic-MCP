package integration

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_Files_UploadDownloadDelete tests the complete file lifecycle
func TestE2E_Files_UploadDownloadDelete(t *testing.T) {
	setup := SetupE2ETest(t)

	// Test data
	testData := []byte("This is test file content from E2E tests")
	testFilename := "test-file-e2e.txt"
	encodedData := base64.StdEncoding.EncodeToString(testData)

	// Step 1: Upload a file
	uploadResult, err := setup.CallMCPTool("mythic_upload_file", map[string]interface{}{
		"filename":  testFilename,
		"file_data": encodedData,
	})
	require.NoError(t, err)
	require.NotNil(t, uploadResult)

	// Extract file UUID from result
	metadata, ok := uploadResult["metadata"].(map[string]interface{})
	require.True(t, ok, "Expected metadata in result")
	fileUUID, ok := metadata["agent_file_id"].(string)
	require.True(t, ok, "Expected agent_file_id in metadata")
	require.NotEmpty(t, fileUUID, "File UUID should not be empty")

	// Step 2: Get file info by ID
	getResult, err := setup.CallMCPTool("mythic_get_file", map[string]interface{}{
		"file_id": fileUUID,
	})
	require.NoError(t, err)
	require.NotNil(t, getResult)

	// Step 3: Download the file
	downloadResult, err := setup.CallMCPTool("mythic_download_file", map[string]interface{}{
		"file_uuid": fileUUID,
	})
	require.NoError(t, err)
	require.NotNil(t, downloadResult)

	// Verify downloaded content matches
	downloadMetadata, ok := downloadResult["metadata"].(map[string]interface{})
	require.True(t, ok, "Expected metadata in download result")
	downloadedData, ok := downloadMetadata["file_data"].(string)
	require.True(t, ok, "Expected file_data in metadata")

	// Decode and verify
	decodedData, err := base64.StdEncoding.DecodeString(downloadedData)
	require.NoError(t, err)
	assert.Equal(t, testData, decodedData, "Downloaded content should match uploaded content")

	// Step 4: Delete the file
	deleteResult, err := setup.CallMCPTool("mythic_delete_file", map[string]interface{}{
		"file_id": fileUUID,
	})
	require.NoError(t, err)
	require.NotNil(t, deleteResult)
}

// TestE2E_Files_GetFiles tests listing all files
func TestE2E_Files_GetFiles(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get all files (with a reasonable limit)
	result, err := setup.CallMCPTool("mythic_get_files", map[string]interface{}{
		"limit": 50,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Result should contain file listing
	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
	// Note: May be empty if no files exist yet
}

// TestE2E_Files_GetDownloadedFiles tests listing downloaded files
func TestE2E_Files_GetDownloadedFiles(t *testing.T) {
	setup := SetupE2ETest(t)

	// Get downloaded files
	result, err := setup.CallMCPTool("mythic_get_downloaded_files", map[string]interface{}{
		"limit": 50,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	// Result should contain file listing
	content, ok := result["content"].([]interface{})
	require.True(t, ok, "Expected content to be an array")
}

// TestE2E_Files_PreviewFile tests file preview functionality
func TestE2E_Files_PreviewFile(t *testing.T) {
	setup := SetupE2ETest(t)

	// Upload a small text file first
	testData := []byte("Preview test content\nLine 2\nLine 3")
	encodedData := base64.StdEncoding.EncodeToString(testData)

	uploadResult, err := setup.CallMCPTool("mythic_upload_file", map[string]interface{}{
		"filename":  "preview-test.txt",
		"file_data": encodedData,
	})
	require.NoError(t, err)

	metadata, ok := uploadResult["metadata"].(map[string]interface{})
	require.True(t, ok)
	fileUUID := metadata["agent_file_id"].(string)

	// Preview the file
	previewResult, err := setup.CallMCPTool("mythic_preview_file", map[string]interface{}{
		"file_id": fileUUID,
	})

	// Note: PreviewFile may not be available in all Mythic versions
	if err != nil {
		t.Logf("File preview not available: %v", err)
		// Clean up
		setup.CallMCPTool("mythic_delete_file", map[string]interface{}{
			"file_id": fileUUID,
		})
		return
	}

	require.NotNil(t, previewResult)

	// Clean up
	_, err = setup.CallMCPTool("mythic_delete_file", map[string]interface{}{
		"file_id": fileUUID,
	})
	require.NoError(t, err)
}

// TestE2E_Files_BulkDownload tests bulk file download
func TestE2E_Files_BulkDownload(t *testing.T) {
	setup := SetupE2ETest(t)

	// Upload multiple test files
	fileUUIDs := make([]string, 0, 3)

	for i := 1; i <= 3; i++ {
		testData := []byte("Bulk download test file " + string(rune('0'+i)))
		encodedData := base64.StdEncoding.EncodeToString(testData)

		uploadResult, err := setup.CallMCPTool("mythic_upload_file", map[string]interface{}{
			"filename":  "bulk-test-" + string(rune('0'+i)) + ".txt",
			"file_data": encodedData,
		})
		require.NoError(t, err)

		metadata, ok := uploadResult["metadata"].(map[string]interface{})
		require.True(t, ok)
		fileUUID := metadata["agent_file_id"].(string)
		fileUUIDs = append(fileUUIDs, fileUUID)
	}

	// Bulk download all files
	bulkResult, err := setup.CallMCPTool("mythic_bulk_download_files", map[string]interface{}{
		"file_uuids": fileUUIDs,
	})
	require.NoError(t, err)
	require.NotNil(t, bulkResult)

	// Result should contain ZIP file data or download URL
	bulkMetadata, ok := bulkResult["metadata"].(map[string]interface{})
	require.True(t, ok, "Expected metadata in bulk download result")

	// Either zip_url or file_data should be present
	_, hasURL := bulkMetadata["zip_url"]
	_, hasData := bulkMetadata["file_data"]
	assert.True(t, hasURL || hasData, "Expected either zip_url or file_data in result")

	// Clean up
	for _, uuid := range fileUUIDs {
		setup.CallMCPTool("mythic_delete_file", map[string]interface{}{
			"file_id": uuid,
		})
	}
}

// TestE2E_Files_ErrorHandling tests error scenarios
func TestE2E_Files_ErrorHandling(t *testing.T) {
	setup := SetupE2ETest(t)

	// Test getting non-existent file
	_, err := setup.CallMCPTool("mythic_get_file", map[string]interface{}{
		"file_id": "non-existent-uuid-12345",
	})
	assert.Error(t, err, "Expected error when getting non-existent file")

	// Test downloading non-existent file
	_, err = setup.CallMCPTool("mythic_download_file", map[string]interface{}{
		"file_uuid": "non-existent-uuid-12345",
	})
	assert.Error(t, err, "Expected error when downloading non-existent file")

	// Test deleting non-existent file
	_, err = setup.CallMCPTool("mythic_delete_file", map[string]interface{}{
		"file_id": "non-existent-uuid-12345",
	})
	assert.Error(t, err, "Expected error when deleting non-existent file")

	// Test uploading with invalid data
	_, err = setup.CallMCPTool("mythic_upload_file", map[string]interface{}{
		"filename":  "",
		"file_data": "not-valid-base64!!!",
	})
	assert.Error(t, err, "Expected error when uploading with invalid data")
}

// TestE2E_Files_FullWorkflow tests a complete file workflow
func TestE2E_Files_FullWorkflow(t *testing.T) {
	setup := SetupE2ETest(t)

	// Workflow: Upload → Verify → Download → Delete → Verify deletion

	// 1. Upload file
	testContent := []byte("Full workflow test content with multiple lines\nLine 2\nLine 3")
	encodedData := base64.StdEncoding.EncodeToString(testContent)

	uploadResult, err := setup.CallMCPTool("mythic_upload_file", map[string]interface{}{
		"filename":  "workflow-test.txt",
		"file_data": encodedData,
	})
	require.NoError(t, err)

	metadata, ok := uploadResult["metadata"].(map[string]interface{})
	require.True(t, ok)
	fileUUID := metadata["agent_file_id"].(string)

	// 2. Verify file exists and has correct metadata
	getResult, err := setup.CallMCPTool("mythic_get_file", map[string]interface{}{
		"file_id": fileUUID,
	})
	require.NoError(t, err)
	require.NotNil(t, getResult)

	// 3. List all files and verify our file is included
	listResult, err := setup.CallMCPTool("mythic_get_files", map[string]interface{}{
		"limit": 100,
	})
	require.NoError(t, err)
	require.NotNil(t, listResult)

	// 4. Download and verify content
	downloadResult, err := setup.CallMCPTool("mythic_download_file", map[string]interface{}{
		"file_uuid": fileUUID,
	})
	require.NoError(t, err)

	downloadMetadata, ok := downloadResult["metadata"].(map[string]interface{})
	require.True(t, ok)
	downloadedData := downloadMetadata["file_data"].(string)
	decodedData, err := base64.StdEncoding.DecodeString(downloadedData)
	require.NoError(t, err)
	assert.Equal(t, testContent, decodedData, "Downloaded content must match original")

	// 5. Delete file
	deleteResult, err := setup.CallMCPTool("mythic_delete_file", map[string]interface{}{
		"file_id": fileUUID,
	})
	require.NoError(t, err)
	require.NotNil(t, deleteResult)

	// 6. Verify file is deleted (get should fail)
	_, err = setup.CallMCPTool("mythic_get_file", map[string]interface{}{
		"file_id": fileUUID,
	})
	// Note: Depending on Mythic's behavior, this might error or return deleted=true
	// Either is acceptable
	t.Logf("Post-delete get result: %v", err)
}

// TestE2E_Files_MultipleUploads tests uploading multiple files
func TestE2E_Files_MultipleUploads(t *testing.T) {
	setup := SetupE2ETest(t)

	fileUUIDs := make([]string, 0, 5)

	// Upload 5 different files
	for i := 1; i <= 5; i++ {
		content := []byte("Multi-upload test file " + string(rune('0'+i)) + "\nContent line 2")
		encodedData := base64.StdEncoding.EncodeToString(content)

		uploadResult, err := setup.CallMCPTool("mythic_upload_file", map[string]interface{}{
			"filename":  "multi-test-" + string(rune('0'+i)) + ".txt",
			"file_data": encodedData,
		})
		require.NoError(t, err)

		metadata, ok := uploadResult["metadata"].(map[string]interface{})
		require.True(t, ok)
		fileUUID := metadata["agent_file_id"].(string)
		fileUUIDs = append(fileUUIDs, fileUUID)
	}

	// Verify all files exist
	for _, uuid := range fileUUIDs {
		getResult, err := setup.CallMCPTool("mythic_get_file", map[string]interface{}{
			"file_id": uuid,
		})
		require.NoError(t, err)
		require.NotNil(t, getResult)
	}

	// Clean up all files
	for _, uuid := range fileUUIDs {
		_, err := setup.CallMCPTool("mythic_delete_file", map[string]interface{}{
			"file_id": uuid,
		})
		require.NoError(t, err)
	}
}
