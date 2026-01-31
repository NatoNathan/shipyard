package history

import (
	"encoding/json"
	"os"
	"time"
)

// Entry represents a version history entry
type Entry struct {
	Version      string        `json:"version"`
	Package      string        `json:"package"`
	Timestamp    time.Time     `json:"timestamp"`
	Consignments []Consignment `json:"consignments"`
}

// Consignment represents a change in a version
type Consignment struct {
	ID         string `json:"id"`
	Summary    string `json:"summary"`
	ChangeType string `json:"changeType"`
}

// ReadHistory reads history entries from a JSON file
func ReadHistory(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}
