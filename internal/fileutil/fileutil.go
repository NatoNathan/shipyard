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
	path = filepath.Clean(path)

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := EnsureDir(dir); err != nil {
		return fmt.Errorf("failed to ensure directory: %w", err)
	}

	// Create temp file in same directory (for atomic rename)
	tmpFile := path + ".tmp"

	// Write to temp file
	if err := WriteFile(tmpFile, data, perm); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpFile, path); err != nil {
		// Clean up temp file on error
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// EnsureDir creates a directory and all parent directories if they don't exist
func EnsureDir(path string) error {
	path = filepath.Clean(path)
	if PathExists(path) {
		return nil
	}

	if err := MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}

	return nil
}

// PathExists checks if a file or directory exists
func PathExists(path string) bool {
	_, err := os.Stat(filepath.Clean(path))
	return err == nil
}

// IsDir checks if the path exists and is a directory
func IsDir(path string) bool {
	info, err := os.Stat(filepath.Clean(path))
	if err != nil {
		return false
	}
	return info.IsDir()
}

// ReadFile reads a project-scoped file path. Callers construct these paths from
// repository configuration, detected package manifests, or temporary files.
func ReadFile(path string) ([]byte, error) {
	cleanPath := filepath.Clean(path)
	return os.ReadFile(cleanPath) // #nosec G304 -- Shipyard intentionally reads configured repository paths.
}

// WriteFile writes a project-scoped file path. Shipyard writes user-visible
// repository artifacts, so 0644 is an intentional default for many callers.
func WriteFile(path string, data []byte, perm os.FileMode) error {
	cleanPath := filepath.Clean(path)
	return os.WriteFile(cleanPath, data, perm) // #nosec G306,G703 -- paths are repository-scoped artifacts selected by config/detection.
}

// MkdirAll creates a project-scoped directory tree. Repository directories are
// intentionally traversable by the owning user and normal development tools.
func MkdirAll(path string, perm os.FileMode) error {
	cleanPath := filepath.Clean(path)
	return os.MkdirAll(cleanPath, perm) // #nosec G301 -- repository artifact directories intentionally use caller-supplied permissions.
}

// OpenFile opens a project-scoped file path with caller-selected flags.
func OpenFile(path string, flag int, perm os.FileMode) (*os.File, error) {
	cleanPath := filepath.Clean(path)
	return os.OpenFile(cleanPath, flag, perm) // #nosec G304 -- path is a repository or executable path selected by Shipyard logic.
}

// ReadYAMLFile reads and unmarshals a YAML file into the provided interface
func ReadYAMLFile(path string, v interface{}) error {
	data, err := ReadFile(path)
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
