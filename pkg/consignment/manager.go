// Package consignment provides functionality for managing consignments in Shipyard projects.
// A consignment represents a set of changes made to one or more packages and is used
// to track changes for release management.
package consignment

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/NatoNathan/shipyard/pkg/config"
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
	var content strings.Builder

	content.WriteString("---\n")

	// Write package changes
	for pkg, changeType := range consignment.Packages {
		content.WriteString(fmt.Sprintf("\"%s\": %s\n", pkg, changeType))
	}

	content.WriteString("---\n\n")
	content.WriteString(consignment.Summary)
	content.WriteString("\n")

	return os.WriteFile(filename, []byte(content.String()), 0644)
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
