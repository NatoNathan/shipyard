package history

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadHistory tests reading history from JSON file
func TestReadHistory(t *testing.T) {
	t.Run("reads valid history file", func(t *testing.T) {
		// Setup: Create history file
		tempDir := t.TempDir()
		historyPath := filepath.Join(tempDir, "history.json")

		content := `[
  {
    "version": "1.1.0",
    "package": "core",
    "timestamp": "2026-01-30T10:00:00Z",
    "consignments": [
      {
        "id": "c1",
        "summary": "Add feature",
        "changeType": "minor"
      }
    ]
  }
]`
		require.NoError(t, os.WriteFile(historyPath, []byte(content), 0644))

		// Test: Read history
		entries, err := ReadHistory(historyPath)

		// Verify: History loaded correctly
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, "1.1.0", entries[0].Version)
		assert.Equal(t, "core", entries[0].Package)
		assert.Len(t, entries[0].Consignments, 1)
	})

	t.Run("handles empty history", func(t *testing.T) {
		// Setup: Create empty history file
		tempDir := t.TempDir()
		historyPath := filepath.Join(tempDir, "history.json")
		require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))

		// Test: Read history
		entries, err := ReadHistory(historyPath)

		// Verify: Empty slice returned
		require.NoError(t, err)
		assert.Len(t, entries, 0)
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		// Test: Read non-existent file
		entries, err := ReadHistory("/nonexistent/history.json")

		// Verify: Error returned
		assert.Error(t, err)
		assert.Nil(t, entries)
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		// Setup: Create invalid JSON file
		tempDir := t.TempDir()
		historyPath := filepath.Join(tempDir, "history.json")
		require.NoError(t, os.WriteFile(historyPath, []byte("invalid json"), 0644))

		// Test: Read history
		entries, err := ReadHistory(historyPath)

		// Verify: Error returned
		assert.Error(t, err)
		assert.Nil(t, entries)
	})
}
