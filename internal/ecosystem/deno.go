package ecosystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/NatoNathan/shipyard/pkg/semver"
)

// DenoEcosystem handles version management for Deno projects
type DenoEcosystem struct {
	path string
}

// NewDenoEcosystem creates a new Deno ecosystem handler
func NewDenoEcosystem(path string) *DenoEcosystem {
	return &DenoEcosystem{path: path}
}

// DenoConfig represents the structure of deno.json/deno.jsonc
type DenoConfig struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version"`
	Exports string `json:"exports,omitempty"`
}

// ReadVersion reads the current version from deno.json or deno.jsonc
func (d *DenoEcosystem) ReadVersion() (semver.Version, error) {
	// Try deno.json first
	denoPath := filepath.Join(d.path, "deno.json")
	if _, err := os.Stat(denoPath); os.IsNotExist(err) {
		// Try deno.jsonc
		denoPath = filepath.Join(d.path, "deno.jsonc")
		if _, err := os.Stat(denoPath); os.IsNotExist(err) {
			return semver.Version{}, fmt.Errorf("no deno.json or deno.jsonc found")
		}
	}

	content, err := os.ReadFile(denoPath)
	if err != nil {
		return semver.Version{}, fmt.Errorf("failed to read %s: %w", filepath.Base(denoPath), err)
	}

	// Strip comments for JSONC support
	content = stripJSONComments(content)

	var config DenoConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return semver.Version{}, fmt.Errorf("failed to parse %s: %w", filepath.Base(denoPath), err)
	}

	if config.Version == "" {
		return semver.Version{}, fmt.Errorf("no version field found in %s", filepath.Base(denoPath))
	}

	return semver.Parse(config.Version)
}

// UpdateVersion updates the version in deno.json or deno.jsonc
func (d *DenoEcosystem) UpdateVersion(version semver.Version) error {
	// Determine which config file exists
	denoPath := filepath.Join(d.path, "deno.json")
	if _, err := os.Stat(denoPath); os.IsNotExist(err) {
		denoPath = filepath.Join(d.path, "deno.jsonc")
		if _, err := os.Stat(denoPath); os.IsNotExist(err) {
			return fmt.Errorf("no deno.json or deno.jsonc found")
		}
	}

	// Read existing content
	content, err := os.ReadFile(denoPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", filepath.Base(denoPath), err)
	}

	// Strip comments for parsing
	cleanContent := stripJSONComments(content)

	// Parse as generic map to preserve structure
	var data map[string]interface{}
	if err := json.Unmarshal(cleanContent, &data); err != nil {
		return fmt.Errorf("failed to parse %s: %w", filepath.Base(denoPath), err)
	}

	// Update version
	data["version"] = version.String()

	// Write back with pretty formatting
	newContent, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode %s: %w", filepath.Base(denoPath), err)
	}

	return os.WriteFile(denoPath, newContent, 0644)
}

// GetVersionFiles returns paths to all version-containing files
func (d *DenoEcosystem) GetVersionFiles() []string {
	// Check for deno.json
	denoPath := filepath.Join(d.path, "deno.json")
	if _, err := os.Stat(denoPath); err == nil {
		return []string{"deno.json"}
	}

	// Check for deno.jsonc
	denoPath = filepath.Join(d.path, "deno.jsonc")
	if _, err := os.Stat(denoPath); err == nil {
		return []string{"deno.jsonc"}
	}

	return []string{}
}

// DetectDenoEcosystem checks if a directory contains a Deno project
func DetectDenoEcosystem(path string) bool {
	// Check for deno.json
	if _, err := os.Stat(filepath.Join(path, "deno.json")); err == nil {
		return true
	}

	// Check for deno.jsonc
	if _, err := os.Stat(filepath.Join(path, "deno.jsonc")); err == nil {
		return true
	}

	// Check for mod.ts (common Deno entry point)
	if _, err := os.Stat(filepath.Join(path, "mod.ts")); err == nil {
		return true
	}

	return false
}

// stripJSONComments removes single-line and multi-line comments from JSON content
// to support JSONC (JSON with Comments) format used by Deno
func stripJSONComments(content []byte) []byte {
	// Remove single-line comments (// ...)
	re := regexp.MustCompile(`//[^\n]*`)
	content = re.ReplaceAll(content, []byte{})

	// Remove multi-line comments (/* ... */)
	re = regexp.MustCompile(`/\*[\s\S]*?\*/`)
	content = re.ReplaceAll(content, []byte{})

	return content
}
