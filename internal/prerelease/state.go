package prerelease

import (
	"fmt"
	"os"

	"github.com/gofrs/flock"
	"gopkg.in/yaml.v3"
)

// State represents the pre-release state file (.shipyard/prerelease.yml)
type State struct {
	Packages map[string]PackageState `yaml:"packages"`
}

// PackageState tracks the pre-release state for a single package
type PackageState struct {
	Stage         string `yaml:"stage"`
	Counter       int    `yaml:"counter"`
	TargetVersion string `yaml:"targetVersion"`
}

// ReadState reads the pre-release state from the given path with shared file locking.
// Returns an empty state (not an error) if the file does not exist.
func ReadState(path string) (*State, error) {
	// Check existence first to avoid lock file creation for missing files
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &State{Packages: make(map[string]PackageState)}, nil
	}

	// Acquire shared (read) lock
	fileLock := flock.New(path + ".lock")
	if err := fileLock.RLock(); err != nil {
		return nil, fmt.Errorf("failed to acquire read lock: %w", err)
	}
	defer func() { _ = fileLock.Unlock() }()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{Packages: make(map[string]PackageState)}, nil
		}
		return nil, fmt.Errorf("failed to read prerelease state: %w", err)
	}

	var state State
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse prerelease state: %w", err)
	}

	if state.Packages == nil {
		state.Packages = make(map[string]PackageState)
	}

	return &state, nil
}

// WriteState writes the pre-release state to the given path with file locking.
// Uses atomic write (write to temp file, then rename).
func WriteState(path string, state *State) error {
	fileLock := flock.New(path + ".lock")
	if err := fileLock.Lock(); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer func() { _ = fileLock.Unlock() }()

	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal prerelease state: %w", err)
	}

	// Write to temp file first
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// DeleteState removes the pre-release state file.
// Returns nil if the file does not exist.
func DeleteState(path string) error {
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete prerelease state: %w", err)
	}
	return nil
}

// Exists checks if the pre-release state file exists at the given path.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
