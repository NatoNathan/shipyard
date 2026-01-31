package consignment

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NatoNathan/shipyard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWriteConsignment_Success tests writing a consignment file
func TestWriteConsignment_Success(t *testing.T) {
	tempDir := t.TempDir()

	cons := &Consignment{
		ID:         "20260130-143022-a1b2c3",
		Timestamp:  time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC),
		Packages:   []string{"core"},
		ChangeType: types.ChangeTypePatch,
		Summary:    "Fixed a bug",
		Metadata:   nil,
	}

	err := WriteConsignment(cons, tempDir)
	require.NoError(t, err, "Should write consignment without error")

	// Verify file was created
	expectedPath := filepath.Join(tempDir, "20260130-143022-a1b2c3.md")
	assert.FileExists(t, expectedPath, "Consignment file should exist")

	// Verify content
	content, err := os.ReadFile(expectedPath)
	require.NoError(t, err, "Should be able to read consignment file")
	assert.Contains(t, string(content), "id: 20260130-143022-a1b2c3", "File should contain ID")
	assert.Contains(t, string(content), "Fixed a bug", "File should contain summary")
}

// TestWriteConsignment_DirectoryCreation tests creating directory if it doesn't exist
func TestWriteConsignment_DirectoryCreation(t *testing.T) {
	tempDir := t.TempDir()
	consignmentsDir := filepath.Join(tempDir, "consignments")

	// Directory doesn't exist yet
	assert.NoDirExists(t, consignmentsDir, "Directory should not exist yet")

	cons := &Consignment{
		ID:         "20260130-143022-a1b2c3",
		Timestamp:  time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC),
		Packages:   []string{"core"},
		ChangeType: types.ChangeTypePatch,
		Summary:    "Test",
		Metadata:   nil,
	}

	err := WriteConsignment(cons, consignmentsDir)
	require.NoError(t, err, "Should create directory and write file")

	// Verify directory was created
	assert.DirExists(t, consignmentsDir, "Directory should be created")

	// Verify file exists
	expectedPath := filepath.Join(consignmentsDir, "20260130-143022-a1b2c3.md")
	assert.FileExists(t, expectedPath, "Consignment file should exist")
}

// TestWriteConsignment_AtomicWrite tests that write is atomic
func TestWriteConsignment_AtomicWrite(t *testing.T) {
	tempDir := t.TempDir()

	cons := &Consignment{
		ID:         "20260130-143022-a1b2c3",
		Timestamp:  time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC),
		Packages:   []string{"core"},
		ChangeType: types.ChangeTypePatch,
		Summary:    "Test atomic write",
		Metadata:   nil,
	}

	// Write first time
	err := WriteConsignment(cons, tempDir)
	require.NoError(t, err, "First write should succeed")

	// Modify and write again (overwrite)
	cons.Summary = "Updated summary"
	err = WriteConsignment(cons, tempDir)
	require.NoError(t, err, "Second write should succeed")

	// Verify content is updated
	expectedPath := filepath.Join(tempDir, "20260130-143022-a1b2c3.md")
	content, err := os.ReadFile(expectedPath)
	require.NoError(t, err, "Should be able to read file")
	assert.Contains(t, string(content), "Updated summary", "File should contain updated summary")
	assert.NotContains(t, string(content), "Test atomic write", "File should not contain old summary")
}

// TestWriteConsignment_WithMetadata tests writing with metadata
func TestWriteConsignment_WithMetadata(t *testing.T) {
	tempDir := t.TempDir()

	cons := &Consignment{
		ID:         "20260130-143022-a1b2c3",
		Timestamp:  time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC),
		Packages:   []string{"core"},
		ChangeType: types.ChangeTypePatch,
		Summary:    "Test",
		Metadata: map[string]interface{}{
			"author": "test@example.com",
			"issue":  "JIRA-123",
		},
	}

	err := WriteConsignment(cons, tempDir)
	require.NoError(t, err, "Should write consignment with metadata")

	// Verify metadata in file
	expectedPath := filepath.Join(tempDir, "20260130-143022-a1b2c3.md")
	content, err := os.ReadFile(expectedPath)
	require.NoError(t, err, "Should be able to read file")
	assert.Contains(t, string(content), "author: test@example.com", "File should contain author metadata")
	assert.Contains(t, string(content), "issue: JIRA-123", "File should contain issue metadata")
}

// TestWriteConsignment_Filename tests correct filename format
func TestWriteConsignment_Filename(t *testing.T) {
	tempDir := t.TempDir()

	cons := &Consignment{
		ID:         "20260130-143022-a1b2c3",
		Timestamp:  time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC),
		Packages:   []string{"core"},
		ChangeType: types.ChangeTypePatch,
		Summary:    "Test",
		Metadata:   nil,
	}

	err := WriteConsignment(cons, tempDir)
	require.NoError(t, err, "Should write consignment")

	// Verify filename matches ID + .md extension
	expectedPath := filepath.Join(tempDir, "20260130-143022-a1b2c3.md")
	assert.FileExists(t, expectedPath, "File should have correct name")

	// Verify no other files were created
	entries, err := os.ReadDir(tempDir)
	require.NoError(t, err, "Should be able to read directory")
	assert.Equal(t, 1, len(entries), "Should have exactly one file")
}
