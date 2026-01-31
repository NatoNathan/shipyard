package ecosystem

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/NatoNathan/shipyard/pkg/semver"
)

// GoEcosystem handles version management for Go projects
type GoEcosystem struct {
	path    string
	options *GoEcosystemOptions
}

// GoEcosystemOptions configures Go ecosystem behavior
type GoEcosystemOptions struct {
	TagOnly bool // If true, only create git tags without updating version files
}

// NewGoEcosystem creates a new Go ecosystem handler with default options
func NewGoEcosystem(path string) *GoEcosystem {
	return &GoEcosystem{
		path:    path,
		options: &GoEcosystemOptions{TagOnly: false},
	}
}

// NewGoEcosystemWithOptions creates a new Go ecosystem handler with custom options
func NewGoEcosystemWithOptions(path string, options *GoEcosystemOptions) *GoEcosystem {
	if options == nil {
		options = &GoEcosystemOptions{TagOnly: false}
	}
	return &GoEcosystem{
		path:    path,
		options: options,
	}
}

// ReadVersion reads the current version from Go version files
// Checks in order: version.go, go.mod
func (g *GoEcosystem) ReadVersion() (semver.Version, error) {
	// Try version.go first
	versionGoPath := filepath.Join(g.path, "version.go")
	if _, err := os.Stat(versionGoPath); err == nil {
		return g.readVersionFromVersionGo(versionGoPath)
	}

	// Try go.mod
	goModPath := filepath.Join(g.path, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		return g.readVersionFromGoMod(goModPath)
	}

	return semver.Version{}, fmt.Errorf("no version file found in Go project at %s", g.path)
}

// UpdateVersion updates the version in Go version files
func (g *GoEcosystem) UpdateVersion(version semver.Version) error {
	// In tag-only mode, skip file updates
	if g.options != nil && g.options.TagOnly {
		return nil
	}

	updated := false

	// Update version.go if it exists
	versionGoPath := filepath.Join(g.path, "version.go")
	if _, err := os.Stat(versionGoPath); err == nil {
		if err := g.updateVersionGo(versionGoPath, version); err != nil {
			return err
		}
		updated = true
	}

	// Update go.mod if it exists and contains a version comment
	goModPath := filepath.Join(g.path, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		if err := g.updateGoMod(goModPath, version); err != nil {
			return err
		}
		updated = true
	}

	if !updated {
		return fmt.Errorf("no version files to update in Go project at %s", g.path)
	}

	return nil
}

// GetVersionFiles returns paths to all version-containing files
func (g *GoEcosystem) GetVersionFiles() []string {
	// In tag-only mode, return empty list (no files to update)
	if g.options != nil && g.options.TagOnly {
		return []string{}
	}

	var files []string

	versionGoPath := filepath.Join(g.path, "version.go")
	if _, err := os.Stat(versionGoPath); err == nil {
		files = append(files, versionGoPath)
	}

	goModPath := filepath.Join(g.path, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		files = append(files, goModPath)
	}

	return files
}

// readVersionFromVersionGo extracts version from version.go file
// Looks for patterns like: const Version = "1.2.3" or var Version = "1.2.3"
func (g *GoEcosystem) readVersionFromVersionGo(path string) (semver.Version, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return semver.Version{}, fmt.Errorf("failed to read version.go: %w", err)
	}

	// Match: const Version = "1.2.3" or var Version = "1.2.3"
	re := regexp.MustCompile(`(?m)^\s*(?:const|var)\s+Version\s*=\s*["']([0-9]+\.[0-9]+\.[0-9]+)["']`)
	matches := re.FindSubmatch(content)

	if len(matches) < 2 {
		return semver.Version{}, fmt.Errorf("no version found in %s", path)
	}

	return semver.Parse(string(matches[1]))
}

// readVersionFromGoMod extracts version from go.mod comment
// Looks for: // version: 1.2.3
func (g *GoEcosystem) readVersionFromGoMod(path string) (semver.Version, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return semver.Version{}, fmt.Errorf("failed to read go.mod: %w", err)
	}

	// Match: // version: 1.2.3
	re := regexp.MustCompile(`(?m)^//\s*version:\s*([0-9]+\.[0-9]+\.[0-9]+)`)
	matches := re.FindSubmatch(content)

	if len(matches) < 2 {
		return semver.Version{}, fmt.Errorf("no version comment found in %s", path)
	}

	return semver.Parse(string(matches[1]))
}

// updateVersionGo updates the version in version.go
func (g *GoEcosystem) updateVersionGo(path string, version semver.Version) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read version.go: %w", err)
	}

	// Replace: const/var Version = "old" with const/var Version = "new"
	re := regexp.MustCompile(`((?:const|var)\s+Version\s*=\s*)["']([0-9]+\.[0-9]+\.[0-9]+)["']`)
	newContent := re.ReplaceAll(content, []byte(fmt.Sprintf(`${1}"%s"`, version.String())))

	if string(newContent) == string(content) {
		return fmt.Errorf("no version declaration found in %s", path)
	}

	return os.WriteFile(path, newContent, 0644)
}

// updateGoMod updates the version comment in go.mod
func (g *GoEcosystem) updateGoMod(path string, version semver.Version) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read go.mod: %w", err)
	}

	// Replace: // version: old with // version: new
	re := regexp.MustCompile(`(//\s*version:\s*)([0-9]+\.[0-9]+\.[0-9]+)`)

	if !re.Match(content) {
		// Version comment doesn't exist, don't add it
		return nil
	}

	newContent := re.ReplaceAll(content, []byte(fmt.Sprintf(`${1}%s`, version.String())))

	return os.WriteFile(path, newContent, 0644)
}

// DetectGoEcosystem checks if a directory contains a Go project
func DetectGoEcosystem(path string) bool {
	// Check for go.mod
	if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
		return true
	}

	// Check for version.go
	if _, err := os.Stat(filepath.Join(path, "version.go")); err == nil {
		return true
	}

	// Check for any .go files
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			return true
		}
	}

	return false
}
