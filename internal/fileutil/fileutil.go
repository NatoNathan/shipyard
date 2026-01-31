package fileutil

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// AtomicWrite writes data to a file atomically using a temp file + rename
// This ensures the file is either fully written or not written at all
func AtomicWrite(path string, data []byte, perm os.FileMode) error {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := EnsureDir(dir); err != nil {
		return fmt.Errorf("failed to ensure directory: %w", err)
	}
	
	// Create temp file in same directory (for atomic rename)
	tmpFile := path + ".tmp"
	
	// Write to temp file
	if err := os.WriteFile(tmpFile, data, perm); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	
	// Atomic rename
	if err := os.Rename(tmpFile, path); err != nil {
		// Clean up temp file on error
		os.Remove(tmpFile)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	
	return nil
}

// EnsureDir creates a directory and all parent directories if they don't exist
func EnsureDir(path string) error {
	if PathExists(path) {
		return nil
	}
	
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	
	return nil
}

// PathExists checks if a file or directory exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir checks if the path exists and is a directory
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// ReadYAMLFile reads and unmarshals a YAML file into the provided interface
func ReadYAMLFile(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}
	
	if err := yaml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal YAML from %s: %w", path, err)
	}
	
	return nil
}

// WriteYAMLFile marshals the provided interface to YAML and writes it atomically
func WriteYAMLFile(path string, v interface{}, perm os.FileMode) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal to YAML: %w", err)
	}
	
	if err := AtomicWrite(path, data, perm); err != nil {
		return fmt.Errorf("failed to write YAML file: %w", err)
	}
	
	return nil
}
