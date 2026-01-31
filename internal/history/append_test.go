package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/NatoNathan/shipyard/internal/consignment"
	"github.com/NatoNathan/shipyard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAppendToHistory_Success tests appending a single consignment to history
func TestAppendToHistory_Success(t *testing.T) {
	// Setup: Create temp history file with existing entries
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.json")

	existing := []consignment.Consignment{
		{
			ID:         "existing-1",
			Timestamp:  time.Now().Add(-24 * time.Hour),
			Packages:   []string{"pkg1"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "Existing change",
		},
	}
	data, err := json.Marshal(existing)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(historyPath, data, 0644))

	// Test: Append new consignment
	newConsignment := &consignment.Consignment{
		ID:         "new-1",
		Timestamp:  time.Now(),
		Packages:   []string{"pkg2"},
		ChangeType: types.ChangeTypeMinor,
		Summary:    "New feature",
		Metadata: map[string]interface{}{
			"author": "test@example.com",
		},
	}

	err = AppendToHistory(historyPath, []*consignment.Consignment{newConsignment})

	// Verify: History file contains both entries
	require.NoError(t, err)

	data, err = os.ReadFile(historyPath)
	require.NoError(t, err)

	var history []consignment.Consignment
	require.NoError(t, json.Unmarshal(data, &history))

	assert.Len(t, history, 2)
	assert.Equal(t, "existing-1", history[0].ID)
	assert.Equal(t, "new-1", history[1].ID)
}

// TestAppendToHistory_Multiple tests appending multiple consignments at once
func TestAppendToHistory_Multiple(t *testing.T) {
	// Setup: Create empty history file
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.json")
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))

	// Test: Append multiple consignments
	consignments := []*consignment.Consignment{
		{
			ID:         "c1",
			Timestamp:  time.Now(),
			Packages:   []string{"pkg1"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "Fix 1",
		},
		{
			ID:         "c2",
			Timestamp:  time.Now(),
			Packages:   []string{"pkg2"},
			ChangeType: types.ChangeTypeMinor,
			Summary:    "Feature 1",
		},
		{
			ID:         "c3",
			Timestamp:  time.Now(),
			Packages:   []string{"pkg1", "pkg2"},
			ChangeType: types.ChangeTypeMajor,
			Summary:    "Breaking change",
		},
	}

	err := AppendToHistory(historyPath, consignments)

	// Verify: All consignments are in history
	require.NoError(t, err)

	data, err := os.ReadFile(historyPath)
	require.NoError(t, err)

	var history []consignment.Consignment
	require.NoError(t, json.Unmarshal(data, &history))

	assert.Len(t, history, 3)
	assert.Equal(t, "c1", history[0].ID)
	assert.Equal(t, "c2", history[1].ID)
	assert.Equal(t, "c3", history[2].ID)
}

// TestAppendToHistory_EmptyFile tests appending to newly initialized file
func TestAppendToHistory_EmptyFile(t *testing.T) {
	// Setup: Create empty JSON array file
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.json")
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))

	// Test: Append first consignment
	c := &consignment.Consignment{
		ID:         "first-1",
		Timestamp:  time.Now(),
		Packages:   []string{"pkg1"},
		ChangeType: types.ChangeTypePatch,
		Summary:    "First entry",
	}

	err := AppendToHistory(historyPath, []*consignment.Consignment{c})

	// Verify: History contains the entry
	require.NoError(t, err)

	data, err := os.ReadFile(historyPath)
	require.NoError(t, err)

	var history []consignment.Consignment
	require.NoError(t, json.Unmarshal(data, &history))

	assert.Len(t, history, 1)
	assert.Equal(t, "first-1", history[0].ID)
}

// TestAppendToHistory_NonexistentFile tests error when history file doesn't exist
func TestAppendToHistory_NonexistentFile(t *testing.T) {
	// Setup: Use path to nonexistent file
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "nonexistent.json")

	// Test: Attempt to append
	c := &consignment.Consignment{
		ID:         "c1",
		Timestamp:  time.Now(),
		Packages:   []string{"pkg1"},
		ChangeType: types.ChangeTypePatch,
		Summary:    "Test",
	}

	err := AppendToHistory(historyPath, []*consignment.Consignment{c})

	// Verify: Error is returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read history")
}

// TestAppendToHistory_InvalidJSON tests error when history file contains invalid JSON
func TestAppendToHistory_InvalidJSON(t *testing.T) {
	// Setup: Create file with invalid JSON
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.json")
	require.NoError(t, os.WriteFile(historyPath, []byte("not valid json"), 0644))

	// Test: Attempt to append
	c := &consignment.Consignment{
		ID:         "c1",
		Timestamp:  time.Now(),
		Packages:   []string{"pkg1"},
		ChangeType: types.ChangeTypePatch,
		Summary:    "Test",
	}

	err := AppendToHistory(historyPath, []*consignment.Consignment{c})

	// Verify: Error is returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal history")
}

// TestAppendToHistory_PreservesMetadata tests that metadata is preserved during archival
func TestAppendToHistory_PreservesMetadata(t *testing.T) {
	// Setup: Create empty history file
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.json")
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))

	// Test: Append consignment with rich metadata
	c := &consignment.Consignment{
		ID:         "meta-1",
		Timestamp:  time.Now(),
		Packages:   []string{"pkg1"},
		ChangeType: types.ChangeTypeMinor,
		Summary:    "Feature with metadata",
		Metadata: map[string]interface{}{
			"author":      "alice@example.com",
			"issue":       "JIRA-123",
			"breaking":    false,
			"tags":        []string{"feature", "api"},
			"reviewers":   []string{"bob", "charlie"},
			"custom_data": map[string]interface{}{"key": "value"},
		},
	}

	err := AppendToHistory(historyPath, []*consignment.Consignment{c})
	require.NoError(t, err)

	// Verify: Metadata is preserved
	data, err := os.ReadFile(historyPath)
	require.NoError(t, err)

	var history []consignment.Consignment
	require.NoError(t, json.Unmarshal(data, &history))

	assert.Len(t, history, 1)
	assert.Equal(t, "alice@example.com", history[0].Metadata["author"])
	assert.Equal(t, "JIRA-123", history[0].Metadata["issue"])
	assert.Equal(t, false, history[0].Metadata["breaking"])
}

// TestAppendToHistory_AtomicWrite tests that writes are atomic (no partial writes)
func TestAppendToHistory_AtomicWrite(t *testing.T) {
	// Setup: Create history file
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.json")
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))

	// Test: Append consignment and verify file is always valid JSON
	c := &consignment.Consignment{
		ID:         "atomic-1",
		Timestamp:  time.Now(),
		Packages:   []string{"pkg1"},
		ChangeType: types.ChangeTypePatch,
		Summary:    "Atomic test",
	}

	err := AppendToHistory(historyPath, []*consignment.Consignment{c})
	require.NoError(t, err)

	// Verify: File contains valid JSON
	data, err := os.ReadFile(historyPath)
	require.NoError(t, err)

	var history []consignment.Consignment
	require.NoError(t, json.Unmarshal(data, &history), "History file should always contain valid JSON")
	assert.Len(t, history, 1)
}

// TestAppendToHistory_ConcurrentWrites tests file locking with concurrent writes
func TestAppendToHistory_ConcurrentWrites(t *testing.T) {
	// Setup: Create empty history file
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.json")
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))

	// Test: Multiple goroutines writing concurrently
	var wg sync.WaitGroup
	numWriters := 10

	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			c := &consignment.Consignment{
				ID:         string(rune('a' + id)),
				Timestamp:  time.Now(),
				Packages:   []string{"pkg1"},
				ChangeType: types.ChangeTypePatch,
				Summary:    "Concurrent write",
			}

			err := AppendToHistory(historyPath, []*consignment.Consignment{c})
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Verify: All writes succeeded and history is valid
	data, err := os.ReadFile(historyPath)
	require.NoError(t, err)

	var history []consignment.Consignment
	require.NoError(t, json.Unmarshal(data, &history), "History must remain valid JSON despite concurrent writes")
	assert.Len(t, history, numWriters, "All concurrent writes should be preserved")
}

// TestAppendToHistory_EmptyConsignmentList tests appending empty list
func TestAppendToHistory_EmptyConsignmentList(t *testing.T) {
	// Setup: Create history file
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.json")
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))

	// Test: Append empty list
	err := AppendToHistory(historyPath, []*consignment.Consignment{})

	// Verify: No error, file unchanged
	require.NoError(t, err)

	data, err := os.ReadFile(historyPath)
	require.NoError(t, err)

	var history []consignment.Consignment
	require.NoError(t, json.Unmarshal(data, &history))
	assert.Len(t, history, 0)
}

// TestAppendToHistory_PreservesOrder tests that consignments are appended in order
func TestAppendToHistory_PreservesOrder(t *testing.T) {
	// Setup: Create history with existing entries
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.json")

	existing := []consignment.Consignment{
		{ID: "1", Timestamp: time.Now(), Packages: []string{"pkg1"}, ChangeType: types.ChangeTypePatch, Summary: "First"},
		{ID: "2", Timestamp: time.Now(), Packages: []string{"pkg1"}, ChangeType: types.ChangeTypePatch, Summary: "Second"},
	}
	data, err := json.Marshal(existing)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(historyPath, data, 0644))

	// Test: Append multiple in specific order
	newConsignments := []*consignment.Consignment{
		{ID: "3", Timestamp: time.Now(), Packages: []string{"pkg1"}, ChangeType: types.ChangeTypePatch, Summary: "Third"},
		{ID: "4", Timestamp: time.Now(), Packages: []string{"pkg1"}, ChangeType: types.ChangeTypePatch, Summary: "Fourth"},
		{ID: "5", Timestamp: time.Now(), Packages: []string{"pkg1"}, ChangeType: types.ChangeTypePatch, Summary: "Fifth"},
	}

	err = AppendToHistory(historyPath, newConsignments)
	require.NoError(t, err)

	// Verify: Order is preserved
	data, err = os.ReadFile(historyPath)
	require.NoError(t, err)

	var history []consignment.Consignment
	require.NoError(t, json.Unmarshal(data, &history))

	require.Len(t, history, 5)
	assert.Equal(t, "1", history[0].ID)
	assert.Equal(t, "2", history[1].ID)
	assert.Equal(t, "3", history[2].ID)
	assert.Equal(t, "4", history[3].ID)
	assert.Equal(t, "5", history[4].ID)
}
