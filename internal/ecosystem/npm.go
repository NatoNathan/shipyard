package ecosystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/NatoNathan/shipyard/pkg/semver"
)

var _ Handler = (*NPMEcosystem)(nil)

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

// UpdateVersion updates the version in package.json using regex replacement
// to preserve original file formatting, indentation, and key order.
func (n *NPMEcosystem) UpdateVersion(version semver.Version) error {
	packageJSONPath := filepath.Join(n.path, "package.json")

	// Read existing content
	content, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return fmt.Errorf("failed to read package.json: %w", err)
	}

	// Use regex to replace only the version value, preserving all formatting
	re := regexp.MustCompile(`("version"\s*:\s*")([^"]+)(")`)
	newContent := re.ReplaceAll(content, []byte(fmt.Sprintf(`${1}%s${3}`, version.String())))

	if string(newContent) == string(content) {
		return fmt.Errorf("no version field found in package.json")
	}

	return os.WriteFile(packageJSONPath, newContent, 0644)
}

// GetVersionFiles returns paths to all version-containing files
func (n *NPMEcosystem) GetVersionFiles() []string {
	packageJSONPath := filepath.Join(n.path, "package.json")

	if _, err := os.Stat(packageJSONPath); err == nil {
		return []string{"package.json"}
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
