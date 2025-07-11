// Package consignment provides functionality for managing consignments in Shipyard projects.
// A consignment represents a set of changes made to one or more packages and is used
// to track changes for release management.
package consignment

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/NatoNathan/shipyard/pkg/config"
	"github.com/NatoNathan/shipyard/pkg/handlers"
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v3"
)

// ChangeType represents the type of change in a consignment
type ChangeType string

const (
	// Patch represents bug fixes and minor updates that don't introduce breaking changes
	Patch ChangeType = "patch"
	// Minor represents new features that are backward compatible
	Minor ChangeType = "minor"
	// Major represents breaking changes
	Major ChangeType = "major"
)

// Consignment represents a set of changes made to packages
type Consignment struct {
	ID       string            `json:"id"`
	Packages map[string]string `json:"packages"` // package name -> change type
	Summary  string            `json:"summary"`
	Created  time.Time         `json:"created"`
}

// Manager handles consignment operations
type Manager struct {
	projectConfig  *config.ProjectConfig
	consignmentDir string
}

// NewManager creates a new consignment manager
func NewManager(projectConfig *config.ProjectConfig) *Manager {
	return &Manager{
		projectConfig:  projectConfig,
		consignmentDir: filepath.Join(".shipyard", "consignments"),
	}
}

// NewManagerWithDir creates a new consignment manager with a custom directory
func NewManagerWithDir(projectConfig *config.ProjectConfig, dir string) *Manager {
	return &Manager{
		projectConfig:  projectConfig,
		consignmentDir: dir,
	}
}

// EnsureConsignmentDir creates the consignment directory if it doesn't exist
func (m *Manager) EnsureConsignmentDir() error {
	return os.MkdirAll(m.consignmentDir, 0755)
}

// GetAvailablePackages returns the list of available packages based on the project configuration
func (m *Manager) GetAvailablePackages() []config.Package {
	if m.projectConfig.Type == config.RepositoryTypeMonorepo {
		return m.projectConfig.Packages
	}
	return []config.Package{m.projectConfig.Package}
}

// CreateConsignment creates a new consignment with the specified packages, change type, and summary
func (m *Manager) CreateConsignment(packages []string, changeType ChangeType, summary string) (*Consignment, error) {
	if err := m.EnsureConsignmentDir(); err != nil {
		return nil, fmt.Errorf("failed to create consignment directory: %w", err)
	}

	if len(packages) == 0 {
		return nil, fmt.Errorf("no packages specified")
	}

	if strings.TrimSpace(summary) == "" {
		return nil, fmt.Errorf("summary cannot be empty")
	}

	// Validate packages exist in configuration
	availablePackages := m.GetAvailablePackages()
	availablePackageNames := make(map[string]bool)
	for _, pkg := range availablePackages {
		availablePackageNames[pkg.Name] = true
	}

	for _, pkg := range packages {
		if !availablePackageNames[pkg] {
			return nil, fmt.Errorf("package %q not found in project configuration", pkg)
		}
	}

	// Create consignment
	consignment := &Consignment{
		ID:       generateConsignmentID(),
		Packages: make(map[string]string),
		Summary:  strings.TrimSpace(summary),
		Created:  time.Now(),
	}

	// Set change type for all packages
	for _, pkg := range packages {
		consignment.Packages[pkg] = string(changeType)
	}

	// Write consignment file
	filename := filepath.Join(m.consignmentDir, fmt.Sprintf("%s.md", consignment.ID))
	if err := m.writeConsignmentFile(filename, consignment); err != nil {
		return nil, fmt.Errorf("failed to write consignment file: %w", err)
	}

	return consignment, nil
}

// writeConsignmentFile writes the consignment to a markdown file
func (m *Manager) writeConsignmentFile(filename string, consignment *Consignment) error {
	// Marshal the packages map to proper YAML
	yamlData, err := yaml.Marshal(consignment.Packages)
	if err != nil {
		return fmt.Errorf("failed to marshal packages to YAML: %w", err)
	}

	var content strings.Builder

	// Write YAML frontmatter
	content.WriteString("---\n")
	content.Write(yamlData)
	content.WriteString("---\n\n")

	// Write the summary content
	content.WriteString(consignment.Summary)
	content.WriteString("\n")

	return os.WriteFile(filename, []byte(content.String()), 0644)
}

func (m *Manager) parseConsignment(content []byte) (*Consignment, error) {
	// Create a structure for frontmatter that matches the package map
	frontMatterData := make(map[string]string)

	// Parse frontmatter using bytes.NewReader to convert []byte to io.Reader
	summary, err := frontmatter.Parse(bytes.NewReader(content), &frontMatterData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	consignment := &Consignment{
		Packages: frontMatterData,
		Summary:  strings.TrimSpace(string(summary)),
	}

	if consignment.Packages == nil {
		consignment.Packages = make(map[string]string)
	}

	return consignment, nil
}

// generateConsignmentID generates a unique ID for the consignment
func generateConsignmentID() string {
	// Generate a random 8-character hex string
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// GetConsignmentDir returns the consignment directory path
func (m *Manager) GetConsignmentDir() string {
	return m.consignmentDir
}

func (m *Manager) GetConsignmens() ([]*Consignment, error) {
	files, err := os.ReadDir(m.consignmentDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read consignment directory: %w", err)
	}

	var consignments []*Consignment

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(m.consignmentDir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read consignment file %s: %w", file.Name(), err)
		}

		consignment, err := m.parseConsignment(content)
		if err != nil {
			return nil, fmt.Errorf("failed to parse consignment file %s: %w", file.Name(), err)
		}

		consignments = append(consignments, consignment)
	}

	return consignments, nil
}

func (m *Manager) GetConsignmentsForPackage(pkgName string) ([]*Consignment, error) {
	consignments, err := m.GetConsignmens()
	if err != nil {
		return nil, fmt.Errorf("failed to get consignments: %w", err)
	}

	var filtered []*Consignment
	for _, consignment := range consignments {
		if _, exists := consignment.Packages[pkgName]; exists {
			filtered = append(filtered, consignment)
		}
	}

	return filtered, nil
}

// CalculateNextVersion calculates the next version for a package based on its consignments
func (m *Manager) CalculateNextVersion(pkgName string) (*semver.Version, error) {
	// Get the package configuration
	pkg := m.projectConfig.GetPackageByName(pkgName)
	if pkg == nil {
		return nil, fmt.Errorf("package %q not found in project configuration", pkgName)
	}

	// Get the current version from the package handler
	handler, ok := handlers.GetHandler(pkg.Ecosystem)
	if !ok {
		return nil, fmt.Errorf("no handler found for ecosystem %s", pkg.Ecosystem)
	}

	ecosystemPkg, err := handler.LoadPackage(pkg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to load package %s: %w", pkgName, err)
	}

	// Parse the current version
	currentVersion, err := semver.Parse(ecosystemPkg.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse current version for package %s: %w", pkgName, err)
	}

	// Get consignments for this package
	consignments, err := m.GetConsignmentsForPackage(pkgName)
	if err != nil {
		return nil, fmt.Errorf("failed to get consignments for package %s: %w", pkgName, err)
	}

	// If no consignments, return current version
	if len(consignments) == 0 {
		return currentVersion, nil
	}

	// Sort consignments by creation time to ensure we process them in order
	sort.Slice(consignments, func(i, j int) bool {
		return consignments[i].Created.Before(consignments[j].Created)
	})

	// Calculate the next version based on the highest change type
	nextVersion := currentVersion.Copy()

	hasMajor := false
	hasMinor := false
	hasPatch := false

	for _, consignment := range consignments {
		if changeType, exists := consignment.Packages[pkgName]; exists {
			switch ChangeType(changeType) {
			case Major:
				hasMajor = true
			case Minor:
				hasMinor = true
			case Patch:
				hasPatch = true
			}
		}
	}

	// Apply version bump based on highest change type
	if hasMajor {
		nextVersion = nextVersion.BumpMajor()
	} else if hasMinor {
		nextVersion = nextVersion.BumpMinor()
	} else if hasPatch {
		nextVersion = nextVersion.BumpPatch()
	}

	return nextVersion, nil
}

// CalculateAllVersions calculates the next version for all packages in the project
func (m *Manager) CalculateAllVersions() (map[string]*semver.Version, error) {
	packages := m.GetAvailablePackages()
	versions := make(map[string]*semver.Version)

	for _, pkg := range packages {
		version, err := m.CalculateNextVersion(pkg.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate version for package %s: %w", pkg.Name, err)
		}
		versions[pkg.Name] = version
	}

	return versions, nil
}

// UpdatePackageVersion updates the version of a package using its ecosystem handler
func (m *Manager) UpdatePackageVersion(pkgName string, version *semver.Version) error {
	// Get the package configuration
	pkg := m.projectConfig.GetPackageByName(pkgName)
	if pkg == nil {
		return fmt.Errorf("package %q not found in project configuration", pkgName)
	}

	// Get the handler for this ecosystem
	handler, ok := handlers.GetHandler(pkg.Ecosystem)
	if !ok {
		return fmt.Errorf("no handler found for ecosystem %s", pkg.Ecosystem)
	}

	// Update the version
	return handler.UpdateVersion(pkg.Path, version.String())
}

// ClearConsignments removes all consignment files from the consignment directory
func (m *Manager) ClearConsignments() error {
	files, err := os.ReadDir(m.consignmentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Directory doesn't exist, nothing to clear
		}
		return fmt.Errorf("failed to read consignment directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(m.consignmentDir, file.Name())
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("failed to remove consignment file %s: %w", filePath, err)
		}
	}

	return nil
}

// ApplyConsignments calculates and applies version updates to all packages, then clears consignments
func (m *Manager) ApplyConsignments() (map[string]*semver.Version, error) {
	// Calculate all versions first
	versions, err := m.CalculateAllVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate versions: %w", err)
	}

	// Apply version updates
	for pkgName, version := range versions {
		if err := m.UpdatePackageVersion(pkgName, version); err != nil {
			return nil, fmt.Errorf("failed to update version for package %s: %w", pkgName, err)
		}
	}

	// Clear consignments after successful application
	if err := m.ClearConsignments(); err != nil {
		return nil, fmt.Errorf("failed to clear consignments: %w", err)
	}

	return versions, nil
}
