package changelog

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/NatoNathan/shipyard/internal/consignment"
	"github.com/NatoNathan/shipyard/internal/template"
	"github.com/NatoNathan/shipyard/pkg/semver"
)

// ParseTagOutput parses tag template output into name and optional message
// Single line: lightweight tag (name only)
// Multi-line: annotated tag (name\n\nmessage)
func ParseTagOutput(output string) (name, message string, err error) {
	// Trim leading/trailing whitespace
	output = strings.TrimSpace(output)

	if output == "" {
		return "", "", fmt.Errorf("tag output is empty")
	}

	// Split into lines
	lines := strings.Split(output, "\n")

	// Single line = lightweight tag (name only)
	if len(lines) == 1 {
		return lines[0], "", nil
	}

	// Multi-line = annotated tag
	// Format: name\n\nmessage
	// Second line must be blank
	if len(lines) >= 2 && lines[1] != "" {
		return "", "", fmt.Errorf("multi-line tag output requires blank line after tag name (line 2)")
	}

	// First line is tag name
	name = strings.TrimSpace(lines[0])

	// Rest (after blank line) is message
	if len(lines) > 2 {
		message = strings.Join(lines[2:], "\n")
	}

	return name, message, nil
}

// ChangelogGenerator handles changelog generation from consignments
type ChangelogGenerator struct {
	loader            *template.TemplateLoader
	renderer          *template.TemplateRenderer
	preserveExisting  bool
}

// PackageTag represents a generated tag with name and optional message
type PackageTag struct {
	Name    string // Tag name (e.g., "core/v1.0.0")
	Message string // Tag annotation message (empty for lightweight tags)
}

// NewChangelogGenerator creates a new changelog generator
func NewChangelogGenerator() *ChangelogGenerator {
	return &ChangelogGenerator{
		loader:           template.NewTemplateLoader(),
		renderer:         template.NewTemplateRenderer(),
		preserveExisting: false,
	}
}

// SetBaseDir sets the base directory for resolving template paths
func (g *ChangelogGenerator) SetBaseDir(dir string) {
	g.loader.SetBaseDir(dir)
}

// SetPreserveExisting sets whether to preserve existing changelog content
func (g *ChangelogGenerator) SetPreserveExisting(preserve bool) {
	g.preserveExisting = preserve
}

// GenerateForPackage generates a changelog for a single package
func (g *ChangelogGenerator) GenerateForPackage(
	consignments []*consignment.Consignment,
	packageName string,
	version semver.Version,
	templateSource string,
) (string, error) {
	// Load template (specify we need a changelog template)
	templateContent, err := g.loader.Load(templateSource, template.TemplateTypeChangelog)
	if err != nil {
		return "", fmt.Errorf("failed to load template: %w", err)
	}

	return g.GenerateForPackageWithTemplate(consignments, packageName, version, templateContent)
}

// GenerateForPackageWithTemplate generates a changelog for a single package using an inline template
func (g *ChangelogGenerator) GenerateForPackageWithTemplate(
	consignments []*consignment.Consignment,
	packageName string,
	version semver.Version,
	inlineTemplate string,
) (string, error) {
	// Filter consignments for this package
	filtered := filterConsignmentsForPackage(consignments, packageName)

	// Build single-package context
	context := g.buildSinglePackageContext(packageName, version, filtered)

	// Render template
	result, err := g.renderer.Render(inlineTemplate, context)
	if err != nil {
		return "", fmt.Errorf("failed to render changelog: %w", err)
	}

	return result, nil
}

// GenerateAll generates changelogs for all packages
func (g *ChangelogGenerator) GenerateAll(
	consignments []*consignment.Consignment,
	versions map[string]semver.Version,
	templateSource string,
) (map[string]string, error) {
	results := make(map[string]string)

	for pkg, ver := range versions {
		changelog, err := g.GenerateForPackage(consignments, pkg, ver, templateSource)
		if err != nil {
			return nil, fmt.Errorf("failed to generate changelog for %s: %w", pkg, err)
		}
		results[pkg] = changelog
	}

	return results, nil
}

// WriteChangelogToFile generates a changelog for a single package and writes it to a file
func (g *ChangelogGenerator) WriteChangelogToFile(
	consignments []*consignment.Consignment,
	packageName string,
	version semver.Version,
	templateSource string,
	outputPath string,
) error {
	// Generate new changelog content
	newContent, err := g.GenerateForPackage(consignments, packageName, version, templateSource)
	if err != nil {
		return err
	}

	// If preserving existing content, prepend new content
	if g.preserveExisting {
		existingContent, err := os.ReadFile(outputPath)
		if err == nil {
			// Prepend new content before existing
			newContent = newContent + "\n" + string(existingContent)
		}
		// If file doesn't exist, just use new content
	}

	// Write to file
	if err := os.WriteFile(outputPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write changelog: %w", err)
	}

	return nil
}

// GeneratePackageTag generates a git tag name and optional message for a single package
// Returns: (tagName, message, error)
// - Single-line output: lightweight tag (name only, empty message)
// - Multi-line output: annotated tag (name + message)
func (g *ChangelogGenerator) GeneratePackageTag(
	consignments []*consignment.Consignment,
	packageName string,
	version semver.Version,
	templateSource string,
) (string, string, error) {
	// Load template (specify we need a tag template)
	templateContent, err := g.loader.Load(templateSource, template.TemplateTypeTag)
	if err != nil {
		return "", "", fmt.Errorf("failed to load template: %w", err)
	}

	return g.GeneratePackageTagWithContext(consignments, packageName, version, templateContent)
}

// GeneratePackageTagWithContext generates a package tag using an inline template with full context
// Filters consignments to only those affecting the package, and includes them in template context
// Returns: (tagName, message, error)
func (g *ChangelogGenerator) GeneratePackageTagWithContext(
	consignments []*consignment.Consignment,
	packageName string,
	version semver.Version,
	inlineTemplate string,
) (string, string, error) {
	// Filter consignments for this package
	filtered := filterConsignmentsForPackage(consignments, packageName)

	// Build single-package context (same as changelog generation)
	context := g.buildSinglePackageContext(packageName, version, filtered)

	// Render template
	result, err := g.renderer.Render(inlineTemplate, context)
	if err != nil {
		return "", "", fmt.Errorf("failed to render package tag: %w", err)
	}

	// Parse output to extract tag name and optional message
	tagName, message, err := ParseTagOutput(result)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse tag output: %w", err)
	}

	return tagName, message, nil
}

// GenerateReleaseTag generates a git tag name and optional message for the entire release
// Returns: (tagName, message, error)
// - Single-line output: lightweight tag (name only, empty message)
// - Multi-line output: annotated tag (name + message)
func (g *ChangelogGenerator) GenerateReleaseTag(
	consignments []*consignment.Consignment,
	packages []string,
	versions map[string]semver.Version,
	templateSource string,
) (string, string, error) {
	// Load template (specify we need a release template)
	templateContent, err := g.loader.Load(templateSource, template.TemplateTypeRelease)
	if err != nil {
		return "", "", fmt.Errorf("failed to load template: %w", err)
	}

	return g.GenerateReleaseTagWithContext(consignments, packages, versions, templateContent)
}

// GenerateReleaseTagWithContext generates a release tag using an inline template with full context
// Includes all consignments, all packages, and all versions in template context
// Returns: (tagName, message, error)
func (g *ChangelogGenerator) GenerateReleaseTagWithContext(
	consignments []*consignment.Consignment,
	packages []string,
	versions map[string]semver.Version,
	inlineTemplate string,
) (string, string, error) {
	// Build multi-package context for release tag
	type Package struct {
		Name string
	}

	packageStructs := make([]Package, len(packages))
	for i, pkg := range packages {
		packageStructs[i] = Package{Name: pkg}
	}

	// Convert semver.Version map to string map
	versionStrings := make(map[string]string)
	for pkg, ver := range versions {
		versionStrings[pkg] = ver.String()
	}

	// Convert consignments to template-friendly format
	type TemplateConsignment struct {
		ID         string
		Timestamp  time.Time
		Packages   []string
		ChangeType string
		Summary    string
		Metadata   map[string]interface{}
	}

	templateConsignments := make([]TemplateConsignment, len(consignments))
	for i, c := range consignments {
		templateConsignments[i] = TemplateConsignment{
			ID:         c.ID,
			Timestamp:  c.Timestamp,
			Packages:   c.Packages,
			ChangeType: string(c.ChangeType),
			Summary:    c.Summary,
			Metadata:   c.Metadata,
		}
	}

	context := map[string]interface{}{
		"Packages":     packageStructs,
		"Versions":     versionStrings,
		"Consignments": templateConsignments,
		"Date":         time.Now(),
		"Metadata":     aggregateMetadata(consignments),
	}

	result, err := g.renderer.Render(inlineTemplate, context)
	if err != nil {
		return "", "", fmt.Errorf("failed to render release tag: %w", err)
	}

	// Parse output to extract tag name and optional message
	tagName, message, err := ParseTagOutput(result)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse tag output: %w", err)
	}

	return tagName, message, nil
}

// GenerateAllPackageTags generates tags for all packages
// Returns map of package name to PackageTag (with name and optional message)
func (g *ChangelogGenerator) GenerateAllPackageTags(
	consignments []*consignment.Consignment,
	versions map[string]semver.Version,
	templateSource string,
) (map[string]PackageTag, error) {
	tags := make(map[string]PackageTag)

	for pkg, version := range versions {
		tagName, message, err := g.GeneratePackageTag(consignments, pkg, version, templateSource)
		if err != nil {
			return nil, fmt.Errorf("failed to generate tag for package %s: %w", pkg, err)
		}
		tags[pkg] = PackageTag{
			Name:    tagName,
			Message: message,
		}
	}

	return tags, nil
}

// buildSinglePackageContext builds the template context for a single package
func (g *ChangelogGenerator) buildSinglePackageContext(
	packageName string,
	version semver.Version,
	consignments []*consignment.Consignment,
) map[string]interface{} {
	// Convert consignments to template-friendly format
	type TemplateConsignment struct {
		ID         string
		Timestamp  time.Time
		Packages   []string
		ChangeType string
		Summary    string
		Metadata   map[string]interface{}
	}

	templateConsignments := make([]TemplateConsignment, len(consignments))
	for i, c := range consignments {
		templateConsignments[i] = TemplateConsignment{
			ID:         c.ID,
			Timestamp:  c.Timestamp,
			Packages:   c.Packages,
			ChangeType: string(c.ChangeType),
			Summary:    c.Summary,
			Metadata:   c.Metadata,
		}
	}

	// Build context
	context := map[string]interface{}{
		"Package":      packageName,
		"Version":      version.String(),
		"Consignments": templateConsignments,
		"Date":         time.Now(),
		"Metadata":     aggregateMetadata(consignments),
	}

	return context
}

// filterConsignmentsForPackage filters consignments to only those affecting the specified package
func filterConsignmentsForPackage(consignments []*consignment.Consignment, packageName string) []*consignment.Consignment {
	var filtered []*consignment.Consignment
	for _, c := range consignments {
		for _, pkg := range c.Packages {
			if pkg == packageName {
				filtered = append(filtered, c)
				break
			}
		}
	}
	return filtered
}

// aggregateMetadata collects all metadata from consignments
func aggregateMetadata(consignments []*consignment.Consignment) map[string]interface{} {
	metadata := make(map[string]interface{})

	for _, c := range consignments {
		if c.Metadata == nil {
			continue
		}

		for key, value := range c.Metadata {
			// For now, last value wins (could be enhanced to collect arrays)
			metadata[key] = value
		}
	}

	return metadata
}

// GenerateReleaseNotes generates release notes for a single package
func (g *ChangelogGenerator) GenerateReleaseNotes(
	consignments []*consignment.Consignment,
	packageName string,
	version semver.Version,
	templateSource string,
) (string, error) {
	// Load template (specify we need a releasenotes template)
	templateContent, err := g.loader.Load(templateSource, template.TemplateTypeReleaseNotes)
	if err != nil {
		return "", fmt.Errorf("failed to load template: %w", err)
	}

	return g.GenerateForPackageWithTemplate(consignments, packageName, version, templateContent)
}

// GetDefaultChangelogTemplate returns the default changelog template source
func GetDefaultChangelogTemplate() string {
	return "builtin:default"
}

// GetDefaultPackageTagTemplate returns the default package tag template source
func GetDefaultPackageTagTemplate() string {
	return "builtin:default"
}

// GetDefaultReleaseTagTemplate returns the default release tag template source
func GetDefaultReleaseTagTemplate() string {
	return "builtin:release-date"
}

// GetDefaultReleaseNotesTemplate returns the default release notes template source
func GetDefaultReleaseNotesTemplate() string {
	return "builtin:default"
}

// GetDefaultCommitTemplate returns the default commit message template source
func GetDefaultCommitTemplate() string {
	return "builtin:default"
}

// VersionBump represents a version bump for a package (to avoid circular import)
type VersionBump struct {
	Package    string
	OldVersion semver.Version
	NewVersion semver.Version
	ChangeType string
}

// GenerateCommitMessage generates a git commit message from version bumps
// This is used when creating commits for version changes
func (g *ChangelogGenerator) GenerateCommitMessage(
	consignments []*consignment.Consignment,
	versionBumps map[string]VersionBump,
	templateSource string,
) (string, error) {
	// Load template (specify we need a commit template)
	templateContent, err := g.loader.Load(templateSource, template.TemplateTypeCommit)
	if err != nil {
		return "", fmt.Errorf("failed to load template: %w", err)
	}

	// Convert version bumps to package info for template
	type PackageInfo struct {
		Name       string
		OldVersion string
		NewVersion string
		ChangeType string
	}

	packages := make([]PackageInfo, 0, len(versionBumps))
	for name, bump := range versionBumps {
		packages = append(packages, PackageInfo{
			Name:       name,
			OldVersion: bump.OldVersion.String(),
			NewVersion: bump.NewVersion.String(),
			ChangeType: bump.ChangeType,
		})
	}

	// Convert consignments to template-friendly format
	type TemplateConsignment struct {
		ID         string
		Timestamp  time.Time
		Packages   []string
		ChangeType string
		Summary    string
		Metadata   map[string]interface{}
	}

	templateConsignments := make([]TemplateConsignment, len(consignments))
	for i, c := range consignments {
		templateConsignments[i] = TemplateConsignment{
			ID:         c.ID,
			Timestamp:  c.Timestamp,
			Packages:   c.Packages,
			ChangeType: string(c.ChangeType),
			Summary:    c.Summary,
			Metadata:   c.Metadata,
		}
	}

	// Build context
	context := map[string]interface{}{
		"Packages":     packages,
		"Consignments": templateConsignments,
		"Date":         time.Now(),
		"Metadata":     aggregateMetadata(consignments),
	}

	// Render template
	result, err := g.renderer.Render(templateContent, context)
	if err != nil {
		return "", fmt.Errorf("failed to render commit message: %w", err)
	}

	return result, nil
}
