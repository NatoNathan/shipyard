package ecosystem

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/NatoNathan/shipyard/pkg/semver"
)

var _ Handler = (*CargoEcosystem)(nil)

// CargoEcosystem handles version management for Rust/Cargo projects
type CargoEcosystem struct {
	path string
}

// NewCargoEcosystem creates a new Cargo ecosystem handler
func NewCargoEcosystem(path string) *CargoEcosystem {
	return &CargoEcosystem{path: path}
}

// CargoPackage represents the [package] section of Cargo.toml
type CargoPackage struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
	Edition string `toml:"edition,omitempty"`
}

// CargoManifest represents the structure of Cargo.toml
type CargoManifest struct {
	Package CargoPackage `toml:"package"`
}

// ReadVersion reads the current version from Cargo.toml
func (c *CargoEcosystem) ReadVersion() (semver.Version, error) {
	cargoPath := filepath.Join(c.path, "Cargo.toml")

	content, err := os.ReadFile(cargoPath)
	if err != nil {
		return semver.Version{}, fmt.Errorf("failed to read Cargo.toml: %w", err)
	}

	var manifest CargoManifest
	if err := toml.Unmarshal(content, &manifest); err != nil {
		return semver.Version{}, fmt.Errorf("failed to parse Cargo.toml: %w", err)
	}

	if manifest.Package.Version == "" {
		return semver.Version{}, fmt.Errorf("no version field found in Cargo.toml [package] section")
	}

	return semver.Parse(manifest.Package.Version)
}

// UpdateVersion updates the version in Cargo.toml using regex replacement
// within the [package] section only, preserving TOML comments and formatting.
func (c *CargoEcosystem) UpdateVersion(version semver.Version) error {
	cargoPath := filepath.Join(c.path, "Cargo.toml")

	// Read existing content
	content, err := os.ReadFile(cargoPath)
	if err != nil {
		return fmt.Errorf("failed to read Cargo.toml: %w", err)
	}

	contentStr := string(content)

	// Find the [package] section
	packageIdx := strings.Index(contentStr, "[package]")
	if packageIdx == -1 {
		return fmt.Errorf("no [package] section found in Cargo.toml")
	}

	// Find the end of the [package] section (next [section] header or EOF)
	packageContent := contentStr[packageIdx:]
	nextSectionRe := regexp.MustCompile(`\n\[`)
	nextSectionLoc := nextSectionRe.FindStringIndex(packageContent[len("[package]"):])

	var packageEnd int
	if nextSectionLoc != nil {
		packageEnd = packageIdx + len("[package]") + nextSectionLoc[0]
	} else {
		packageEnd = len(contentStr)
	}

	// Replace version within the [package] section only
	packageSection := contentStr[packageIdx:packageEnd]
	versionRe := regexp.MustCompile(`(version\s*=\s*")([^"]+)(")`)
	newPackageSection := versionRe.ReplaceAllString(packageSection, fmt.Sprintf(`${1}%s${3}`, version.String()))

	if newPackageSection == packageSection {
		return fmt.Errorf("no version field found in [package] section of Cargo.toml")
	}

	newContent := contentStr[:packageIdx] + newPackageSection + contentStr[packageEnd:]

	return os.WriteFile(cargoPath, []byte(newContent), 0644)
}

// GetVersionFiles returns paths to all version-containing files
func (c *CargoEcosystem) GetVersionFiles() []string {
	cargoPath := filepath.Join(c.path, "Cargo.toml")
	if _, err := os.Stat(cargoPath); err == nil {
		return []string{"Cargo.toml"}
	}
	return []string{}
}

// DetectCargoEcosystem checks if a directory contains a Rust/Cargo project
func DetectCargoEcosystem(path string) bool {
	// Check for Cargo.toml
	if _, err := os.Stat(filepath.Join(path, "Cargo.toml")); err == nil {
		return true
	}

	// Check for src directory with .rs files
	srcDir := filepath.Join(path, "src")
	if entries, err := os.ReadDir(srcDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".rs") {
				return true
			}
		}
	}

	return false
}
