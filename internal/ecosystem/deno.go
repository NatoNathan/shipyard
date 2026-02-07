package ecosystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/NatoNathan/shipyard/pkg/semver"
)

var _ Handler = (*DenoEcosystem)(nil)

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

// UpdateVersion updates the version in deno.json or deno.jsonc using regex
// replacement to preserve original formatting and comments.
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

	// Use regex to replace only the version value, preserving all formatting and comments
	re := regexp.MustCompile(`("version"\s*:\s*")([^"]+)(")`)
	newContent := re.ReplaceAll(content, []byte(fmt.Sprintf(`${1}%s${3}`, version.String())))

	if string(newContent) == string(content) {
		return fmt.Errorf("no version field found in %s", filepath.Base(denoPath))
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

// stripJSONComments removes single-line (//) and multi-line (/* */) comments
// from JSONC content using a state-machine parser. This correctly handles
// comments inside strings (e.g., URLs containing "//") by tracking whether
// we are inside a JSON string literal.
func stripJSONComments(content []byte) []byte {
	result := make([]byte, 0, len(content))
	i := 0
	n := len(content)

	for i < n {
		ch := content[i]

		// Handle string literals - pass through unchanged
		if ch == '"' {
			result = append(result, ch)
			i++
			for i < n {
				sch := content[i]
				result = append(result, sch)
				if sch == '\\' {
					// Skip escaped character
					i++
					if i < n {
						result = append(result, content[i])
						i++
					}
					continue
				}
				if sch == '"' {
					i++
					break
				}
				i++
			}
			continue
		}

		// Check for comments (only outside strings)
		if ch == '/' && i+1 < n {
			next := content[i+1]

			// Single-line comment: skip until end of line
			if next == '/' {
				i += 2
				for i < n && content[i] != '\n' {
					i++
				}
				continue
			}

			// Multi-line comment: skip until */
			if next == '*' {
				i += 2
				for i+1 < n {
					if content[i] == '*' && content[i+1] == '/' {
						i += 2
						break
					}
					i++
				}
				continue
			}
		}

		result = append(result, ch)
		i++
	}

	return result
}
