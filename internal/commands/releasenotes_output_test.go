package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReleaseNotesCommand_OutputFile tests writing release notes to a file
func TestReleaseNotesCommand_OutputFile(t *testing.T) {
	t.Run("writes to specified output file", func(t *testing.T) {
		// Setup: Create repo with history
		tempDir := setupReleaseNotesTestRepo(t)
		defer changeToDir(t, tempDir)()

		outputFile := filepath.Join(tempDir, "RELEASE_NOTES.md")

		// Test: Run with --output flag
		cmd := NewReleaseNotesCommand()
		cmd.SetArgs([]string{"--output", outputFile})

		err := cmd.Execute()
		require.NoError(t, err)

		// Verify: File created with content
		assert.FileExists(t, outputFile, "Output file should be created")
		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		assert.Contains(t, string(content), "Release Notes", "File should contain release notes header")
		assert.Contains(t, string(content), "1.1.0", "File should contain version")
	})

	t.Run("overwrites existing output file", func(t *testing.T) {
		// Setup: Create repo with history
		tempDir := setupReleaseNotesTestRepo(t)
		defer changeToDir(t, tempDir)()

		outputFile := filepath.Join(tempDir, "RELEASE_NOTES.md")

		// Create existing file
		require.NoError(t, os.WriteFile(outputFile, []byte("old content"), 0644))

		// Test: Run with --output flag
		cmd := NewReleaseNotesCommand()
		cmd.SetArgs([]string{"--output", outputFile})

		err := cmd.Execute()
		require.NoError(t, err)

		// Verify: File overwritten
		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		assert.NotContains(t, string(content), "old content", "Old content should be replaced")
		assert.Contains(t, string(content), "Release Notes", "File should contain new release notes")
	})

	t.Run("creates output file in subdirectory", func(t *testing.T) {
		// Setup: Create repo with history
		tempDir := setupReleaseNotesTestRepo(t)
		defer changeToDir(t, tempDir)()

		// Create subdirectory
		docsDir := filepath.Join(tempDir, "docs")
		require.NoError(t, os.MkdirAll(docsDir, 0755))

		outputFile := filepath.Join(docsDir, "CHANGELOG.md")

		// Test: Run with --output flag to subdirectory
		cmd := NewReleaseNotesCommand()
		cmd.SetArgs([]string{"--output", outputFile})

		err := cmd.Execute()
		require.NoError(t, err)

		// Verify: File created in subdirectory
		assert.FileExists(t, outputFile, "Output file should be created in subdirectory")
		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		assert.Contains(t, string(content), "Release Notes")
	})

	t.Run("handles error for invalid output path", func(t *testing.T) {
		// Setup: Create repo with history
		tempDir := setupReleaseNotesTestRepo(t)
		defer changeToDir(t, tempDir)()

		// Test: Run with invalid output path
		outputFile := filepath.Join("/nonexistent", "directory", "file.md")
		cmd := NewReleaseNotesCommand()
		cmd.SetArgs([]string{"--output", outputFile})

		err := cmd.Execute()

		// Verify: Error returned
		assert.Error(t, err, "Should error for invalid output path")
	})

	t.Run("writes empty release notes for no history", func(t *testing.T) {
		// Setup: Create repo without history
		tempDir := t.TempDir()
		setupMinimalShipyardDir(t, tempDir)
		defer changeToDir(t, tempDir)()

		outputFile := filepath.Join(tempDir, "RELEASE_NOTES.md")

		// Test: Run with --output flag
		cmd := NewReleaseNotesCommand()
		cmd.SetArgs([]string{"--output", outputFile})

		err := cmd.Execute()
		require.NoError(t, err)

		// Verify: File created with "no releases" message
		assert.FileExists(t, outputFile)
		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		assert.Contains(t, string(content), "No releases found")
	})
}

// TestReleaseNotesCommand_OutputFormats tests different output scenarios
func TestReleaseNotesCommand_OutputFormats(t *testing.T) {
	t.Run("filters by package and writes to file", func(t *testing.T) {
		// Setup: Create repo with multi-package history
		tempDir := setupReleaseNotesTestRepo(t)
		defer changeToDir(t, tempDir)()

		outputFile := filepath.Join(tempDir, "CORE_RELEASES.md")

		// Test: Run with both --package and --output flags
		cmd := NewReleaseNotesCommand()
		cmd.SetArgs([]string{"--package", "core", "--output", outputFile})

		err := cmd.Execute()
		require.NoError(t, err)

		// Verify: File contains only core package
		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		assert.Contains(t, string(content), "core")
		assert.Contains(t, string(content), "1.1.0")
	})

	t.Run("filters by version and writes to file", func(t *testing.T) {
		// Setup: Create repo with history
		tempDir := setupReleaseNotesTestRepo(t)
		defer changeToDir(t, tempDir)()

		outputFile := filepath.Join(tempDir, "v1.1.0.md")

		// Test: Run with both --version and --output flags
		cmd := NewReleaseNotesCommand()
		cmd.SetArgs([]string{"--version", "1.1.0", "--output", outputFile})

		err := cmd.Execute()
		require.NoError(t, err)

		// Verify: File contains only specified version
		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		assert.Contains(t, string(content), "1.1.0")
		assert.NotContains(t, string(content), "1.0.1")
	})
}
