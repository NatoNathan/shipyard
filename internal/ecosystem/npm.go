package ecosystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/NatoNathan/shipyard/pkg/semver"
)

// NPMEcosystem handles version management for NPM/Node.js projects
type NPMEcosystem struct {
	path string
}

// NewNPMEcosystem creates a new NPM ecosystem handler
func NewNPMEcosystem(path string) *NPMEcosystem {
	return &NPMEcosystem{path: path}
}

// ReadVersion reads the current version from package.json
func (n *NPMEcosystem) ReadVersion() (semver.Version, error) {
	packageJSONPath := filepath.Join(n.path, "package.json")

	content, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return semver.Version{}, fmt.Errorf("failed to read package.json: %w", err)
	}

	var packageJSON map[string]interface{}
	if err := json.Unmarshal(content, &packageJSON); err != nil {
		return semver.Version{}, fmt.Errorf("failed to parse package.json: %w", err)
	}

	versionStr, ok := packageJSON["version"].(string)
	if !ok {
		return semver.Version{}, fmt.Errorf("no version field found in package.json")
	}

	return semver.Parse(versionStr)
}

// UpdateVersion updates the version in package.json
func (n *NPMEcosystem) UpdateVersion(version semver.Version) error {
	packageJSONPath := filepath.Join(n.path, "package.json")

	// Read existing content
	content, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return fmt.Errorf("failed to read package.json: %w", err)
	}

	// Parse JSON
	var packageJSON map[string]interface{}
	if err := json.Unmarshal(content, &packageJSON); err != nil {
		return fmt.Errorf("failed to parse package.json: %w", err)
	}

	// Update version field
	packageJSON["version"] = version.String()

	// Marshal back to JSON with indentation
	newContent, err := json.MarshalIndent(packageJSON, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal package.json: %w", err)
	}

	// Add newline at end of file (standard for package.json)
	newContent = append(newContent, '\n')

	// Write updated content
	if err := os.WriteFile(packageJSONPath, newContent, 0644); err != nil {
		return fmt.Errorf("failed to write package.json: %w", err)
	}

	return nil
}

// GetVersionFiles returns paths to all version-containing files
func (n *NPMEcosystem) GetVersionFiles() []string {
	packageJSONPath := filepath.Join(n.path, "package.json")

	if _, err := os.Stat(packageJSONPath); err == nil {
		return []string{packageJSONPath}
	}

	return []string{}
}

// DetectNPMEcosystem checks if a directory contains an NPM project
func DetectNPMEcosystem(path string) bool {
	// Check for package.json
	if _, err := os.Stat(filepath.Join(path, "package.json")); err == nil {
		return true
	}

	// Check for package-lock.json
	if _, err := os.Stat(filepath.Join(path, "package-lock.json")); err == nil {
		return true
	}

	// Check for node_modules directory
	if _, err := os.Stat(filepath.Join(path, "node_modules")); err == nil {
		return true
	}

	return false
}
