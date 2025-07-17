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
	"github.com/NatoNathan/shipyard/pkg/shipment"
	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v3"
)

// ChangeType represents the type of change in a consignment
type ChangeType string

// GetAvailableChangeTypes returns the available change types from the project configuration
func (m *Manager) GetAvailableChangeTypes() []config.ChangeTypeConfig {
	return m.projectConfig.GetChangeTypes()
}

// GetChangeTypeConfig returns the configuration for a specific change type
func (m *Manager) GetChangeTypeConfig(changeTypeName string) *config.ChangeTypeConfig {
	return m.projectConfig.GetChangeTypeByName(changeTypeName)
}

// ValidateChangeType validates if a change type is supported
func (m *Manager) ValidateChangeType(changeTypeName string) error {
	if m.GetChangeTypeConfig(changeTypeName) == nil {
		availableTypes := m.projectConfig.GetChangeTypeNames()
		return fmt.Errorf("unsupported change type '%s'. Available types: %s", changeTypeName, strings.Join(availableTypes, ", "))
	}
	return nil
}

// Consignment represents a set of changes made to packages
type Consignment struct {
	ID       string            `json:"id"`
	Packages map[string]string `json:"packages"` // package name -> change type
	Summary  string            `json:"summary"`
	Created  time.Time         `json:"created"`
}

// Manager handles consignment operations
type Manager struct {
	projectConfig      *config.ProjectConfig
	consignmentDir     string
	shipmentHistoryDir string
}

// NewManager creates a new consignment manager
func NewManager(projectConfig *config.ProjectConfig) *Manager {
	return &Manager{
		projectConfig:      projectConfig,
		consignmentDir:     filepath.Join(".shipyard", "consignments"),
		shipmentHistoryDir: ".shipyard",
	}
}

// NewManagerWithDir creates a new consignment manager with a custom directory
func NewManagerWithDir(projectConfig *config.ProjectConfig, dir string) *Manager {
	return &Manager{
		projectConfig:      projectConfig,
		consignmentDir:     dir,
		shipmentHistoryDir: filepath.Dir(dir),
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
func (m *Manager) CreateConsignment(packages []string, changeType string, summary string) (*Consignment, error) {
	if err := m.EnsureConsignmentDir(); err != nil {
		return nil, fmt.Errorf("failed to create consignment directory: %w", err)
	}

	if len(packages) == 0 {
		return nil, fmt.Errorf("no packages specified")
	}

	if strings.TrimSpace(summary) == "" {
		return nil, fmt.Errorf("summary cannot be empty")
	}

	// Validate change type
	if err := m.ValidateChangeType(changeType); err != nil {
		return nil, err
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
		consignment.Packages[pkg] = changeType
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
	// Create frontmatter structure with all consignment metadata
	frontmatter := struct {
		ID       string            `yaml:"id"`
		Created  time.Time         `yaml:"created"`
		Packages map[string]string `yaml:"packages"`
	}{
		ID:       consignment.ID,
		Created:  consignment.Created,
		Packages: consignment.Packages,
	}

	// Marshal the frontmatter to YAML
	yamlData, err := yaml.Marshal(frontmatter)
	if err != nil {
		return fmt.Errorf("failed to marshal frontmatter to YAML: %w", err)
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
	// Create a structure for frontmatter that includes all consignment metadata
	frontMatterData := struct {
		ID       string            `yaml:"id"`
		Created  time.Time         `yaml:"created"`
		Packages map[string]string `yaml:"packages"`
	}{}

	// Parse frontmatter using bytes.NewReader to convert []byte to io.Reader
	summary, err := frontmatter.Parse(bytes.NewReader(content), &frontMatterData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	consignment := &Consignment{
		ID:       frontMatterData.ID,
		Created:  frontMatterData.Created,
		Packages: frontMatterData.Packages,
		Summary:  strings.TrimSpace(string(summary)),
	}

	if consignment.Packages == nil {
		consignment.Packages = make(map[string]string)
	}

	// If ID is missing, extract from filename as fallback
	if consignment.ID == "" {
		// This is for backward compatibility with old consignment files
		consignment.ID = generateConsignmentID()
	}

	// If Created is zero, use current time as fallback
	if consignment.Created.IsZero() {
		consignment.Created = time.Now()
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

// getShipmentHistory returns a shipment history instance using the manager's configuration
func (m *Manager) getShipmentHistory() *shipment.ShipmentHistory {
	historyFile := filepath.Join(m.shipmentHistoryDir, "shipment-history.json")
	return shipment.NewShipmentHistoryWithFile(m.projectConfig, historyFile)
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

	// Parse the current version from manifest
	manifestVersion, err := semver.Parse(ecosystemPkg.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse current version for package %s: %w", pkgName, err)
	}

	// Check shipment history to find the latest version for this package
	var currentVersion *semver.Version
	shipmentHistory := m.getShipmentHistory()
	history, err := shipmentHistory.LoadHistory()
	if err == nil && len(history) > 0 {
		// Find the latest version for this package in shipment history
		for _, s := range history {
			if version, exists := s.Versions[pkgName]; exists {
				if currentVersion == nil || version.GreaterThan(currentVersion) {
					currentVersion = version
				}
			}
		}
	}

	// If no version found in history, use manifest version
	if currentVersion == nil {
		currentVersion = manifestVersion
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
		if changeTypeName, exists := consignment.Packages[pkgName]; exists {
			changeTypeConfig := m.GetChangeTypeConfig(changeTypeName)
			if changeTypeConfig != nil {
				switch changeTypeConfig.SemverBump {
				case "major":
					hasMajor = true
				case "minor":
					hasMinor = true
				case "patch":
					hasPatch = true
				}
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

// convertToShipmentConsignments converts consignment.Consignment to shipment.Consignment
func convertToShipmentConsignments(consignments []*Consignment) []*shipment.Consignment {
	shipmentConsignments := make([]*shipment.Consignment, len(consignments))
	for i, c := range consignments {
		shipmentConsignments[i] = &shipment.Consignment{
			ID:       c.ID,
			Packages: c.Packages,
			Summary:  c.Summary,
			Created:  c.Created,
		}
	}
	return shipmentConsignments
}

// ApplyConsignments calculates and applies version updates to all packages, then clears consignments
func (m *Manager) ApplyConsignments() (map[string]*semver.Version, error) {
	return m.ApplyConsignmentsWithTemplate("")
}

// RecordShipmentHistoryOnly records the shipment history without applying version updates or clearing consignments
func (m *Manager) RecordShipmentHistoryOnly(templateName string) (map[string]*semver.Version, error) {
	return m.RecordShipmentHistoryWithTags(templateName, nil)
}

// RecordShipmentHistoryWithTags records the shipment history with git tags without applying version updates or clearing consignments
func (m *Manager) RecordShipmentHistoryWithTags(templateName string, gitTags map[string]string) (map[string]*semver.Version, error) {
	// Get current consignments
	consignments, err := m.GetConsignmens()
	if err != nil {
		return nil, fmt.Errorf("failed to get consignments: %w", err)
	}

	if len(consignments) == 0 {
		return make(map[string]*semver.Version), nil
	}

	// Calculate all versions
	versions, err := m.CalculateAllVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate versions: %w", err)
	}

	// Record shipment history
	shipmentHistory := m.getShipmentHistory()

	// Use provided template or fall back to project config
	template := templateName
	if template == "" {
		template = m.projectConfig.Changelog.Template
	}

	// Convert consignments to shipment format
	shipmentConsignments := convertToShipmentConsignments(consignments)
	_, err = shipmentHistory.RecordShipmentWithTags(shipmentConsignments, versions, template, gitTags)
	if err != nil {
		return nil, fmt.Errorf("failed to record shipment history: %w", err)
	}

	return versions, nil
}

// ApplyVersionUpdatesAndClearConsignments applies version updates and clears consignments
func (m *Manager) ApplyVersionUpdatesAndClearConsignments(versions map[string]*semver.Version) error {
	// Apply version updates
	for pkgName, version := range versions {
		if err := m.UpdatePackageVersion(pkgName, version); err != nil {
			return fmt.Errorf("failed to update version for package %s: %w", pkgName, err)
		}
	}

	// Clear consignments after successful application
	if err := m.ClearConsignments(); err != nil {
		return fmt.Errorf("failed to clear consignments: %w", err)
	}

	return nil
}

// ApplyConsignmentsWithTemplate calculates and applies version updates, records shipment history, then clears consignments
func (m *Manager) ApplyConsignmentsWithTemplate(templateName string) (map[string]*semver.Version, error) {
	return m.ApplyConsignmentsWithTemplateAndTags(templateName, nil)
}

// ApplyConsignmentsWithTemplateAndTags calculates and applies version updates, records shipment history with git tags, then clears consignments
func (m *Manager) ApplyConsignmentsWithTemplateAndTags(templateName string, gitTags map[string]string) (map[string]*semver.Version, error) {
	// Get current consignments before we clear them
	consignments, err := m.GetConsignmens()
	if err != nil {
		return nil, fmt.Errorf("failed to get consignments: %w", err)
	}

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
	// Record shipment history if we have consignments to ship
	if len(consignments) > 0 {
		shipmentHistory := m.getShipmentHistory()

		// Use provided template or fall back to project config
		template := templateName
		if template == "" {
			template = m.projectConfig.Changelog.Template
		}

		// Convert consignments to shipment format
		shipmentConsignments := convertToShipmentConsignments(consignments)
		_, err := shipmentHistory.RecordShipmentWithTags(shipmentConsignments, versions, template, gitTags)
		if err != nil {
			return nil, fmt.Errorf("failed to record shipment history: %w", err)
		}
	}

	// Clear consignments after successful application and history recording
	if err := m.ClearConsignments(); err != nil {
		return nil, fmt.Errorf("failed to clear consignments: %w", err)
	}

	return versions, nil
}
