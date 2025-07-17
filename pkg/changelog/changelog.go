// Package changelog provides functionality for generating changelogs from consignments.
// It supports various changelog formats and templates.
package changelog

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/NatoNathan/shipyard/pkg/config"
	"github.com/NatoNathan/shipyard/pkg/consignment"
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/NatoNathan/shipyard/pkg/shipment"
	"github.com/NatoNathan/shipyard/pkg/templates"
)

// ChangelogEntry represents a single changelog entry
type ChangelogEntry struct {
	Version     string
	Date        time.Time
	Changes     map[string][]ChangelogChange // change type -> list of changes
	PackageName string                       // for monorepo projects
}

// ChangelogChange represents a single change in a changelog entry
type ChangelogChange struct {
	Summary     string
	PackageName string                   // for monorepo projects
	Consignment *consignment.Consignment // reference to the original consignment
}

// Generator handles changelog generation
type Generator struct {
	projectConfig  *config.ProjectConfig
	templateEngine *templates.TemplateEngine
}

// NewGenerator creates a new changelog generator
func NewGenerator(projectConfig *config.ProjectConfig) (*Generator, error) {
	templateEngine := templates.NewTemplateEngine(projectConfig)

	return &Generator{
		projectConfig:  projectConfig,
		templateEngine: templateEngine,
	}, nil
}

// GenerateChangelog generates a changelog from consignments
func (g *Generator) GenerateChangelog(consignments []*consignment.Consignment, versions map[string]*semver.Version) (string, error) {
	if len(consignments) == 0 {
		return "", fmt.Errorf("no consignments provided")
	}

	// For monorepo, generate combined changelog with all packages
	// Use GenerateChangelogsForPackages for separate package changelogs
	if g.projectConfig.Type == config.RepositoryTypeMonorepo {
		// Group consignments by package and version
		changelogEntries := g.groupConsignmentsByPackage(consignments, versions)

		// Convert to template entries
		templateEntries := make([]templates.ChangelogEntry, len(changelogEntries))
		for i, entry := range changelogEntries {
			templateEntries[i] = templates.ChangelogEntry{
				Version:      entry.Version,
				Date:         entry.Date.Format("2006-01-02"),
				DateTime:     entry.Date.Format(time.RFC3339),
				ShipmentDate: entry.Date.Format("2006-01-02"),
				Changes:      g.convertChanges(entry.Changes),
				PackageName:  entry.PackageName,
				ShipmentID:   "", // Not available for consignments
			}
		}

		// Sort entries by version (newest first)
		sort.Slice(templateEntries, func(i, j int) bool {
			versionI, _ := semver.Parse(templateEntries[i].Version)
			versionJ, _ := semver.Parse(templateEntries[j].Version)
			if versionI != nil && versionJ != nil {
				return versionI.GreaterThan(versionJ)
			}
			// Fallback to date comparison if version parsing fails
			dateI, _ := time.Parse("2006-01-02", templateEntries[i].Date)
			dateJ, _ := time.Parse("2006-01-02", templateEntries[j].Date)
			return dateI.After(dateJ)
		})

		// Generate changelog using the template engine
		return g.templateEngine.RenderChangelogTemplate(templateEntries, g.projectConfig.Changelog.Template)
	}

	// For single repo, use the original logic
	changelogEntries := g.groupConsignmentsByPackage(consignments, versions)

	// Convert to template entries
	templateEntries := make([]templates.ChangelogEntry, len(changelogEntries))
	for i, entry := range changelogEntries {
		templateEntries[i] = templates.ChangelogEntry{
			Version:      entry.Version,
			Date:         entry.Date.Format("2006-01-02"),
			DateTime:     entry.Date.Format(time.RFC3339),
			ShipmentDate: entry.Date.Format("2006-01-02"),
			Changes:      g.convertChanges(entry.Changes),
			PackageName:  entry.PackageName,
			ShipmentID:   "", // Not available for consignments
		}
	}

	// Sort entries by version (newest first)
	sort.Slice(templateEntries, func(i, j int) bool {
		versionI, _ := semver.Parse(templateEntries[i].Version)
		versionJ, _ := semver.Parse(templateEntries[j].Version)
		if versionI != nil && versionJ != nil {
			return versionI.GreaterThan(versionJ)
		}
		// Fallback to date comparison if version parsing fails
		dateI, _ := time.Parse("2006-01-02", templateEntries[i].Date)
		dateJ, _ := time.Parse("2006-01-02", templateEntries[j].Date)
		return dateI.After(dateJ)
	})

	// Generate changelog using the template engine
	return g.templateEngine.RenderChangelogTemplate(templateEntries, g.projectConfig.Changelog.Template)
}

// GenerateChangelogsForPackages generates separate changelogs for each package in a monorepo
func (g *Generator) GenerateChangelogsForPackages(consignments []*consignment.Consignment, versions map[string]*semver.Version) (map[string]string, error) {
	if len(consignments) == 0 {
		return nil, fmt.Errorf("no consignments provided")
	}

	changelogs := make(map[string]string)

	// Generate a separate changelog for each package
	for packageName, version := range versions {
		changelogContent, err := g.GenerateChangelogForPackage(packageName, consignments, version)
		if err != nil {
			return nil, fmt.Errorf("failed to generate changelog for package %s: %w", packageName, err)
		}
		changelogs[packageName] = changelogContent
	}

	return changelogs, nil
}

// convertChanges converts changelog changes to template changes
func (g *Generator) convertChanges(changes map[string][]ChangelogChange) map[string][]templates.ChangelogChange {
	templateChanges := make(map[string][]templates.ChangelogChange)
	for section, changeList := range changes {
		templateChangeList := make([]templates.ChangelogChange, len(changeList))
		for i, change := range changeList {
			// Get change type from consignment if available
			changeType := ""
			if change.Consignment != nil {
				for _, ct := range change.Consignment.Packages {
					changeType = ct
					break // Use the first one found
				}
			}

			templateChangeList[i] = templates.ChangelogChange{
				Summary:     change.Summary,
				ChangeType:  changeType,
				Section:     section,
				PackageName: change.PackageName,
			}
		}
		templateChanges[section] = templateChangeList
	}
	return templateChanges
}

// GenerateChangelogForPackage generates a changelog for a specific package
func (g *Generator) GenerateChangelogForPackage(packageName string, consignments []*consignment.Consignment, version *semver.Version) (string, error) {
	// Filter consignments for this package
	var filteredConsignments []*consignment.Consignment
	for _, c := range consignments {
		if _, exists := c.Packages[packageName]; exists {
			filteredConsignments = append(filteredConsignments, c)
		}
	}

	if len(filteredConsignments) == 0 {
		return "", fmt.Errorf("no consignments found for package %s", packageName)
	}

	// Create version map
	versions := map[string]*semver.Version{
		packageName: version,
	}

	return g.GenerateChangelog(filteredConsignments, versions)
}

// groupConsignmentsByPackage groups consignments by package and creates changelog entries
func (g *Generator) groupConsignmentsByPackage(consignments []*consignment.Consignment, versions map[string]*semver.Version) []ChangelogEntry {
	var entries []ChangelogEntry

	if g.projectConfig.Type == config.RepositoryTypeMonorepo {
		// For monorepo, create entries by package
		for packageName, version := range versions {
			entry := g.createEntryForPackage(packageName, version, consignments)
			if len(entry.Changes) > 0 {
				entries = append(entries, entry)
			}
		}
	} else {
		// For single repo, create one entry
		packageName := g.projectConfig.Package.Name
		if version, exists := versions[packageName]; exists {
			entry := g.createEntryForPackage(packageName, version, consignments)
			if len(entry.Changes) > 0 {
				entries = append(entries, entry)
			}
		}
	}

	return entries
}

// createEntryForPackage creates a changelog entry for a specific package
func (g *Generator) createEntryForPackage(packageName string, version *semver.Version, consignments []*consignment.Consignment) ChangelogEntry {
	entry := ChangelogEntry{
		Version:     version.String(),
		Date:        time.Now(),
		Changes:     make(map[string][]ChangelogChange),
		PackageName: packageName,
	}

	// Process consignments for this package
	for _, c := range consignments {
		if changeType, exists := c.Packages[packageName]; exists {
			// Map change type to changelog section
			section := g.mapChangeTypeToSection(changeType)

			change := ChangelogChange{
				Summary:     c.Summary,
				PackageName: packageName,
				Consignment: c,
			}

			entry.Changes[section] = append(entry.Changes[section], change)
		}
	}

	return entry
}

// mapChangeTypeToSection maps a change type to a changelog section
func (g *Generator) mapChangeTypeToSection(changeType string) string {
	// Look up the change type in the configuration
	for _, ct := range g.projectConfig.ChangeTypes {
		if ct.Name == changeType {
			// Use the configured section if available, otherwise use display name or name
			if ct.Section != "" {
				return ct.Section
			}
			if ct.DisplayName != "" {
				return ct.DisplayName
			}
			return ct.Name
		}
	}

	// Fallback for unknown change types
	return "Changed"
}

// GetAvailableTemplates returns a list of available changelog templates
func GetAvailableTemplates() []string {
	return []string{
		"keepachangelog",
		"conventional",
		"simple",
	}
}

// ValidateTemplate validates if a template name is supported
func ValidateTemplate(templateName string) error {
	available := GetAvailableTemplates()
	for _, name := range available {
		if name == templateName {
			return nil
		}
	}
	return fmt.Errorf("unsupported changelog template: %s. Available templates: %s",
		templateName, strings.Join(available, ", "))
}

// GenerateChangelogFromHistory generates a changelog from shipment history
func (g *Generator) GenerateChangelogFromHistory() (string, error) {
	shipmentHistory := shipment.NewShipmentHistory(g.projectConfig)
	history, err := shipmentHistory.LoadHistory()
	if err != nil {
		return "", fmt.Errorf("failed to load shipment history: %w", err)
	}

	if len(history) == 0 {
		return "", fmt.Errorf("no shipment history found")
	}

	// Convert all shipments to changelog entries
	var allEntries []ChangelogEntry
	for _, ship := range history {
		entries := g.convertShipmentToEntries(ship)
		allEntries = append(allEntries, entries...)
	}

	// Convert to template entries
	templateEntries := make([]templates.ChangelogEntry, len(allEntries))
	for i, entry := range allEntries {
		templateEntries[i] = templates.ChangelogEntry{
			Version:      entry.Version,
			Date:         entry.Date.Format("2006-01-02"),
			DateTime:     entry.Date.Format(time.RFC3339),
			ShipmentDate: entry.Date.Format("2006-01-02"),
			Changes:      g.convertChanges(entry.Changes),
			PackageName:  entry.PackageName,
			ShipmentID:   "", // Could be enhanced to track shipment ID
		}
	}

	// Sort entries by version (newest first)
	sort.Slice(templateEntries, func(i, j int) bool {
		versionI, _ := semver.Parse(templateEntries[i].Version)
		versionJ, _ := semver.Parse(templateEntries[j].Version)
		if versionI != nil && versionJ != nil {
			return versionI.GreaterThan(versionJ)
		}
		// Fallback to date comparison if version parsing fails
		dateI, _ := time.Parse("2006-01-02", templateEntries[i].Date)
		dateJ, _ := time.Parse("2006-01-02", templateEntries[j].Date)
		return dateI.After(dateJ)
	})

	// Generate changelog using the template engine
	return g.templateEngine.RenderChangelogTemplate(templateEntries, g.projectConfig.Changelog.Template)
}

// GenerateChangelogsFromHistoryForPackages generates separate changelogs for each package from shipment history
func (g *Generator) GenerateChangelogsFromHistoryForPackages() (map[string]string, error) {
	if g.projectConfig.Type != config.RepositoryTypeMonorepo {
		return nil, fmt.Errorf("this function is only for monorepo projects")
	}

	shipmentHistory := shipment.NewShipmentHistory(g.projectConfig)
	history, err := shipmentHistory.LoadHistory()
	if err != nil {
		return nil, fmt.Errorf("failed to load shipment history: %w", err)
	}

	if len(history) == 0 {
		return nil, fmt.Errorf("no shipment history found")
	}

	// Get all package names from the project config
	packageNames := g.projectConfig.GetPackageNames()
	if len(packageNames) == 0 {
		return nil, fmt.Errorf("no packages found in monorepo configuration")
	}

	// Generate a separate changelog for each package
	changelogs := make(map[string]string)
	for _, packageName := range packageNames {
		changelogContent, err := g.GenerateChangelogFromHistoryForPackage(packageName)
		if err != nil {
			// Log the error but continue with other packages
			continue
		}
		changelogs[packageName] = changelogContent
	}

	if len(changelogs) == 0 {
		return nil, fmt.Errorf("no changelogs could be generated for any package")
	}

	return changelogs, nil
}

// GenerateChangelogFromHistoryForPackage generates a changelog for a specific package from shipment history
func (g *Generator) GenerateChangelogFromHistoryForPackage(packageName string) (string, error) {
	shipmentHistory := shipment.NewShipmentHistory(g.projectConfig)
	shipments, err := shipmentHistory.GetShipmentsForPackage(packageName)
	if err != nil {
		return "", fmt.Errorf("failed to load shipment history for package %s: %w", packageName, err)
	}

	if len(shipments) == 0 {
		return "", fmt.Errorf("no shipment history found for package %s", packageName)
	}

	// Convert shipments to changelog entries for this package
	var allEntries []ChangelogEntry
	for _, ship := range shipments {
		entries := g.convertShipmentToEntriesForPackage(ship, packageName)
		allEntries = append(allEntries, entries...)
	}

	// Convert to template entries
	templateEntries := make([]templates.ChangelogEntry, len(allEntries))
	for i, entry := range allEntries {
		templateEntries[i] = templates.ChangelogEntry{
			Version:      entry.Version,
			Date:         entry.Date.Format("2006-01-02"),
			DateTime:     entry.Date.Format(time.RFC3339),
			ShipmentDate: entry.Date.Format("2006-01-02"),
			Changes:      g.convertChanges(entry.Changes),
			PackageName:  entry.PackageName,
			ShipmentID:   "", // Could be enhanced to track shipment ID
		}
	}

	// Sort entries by version (newest first)
	sort.Slice(templateEntries, func(i, j int) bool {
		versionI, _ := semver.Parse(templateEntries[i].Version)
		versionJ, _ := semver.Parse(templateEntries[j].Version)
		if versionI != nil && versionJ != nil {
			return versionI.GreaterThan(versionJ)
		}
		// Fallback to date comparison if version parsing fails
		dateI, _ := time.Parse("2006-01-02", templateEntries[i].Date)
		dateJ, _ := time.Parse("2006-01-02", templateEntries[j].Date)
		return dateI.After(dateJ)
	})

	// Generate changelog using the template engine
	return g.templateEngine.RenderChangelogTemplate(templateEntries, g.projectConfig.Changelog.Template)
}

// convertShipmentToEntries converts a shipment to changelog entries for all packages
func (g *Generator) convertShipmentToEntries(ship *shipment.Shipment) []ChangelogEntry {
	var entries []ChangelogEntry

	if g.projectConfig.Type == config.RepositoryTypeMonorepo {
		// For monorepo, create entries by package
		for packageName, version := range ship.Versions {
			entry := g.createEntryForPackageFromShipment(packageName, version, ship.Consignments, ship.Date)
			if len(entry.Changes) > 0 {
				entries = append(entries, entry)
			}
		}
	} else {
		// For single repo, create one entry
		packageName := g.projectConfig.Package.Name
		if version, exists := ship.Versions[packageName]; exists {
			entry := g.createEntryForPackageFromShipment(packageName, version, ship.Consignments, ship.Date)
			if len(entry.Changes) > 0 {
				entries = append(entries, entry)
			}
		}
	}

	return entries
}

// convertShipmentToEntriesForPackage converts a shipment to changelog entries for a specific package
func (g *Generator) convertShipmentToEntriesForPackage(ship *shipment.Shipment, packageName string) []ChangelogEntry {
	var entries []ChangelogEntry

	if version, exists := ship.Versions[packageName]; exists {
		entry := g.createEntryForPackageFromShipment(packageName, version, ship.Consignments, ship.Date)
		if len(entry.Changes) > 0 {
			entries = append(entries, entry)
		}
	}

	return entries
}

// createEntryForPackageFromShipment creates a changelog entry for a specific package from shipment consignments
func (g *Generator) createEntryForPackageFromShipment(packageName string, version *semver.Version, consignments []*shipment.Consignment, date time.Time) ChangelogEntry {
	entry := ChangelogEntry{
		Version:     version.String(),
		Date:        date,
		Changes:     make(map[string][]ChangelogChange),
		PackageName: packageName,
	}

	// Process consignments for this package
	for _, c := range consignments {
		if changeType, exists := c.Packages[packageName]; exists {
			// Map change type to changelog section
			section := g.mapChangeTypeToSection(changeType)

			change := ChangelogChange{
				Summary:     c.Summary,
				PackageName: packageName,
				Consignment: &consignment.Consignment{
					ID:       c.ID,
					Packages: c.Packages,
					Summary:  c.Summary,
					Created:  c.Created,
				},
			}

			entry.Changes[section] = append(entry.Changes[section], change)
		}
	}

	return entry
}
