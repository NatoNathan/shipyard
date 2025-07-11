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
	projectConfig *config.ProjectConfig
	template      Template
}

// NewGenerator creates a new changelog generator
func NewGenerator(projectConfig *config.ProjectConfig) (*Generator, error) {
	template, err := GetTemplate(projectConfig.Changelog.Template)
	if err != nil {
		return nil, fmt.Errorf("failed to get changelog template: %w", err)
	}

	return &Generator{
		projectConfig: projectConfig,
		template:      template,
	}, nil
}

// GenerateChangelog generates a changelog from consignments
func (g *Generator) GenerateChangelog(consignments []*consignment.Consignment, versions map[string]*semver.Version) (string, error) {
	if len(consignments) == 0 {
		return "", fmt.Errorf("no consignments provided")
	}

	// Group consignments by package and version
	entries := g.groupConsignmentsByPackage(consignments, versions)

	// Sort entries by version (newest first)
	sort.Slice(entries, func(i, j int) bool {
		versionI, _ := semver.Parse(entries[i].Version)
		versionJ, _ := semver.Parse(entries[j].Version)
		if versionI != nil && versionJ != nil {
			return versionI.GreaterThan(versionJ)
		}
		return entries[i].Date.After(entries[j].Date)
	})

	// Generate changelog using the template
	return g.template.Generate(entries, g.projectConfig)
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
			section := g.mapChangeTypeToSection(consignment.ChangeType(changeType))

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
func (g *Generator) mapChangeTypeToSection(changeType consignment.ChangeType) string {
	switch changeType {
	case consignment.Major:
		return "Breaking Changes"
	case consignment.Minor:
		return "Added"
	case consignment.Patch:
		return "Fixed"
	default:
		return "Changed"
	}
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
