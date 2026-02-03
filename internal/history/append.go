package history

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gofrs/flock"
)

// AppendToHistory appends history entries to the history file with file locking
// Returns error if file doesn't exist, contains invalid JSON, or write fails
func AppendToHistory(historyPath string, entries []Entry) error {
	// Early return for empty list
	if len(entries) == 0 {
		return nil
	}

	// Create file lock
	fileLock := flock.New(historyPath + ".lock")

	// Acquire exclusive lock
	if err := fileLock.Lock(); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer fileLock.Unlock()

	// Read existing history
	data, err := os.ReadFile(historyPath)
	if err != nil {
		return fmt.Errorf("failed to read history file: %w", err)
	}

	// Parse existing history
	var history []Entry
	if err := json.Unmarshal(data, &history); err != nil {
		return fmt.Errorf("failed to unmarshal history: %w", err)
	}

	// Append new entries
	history = append(history, entries...)

	// Marshal updated history
	updatedData, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	// Write atomically: write to temp file, then rename
	tempPath := historyPath + ".tmp"
	if err := os.WriteFile(tempPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, historyPath); err != nil {
		// Clean up temp file on failure
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
