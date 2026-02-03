package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReleaseNotesCommand_Basic tests basic release notes generation
func TestReleaseNotesCommand_Basic(t *testing.T) {
	t.Run("generates release notes from history", func(t *testing.T) {
		// Setup: Create repo with history
		tempDir := setupReleaseNotesTestRepo(t)
		defer changeToDir(t, tempDir)()

		// Test: Run release-notes command (multi-package repo requires --package)
		cmd := NewReleaseNotesCommand()
		cmd.SetArgs([]string{"--package", "core"})

		output := captureOutput(func() {
			err := cmd.Execute()
			require.NoError(t, err)
		})

		// Verify: Output contains release notes
		assert.Contains(t, output, "Release Notes", "Should contain release notes header")
		assert.Contains(t, output, "1.1.0", "Should contain version from history")
	})

	t.Run("filters by package", func(t *testing.T) {
		// Setup: Create repo with multi-package history
		tempDir := setupReleaseNotesTestRepo(t)
		defer changeToDir(t, tempDir)()

		// Test: Run with --package filter
		cmd := NewReleaseNotesCommand()
		cmd.SetArgs([]string{"--package", "core"})

		output := captureOutput(func() {
			err := cmd.Execute()
			require.NoError(t, err)
		})

		// Verify: Only core package shown
		assert.Contains(t, output, "core", "Should show core package")
	})
}

// TestReleaseNotesCommand_NoHistory tests when no history exists
func TestReleaseNotesCommand_NoHistory(t *testing.T) {
	// Setup: Create repo without history
	tempDir := t.TempDir()
	setupMinimalShipyardDir(t, tempDir)
	defer changeToDir(t, tempDir)()

	// Test: Run release-notes command
	cmd := NewReleaseNotesCommand()
	cmd.SetArgs([]string{})

	output := captureOutput(func() {
		err := cmd.Execute()
		require.NoError(t, err)
	})

	// Verify: Message about no releases
	assert.Contains(t, output, "No releases found", "Should indicate no releases")
}

// Helper functions

func setupReleaseNotesTestRepo(t *testing.T) string {
	t.Helper()
	tempDir := t.TempDir()

	// Create .shipyard structure
	shipyardDir := filepath.Join(tempDir, ".shipyard")
	require.NoError(t, os.MkdirAll(shipyardDir, 0755))

	// Create config
	configContent := `packages:
  - name: core
    path: ./core
    ecosystem: go
  - name: api
    path: ./api
    ecosystem: go
`
	require.NoError(t, os.WriteFile(filepath.Join(shipyardDir, "shipyard.yaml"), []byte(configContent), 0644))

	// Create history with sample releases
	historyContent := `[
  {
    "version": "1.1.0",
    "package": "core",
    "timestamp": "2026-01-30T00:00:00Z",
    "consignments": [
      {
        "id": "c1",
        "summary": "Add new feature",
        "changeType": "minor"
      }
    ]
  },
  {
    "version": "1.0.1",
    "package": "core",
    "timestamp": "2026-01-29T00:00:00Z",
    "consignments": [
      {
        "id": "c2",
        "summary": "Fix bug",
        "changeType": "patch"
      }
    ]
  }
]`
	require.NoError(t, os.WriteFile(filepath.Join(shipyardDir, "history.json"), []byte(historyContent), 0644))

	return tempDir
}

func setupMinimalShipyardDir(t *testing.T, dir string) {
	t.Helper()
	shipyardDir := filepath.Join(dir, ".shipyard")
	require.NoError(t, os.MkdirAll(shipyardDir, 0755))

	// Create empty history
	require.NoError(t, os.WriteFile(filepath.Join(shipyardDir, "history.json"), []byte("[]"), 0644))

	// Create minimal config
	configContent := `packages:
  - name: test
    path: ./
    ecosystem: go
`
	require.NoError(t, os.WriteFile(filepath.Join(shipyardDir, "shipyard.yaml"), []byte(configContent), 0644))
}
