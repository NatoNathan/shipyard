package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAppendToHistory_WithEntries tests appending Entry objects (new format)
func TestAppendToHistory_WithEntries(t *testing.T) {
	// Setup: Create empty history file
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.json")
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))

	// Test: Append Entry objects with version context
	entries := []Entry{
		{
			Version:   "1.0.0",
			Package:   "core",
			Tag:       "core-v1.0.0",
			Timestamp: time.Now(),
			Consignments: []Consignment{
				{ID: "c1", Summary: "Fix bug", ChangeType: "patch"},
				{ID: "c2", Summary: "Add test", ChangeType: "patch"},
			},
		},
		{
			Version:   "1.0.0",
			Package:   "api",
			Tag:       "api-v1.0.0",
			Timestamp: time.Now(),
			Consignments: []Consignment{
				{ID: "c3", Summary: "Add endpoint", ChangeType: "minor"},
			},
		},
	}

	err := AppendToHistory(historyPath, entries)

	// Verify: Entries written correctly
	require.NoError(t, err)

	// Read back and verify structure
	readEntries, err := ReadHistory(historyPath)
	require.NoError(t, err)
	require.Len(t, readEntries, 2)

	// Verify first entry
	assert.Equal(t, "1.0.0", readEntries[0].Version)
	assert.Equal(t, "core", readEntries[0].Package)
	assert.Equal(t, "core-v1.0.0", readEntries[0].Tag)
	assert.Len(t, readEntries[0].Consignments, 2)

	// Verify second entry
	assert.Equal(t, "1.0.0", readEntries[1].Version)
	assert.Equal(t, "api", readEntries[1].Package)
	assert.Equal(t, "api-v1.0.0", readEntries[1].Tag)
	assert.Len(t, readEntries[1].Consignments, 1)
}

// TestAppendToHistory_ConcurrentWritesWithEntries tests concurrent writes with Entry format
func TestAppendToHistory_ConcurrentWritesWithEntries(t *testing.T) {
	// Setup: Create empty history file
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.json")
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))

	// Test: Multiple goroutines writing entries concurrently
	var wg sync.WaitGroup
	numWriters := 5

	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			entry := Entry{
				Version:   "1.0.0",
				Package:   string(rune('a' + id)),
				Tag:       string(rune('a'+id)) + "-v1.0.0",
				Timestamp: time.Now(),
				Consignments: []Consignment{
					{ID: "c" + string(rune('1'+id)), Summary: "Test", ChangeType: "patch"},
				},
			}

			err := AppendToHistory(historyPath, []Entry{entry})
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Verify: All writes succeeded
	entries, err := ReadHistory(historyPath)
	require.NoError(t, err)
	assert.Len(t, entries, numWriters, "All concurrent writes should be preserved")
}

// TestAppendToHistory_EmptyConsignmentsInEntry tests appending entries with no consignments
func TestAppendToHistory_EmptyConsignmentsInEntry(t *testing.T) {
	// Setup: Create empty history file
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.json")
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))

	// Test: Append entry with no consignments
	entry := Entry{
		Version:      "1.0.0",
		Package:      "core",
		Tag:          "core-v1.0.0",
		Timestamp:    time.Now(),
		Consignments: []Consignment{},
	}

	err := AppendToHistory(historyPath, []Entry{entry})

	// Verify: Entry written successfully
	require.NoError(t, err)

	entries, err := ReadHistory(historyPath)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Len(t, entries[0].Consignments, 0)
}

// TestAppendToHistory_EmptyEntriesList tests appending empty list
func TestAppendToHistory_EmptyEntriesList(t *testing.T) {
	// Setup: Create history file
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.json")
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))

	// Test: Append empty list
	err := AppendToHistory(historyPath, []Entry{})

	// Verify: No error, file unchanged
	require.NoError(t, err)

	data, err := os.ReadFile(historyPath)
	require.NoError(t, err)

	var history []Entry
	require.NoError(t, json.Unmarshal(data, &history))
	assert.Len(t, history, 0)
}

// TestAppendToHistory_NonexistentFile tests error when history file doesn't exist
func TestAppendToHistory_NonexistentFile(t *testing.T) {
	// Setup: Use path to nonexistent file
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "nonexistent.json")

	// Test: Attempt to append
	entry := Entry{
		Version:   "1.0.0",
		Package:   "core",
		Tag:       "core-v1.0.0",
		Timestamp: time.Now(),
		Consignments: []Consignment{
			{ID: "c1", Summary: "Test", ChangeType: "patch"},
		},
	}

	err := AppendToHistory(historyPath, []Entry{entry})

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
	entry := Entry{
		Version:   "1.0.0",
		Package:   "core",
		Tag:       "core-v1.0.0",
		Timestamp: time.Now(),
		Consignments: []Consignment{
			{ID: "c1", Summary: "Test", ChangeType: "patch"},
		},
	}

	err := AppendToHistory(historyPath, []Entry{entry})

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

	// Test: Append entry with rich metadata
	entry := Entry{
		Version:   "1.0.0",
		Package:   "core",
		Tag:       "core-v1.0.0",
		Timestamp: time.Now(),
		Consignments: []Consignment{
			{
				ID:         "meta-1",
				Summary:    "Feature with metadata",
				ChangeType: "minor",
				Metadata: map[string]interface{}{
					"author":   "alice@example.com",
					"issue":    "JIRA-123",
					"breaking": false,
				},
			},
		},
	}

	err := AppendToHistory(historyPath, []Entry{entry})
	require.NoError(t, err)

	// Verify: Metadata is preserved
	entries, err := ReadHistory(historyPath)
	require.NoError(t, err)

	assert.Len(t, entries, 1)
	assert.Len(t, entries[0].Consignments, 1)
	assert.Equal(t, "alice@example.com", entries[0].Consignments[0].Metadata["author"])
	assert.Equal(t, "JIRA-123", entries[0].Consignments[0].Metadata["issue"])
	assert.Equal(t, false, entries[0].Consignments[0].Metadata["breaking"])
}

// TestAppendToHistory_AtomicWrite tests that writes are atomic (no partial writes)
func TestAppendToHistory_AtomicWrite(t *testing.T) {
	// Setup: Create history file
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.json")
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))

	// Test: Append entry and verify file is always valid JSON
	entry := Entry{
		Version:   "1.0.0",
		Package:   "core",
		Tag:       "core-v1.0.0",
		Timestamp: time.Now(),
		Consignments: []Consignment{
			{ID: "atomic-1", Summary: "Atomic test", ChangeType: "patch"},
		},
	}

	err := AppendToHistory(historyPath, []Entry{entry})
	require.NoError(t, err)

	// Verify: File contains valid JSON
	data, err := os.ReadFile(historyPath)
	require.NoError(t, err)

	var history []Entry
	require.NoError(t, json.Unmarshal(data, &history), "History file should always contain valid JSON")
	assert.Len(t, history, 1)
}

// TestAppendToHistory_PreservesOrder tests that entries are appended in order
func TestAppendToHistory_PreservesOrder(t *testing.T) {
	// Setup: Create history with existing entries
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.json")

	existing := []Entry{
		{Version: "1.0.0", Package: "core", Tag: "core-v1.0.0", Timestamp: time.Now(), Consignments: []Consignment{{ID: "1"}}},
		{Version: "1.1.0", Package: "core", Tag: "core-v1.1.0", Timestamp: time.Now(), Consignments: []Consignment{{ID: "2"}}},
	}
	data, err := json.Marshal(existing)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(historyPath, data, 0644))

	// Test: Append multiple in specific order
	newEntries := []Entry{
		{Version: "1.2.0", Package: "core", Tag: "core-v1.2.0", Timestamp: time.Now(), Consignments: []Consignment{{ID: "3"}}},
		{Version: "1.3.0", Package: "core", Tag: "core-v1.3.0", Timestamp: time.Now(), Consignments: []Consignment{{ID: "4"}}},
		{Version: "1.4.0", Package: "core", Tag: "core-v1.4.0", Timestamp: time.Now(), Consignments: []Consignment{{ID: "5"}}},
	}

	err = AppendToHistory(historyPath, newEntries)
	require.NoError(t, err)

	// Verify: Order is preserved
	entries, err := ReadHistory(historyPath)
	require.NoError(t, err)

	require.Len(t, entries, 5)
	assert.Equal(t, "1", entries[0].Consignments[0].ID)
	assert.Equal(t, "2", entries[1].Consignments[0].ID)
	assert.Equal(t, "3", entries[2].Consignments[0].ID)
	assert.Equal(t, "4", entries[3].Consignments[0].ID)
	assert.Equal(t, "5", entries[4].Consignments[0].ID)
}
