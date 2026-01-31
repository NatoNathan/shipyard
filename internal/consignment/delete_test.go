package consignment

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeleteConsignment_Success tests deleting a single consignment file
func TestDeleteConsignment_Success(t *testing.T) {
	// Setup: Create temp consignment file
	tempDir := t.TempDir()
	consignmentPath := filepath.Join(tempDir, "test-consignment.md")
	require.NoError(t, os.WriteFile(consignmentPath, []byte("test content"), 0644))

	// Verify file exists
	_, err := os.Stat(consignmentPath)
	require.NoError(t, err)

	// Test: Delete consignment
	err = DeleteConsignment(consignmentPath)

	// Verify: File was deleted
	require.NoError(t, err)
	_, err = os.Stat(consignmentPath)
	assert.True(t, os.IsNotExist(err), "File should no longer exist")
}

// TestDeleteConsignment_NonexistentFile tests deleting file that doesn't exist
func TestDeleteConsignment_NonexistentFile(t *testing.T) {
	// Setup: Use path to nonexistent file
	tempDir := t.TempDir()
	consignmentPath := filepath.Join(tempDir, "nonexistent.md")

	// Test: Attempt to delete
	err := DeleteConsignment(consignmentPath)

	// Verify: Error is returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete")
}

// TestDeleteConsignments_Multiple tests deleting multiple consignment files
func TestDeleteConsignments_Multiple(t *testing.T) {
	// Setup: Create multiple temp consignment files
	tempDir := t.TempDir()

	files := []string{
		filepath.Join(tempDir, "consignment1.md"),
		filepath.Join(tempDir, "consignment2.md"),
		filepath.Join(tempDir, "consignment3.md"),
	}

	for _, file := range files {
		require.NoError(t, os.WriteFile(file, []byte("content"), 0644))
	}

	// Verify all files exist
	for _, file := range files {
		_, err := os.Stat(file)
		require.NoError(t, err)
	}

	// Test: Delete all consignments
	err := DeleteConsignments(files)

	// Verify: All files were deleted
	require.NoError(t, err)
	for _, file := range files {
		_, err := os.Stat(file)
		assert.True(t, os.IsNotExist(err), "File %s should no longer exist", file)
	}
}

// TestDeleteConsignments_PartialFailure tests handling when some files fail to delete
func TestDeleteConsignments_PartialFailure(t *testing.T) {
	// Setup: Create some files, leave others nonexistent
	tempDir := t.TempDir()

	existingFile := filepath.Join(tempDir, "existing.md")
	require.NoError(t, os.WriteFile(existingFile, []byte("content"), 0644))

	nonexistentFile := filepath.Join(tempDir, "nonexistent.md")

	files := []string{existingFile, nonexistentFile}

	// Test: Attempt to delete all
	err := DeleteConsignments(files)

	// Verify: Error is returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete")
}

// TestDeleteConsignments_EmptyList tests deleting empty list
func TestDeleteConsignments_EmptyList(t *testing.T) {
	// Test: Delete empty list
	err := DeleteConsignments([]string{})

	// Verify: No error
	require.NoError(t, err)
}

// TestDeleteConsignment_Directory tests error when path is a directory
func TestDeleteConsignment_Directory(t *testing.T) {
	// Setup: Create a directory instead of file
	tempDir := t.TempDir()
	dirPath := filepath.Join(tempDir, "testdir")
	require.NoError(t, os.Mkdir(dirPath, 0755))

	// Test: Attempt to delete directory
	err := DeleteConsignment(dirPath)

	// Verify: Error is returned (directories shouldn't be deleted as consignments)
	assert.Error(t, err)
}

// TestDeleteConsignmentByID_Success tests deleting consignment by ID
func TestDeleteConsignmentByID_Success(t *testing.T) {
	// Setup: Create consignment directory with files
	tempDir := t.TempDir()
	consignmentsDir := filepath.Join(tempDir, "consignments")
	require.NoError(t, os.Mkdir(consignmentsDir, 0755))

	consignmentID := "test-id-123"
	consignmentPath := filepath.Join(consignmentsDir, consignmentID+".md")
	require.NoError(t, os.WriteFile(consignmentPath, []byte("content"), 0644))

	// Test: Delete by ID
	err := DeleteConsignmentByID(consignmentsDir, consignmentID)

	// Verify: File was deleted
	require.NoError(t, err)
	_, err = os.Stat(consignmentPath)
	assert.True(t, os.IsNotExist(err))
}

// TestDeleteConsignmentByID_NotFound tests deleting nonexistent consignment by ID
func TestDeleteConsignmentByID_NotFound(t *testing.T) {
	// Setup: Create empty consignments directory
	tempDir := t.TempDir()
	consignmentsDir := filepath.Join(tempDir, "consignments")
	require.NoError(t, os.Mkdir(consignmentsDir, 0755))

	// Test: Attempt to delete nonexistent ID
	err := DeleteConsignmentByID(consignmentsDir, "nonexistent-id")

	// Verify: Error is returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete")
}

// TestDeleteConsignmentsByIDs_Multiple tests deleting multiple consignments by ID
func TestDeleteConsignmentsByIDs_Multiple(t *testing.T) {
	// Setup: Create consignment directory with multiple files
	tempDir := t.TempDir()
	consignmentsDir := filepath.Join(tempDir, "consignments")
	require.NoError(t, os.Mkdir(consignmentsDir, 0755))

	ids := []string{"id1", "id2", "id3"}
	for _, id := range ids {
		path := filepath.Join(consignmentsDir, id+".md")
		require.NoError(t, os.WriteFile(path, []byte("content"), 0644))
	}

	// Test: Delete all by IDs
	err := DeleteConsignmentsByIDs(consignmentsDir, ids)

	// Verify: All files were deleted
	require.NoError(t, err)
	for _, id := range ids {
		path := filepath.Join(consignmentsDir, id+".md")
		_, err := os.Stat(path)
		assert.True(t, os.IsNotExist(err), "File for ID %s should no longer exist", id)
	}
}

// TestDeleteConsignmentsByIDs_EmptyList tests deleting empty ID list
func TestDeleteConsignmentsByIDs_EmptyList(t *testing.T) {
	// Setup: Create consignments directory
	tempDir := t.TempDir()
	consignmentsDir := filepath.Join(tempDir, "consignments")
	require.NoError(t, os.Mkdir(consignmentsDir, 0755))

	// Test: Delete empty list
	err := DeleteConsignmentsByIDs(consignmentsDir, []string{})

	// Verify: No error
	require.NoError(t, err)
}

// TestDeleteConsignments_PreserveUnrelated tests that unrelated files are not deleted
func TestDeleteConsignments_PreserveUnrelated(t *testing.T) {
	// Setup: Create consignment files and other files
	tempDir := t.TempDir()

	consignmentFile := filepath.Join(tempDir, "consignment.md")
	otherFile := filepath.Join(tempDir, "other.txt")

	require.NoError(t, os.WriteFile(consignmentFile, []byte("consignment"), 0644))
	require.NoError(t, os.WriteFile(otherFile, []byte("other"), 0644))

	// Test: Delete only consignment file
	err := DeleteConsignments([]string{consignmentFile})
	require.NoError(t, err)

	// Verify: Consignment deleted, other file preserved
	_, err = os.Stat(consignmentFile)
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(otherFile)
	assert.NoError(t, err, "Unrelated file should still exist")
}
