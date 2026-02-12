package config

import (
	"fmt"
	"sort"
	"strings"
)

// Ecosystem types
const (
	EcosystemGo     = "go"
	EcosystemNPM    = "npm"
	EcosystemPython = "python"
	EcosystemHelm   = "helm"
	EcosystemCargo  = "cargo"
	EcosystemDeno   = "deno"
)

// Config represents the project-specific settings
type Config struct {
	Extends      []RemoteConfig    `yaml:"extends,omitempty"`
	Packages     []Package         `yaml:"packages"`
	Templates    TemplateConfig    `yaml:"templates,omitempty"`
	Metadata     MetadataConfig    `yaml:"metadata,omitempty"`
	Consignments ConsignmentConfig `yaml:"consignments,omitempty"`
	History      HistoryConfig     `yaml:"history,omitempty"`
	GitHub       GitHubConfig      `yaml:"github,omitempty"`
	PreRelease   PreReleaseConfig  `yaml:"prerelease,omitempty"`
}

// PreReleaseConfig holds pre-release stage definitions and snapshot template
type PreReleaseConfig struct {
	Stages              []StageConfig `yaml:"stages,omitempty"`
	SnapshotTagTemplate string        `yaml:"snapshotTagTemplate,omitempty"`
}

// StageConfig defines a pre-release stage
type StageConfig struct {
	Name        string `yaml:"name"`
	Order       int    `yaml:"order"`
	TagTemplate string `yaml:"tagTemplate,omitempty"`
}

// GetLowestOrderStage returns the stage with the lowest order value
func (c *PreReleaseConfig) GetLowestOrderStage() (StageConfig, bool) {
	if len(c.Stages) == 0 {
		return StageConfig{}, false
	}
	sorted := c.sortedStages()
	return sorted[0], true
}

// GetStageByName returns the stage with the given name
func (c *PreReleaseConfig) GetStageByName(name string) (StageConfig, bool) {
	for _, s := range c.Stages {
		if s.Name == name {
			return s, true
		}
	}
	return StageConfig{}, false
}

// GetNextStage returns the stage with the next-highest order after the given stage name
func (c *PreReleaseConfig) GetNextStage(currentName string) (StageConfig, bool) {
	current, ok := c.GetStageByName(currentName)
	if !ok {
		return StageConfig{}, false
	}
	sorted := c.sortedStages()
	for _, s := range sorted {
		if s.Order > current.Order {
			return s, true
		}
	}
	return StageConfig{}, false
}

// IsHighestStage returns true if the given stage name is the highest-order stage
func (c *PreReleaseConfig) IsHighestStage(name string) bool {
	current, ok := c.GetStageByName(name)
	if !ok {
		return false
	}
	sorted := c.sortedStages()
	return sorted[len(sorted)-1].Order == current.Order
}

// Validate checks if the pre-release configuration is valid (only when stages are defined)
func (c *PreReleaseConfig) Validate() error {
	if len(c.Stages) == 0 {
		return nil // No stages defined is valid (pre-release not configured)
	}
	names := make(map[string]bool)
	orders := make(map[int]bool)
	for _, s := range c.Stages {
		if s.Name == "" {
			return fmt.Errorf("stage name is required")
		}
		if names[s.Name] {
			return fmt.Errorf("duplicate stage name: %s", s.Name)
		}
		names[s.Name] = true
		if orders[s.Order] {
			return fmt.Errorf("duplicate stage order: %d", s.Order)
		}
		orders[s.Order] = true
	}
	return nil
}

// sortedStages returns stages sorted by order (ascending)
func (c *PreReleaseConfig) sortedStages() []StageConfig {
	sorted := make([]StageConfig, len(c.Stages))
	copy(sorted, c.Stages)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Order < sorted[j].Order
	})
	return sorted
}

// RemoteConfig represents a remote configuration source
type RemoteConfig struct {
	URL  string `yaml:"url,omitempty"`
	Git  string `yaml:"git,omitempty"`
	Path string `yaml:"path,omitempty"`
	Ref  string `yaml:"ref,omitempty"`
	Auth string `yaml:"auth,omitempty"`
}

// TemplateConfig holds template definitions
type TemplateConfig struct {
	Changelog     *TemplateSource `yaml:"changelog,omitempty"`
	TagName       *TemplateSource `yaml:"tagName,omitempty"`
	ReleaseNotes  *TemplateSource `yaml:"releaseNotes,omitempty"`
	CommitMessage *TemplateSource `yaml:"commitMessage,omitempty"`
}

// TemplateSource represents a template source
type TemplateSource struct {
	Source string `yaml:"source,omitempty"`
	Inline string `yaml:"inline,omitempty"`
}

// MetadataConfig defines custom metadata fields
type MetadataConfig struct {
	Fields []MetadataField `yaml:"fields,omitempty"`
}

// MetadataField defines a custom metadata field
type MetadataField struct {
	Name     string `yaml:"name"`
	Required bool   `yaml:"required,omitempty"`
	Type     string `yaml:"type,omitempty"` // "string", "int", "list", "map"

	// String validation
	Pattern   string `yaml:"pattern,omitempty"`
	MinLength *int   `yaml:"minLength,omitempty"`
	MaxLength *int   `yaml:"maxLength,omitempty"`

	// Integer validation
	Min *int `yaml:"min,omitempty"`
	Max *int `yaml:"max,omitempty"`

	// List validation
	ItemType string `yaml:"itemType,omitempty"` // "string", "int"
	MinItems *int   `yaml:"minItems,omitempty"`
	MaxItems *int   `yaml:"maxItems,omitempty"`

	// Common
	AllowedValues []string `yaml:"allowedValues,omitempty"`
	Default       string   `yaml:"default,omitempty"`
	Description   string   `yaml:"description,omitempty"`
}

// ConsignmentConfig holds consignment storage settings
type ConsignmentConfig struct {
	Path string `yaml:"path,omitempty"`
}

// HistoryConfig holds history file settings
type HistoryConfig struct {
	Path string `yaml:"path,omitempty"`
}

// GitHubConfig holds GitHub integration settings
type GitHubConfig struct {
	Owner string `yaml:"owner,omitempty"`
	Repo  string `yaml:"repo,omitempty"`
	Token string `yaml:"token,omitempty"` // Format: "env:VAR_NAME"
}

// Package represents a versionable package
type Package struct {
	Name         string                 `yaml:"name"`
	Path         string                 `yaml:"path"`
	Ecosystem    string                 `yaml:"ecosystem,omitempty"`
	VersionFiles []string               `yaml:"versionFiles,omitempty"` // Use ["tag-only"] for tag-only mode
	Dependencies []Dependency           `yaml:"dependencies,omitempty"`
	Templates    *TemplateConfig        `yaml:"templates,omitempty"`
	Options      map[string]interface{} `yaml:"options,omitempty"`
}

// IsTagOnly returns true if this package uses tag-only versioning (no file updates)
// This is indicated by versionFiles containing "tag-only" keyword
func (p *Package) IsTagOnly() bool {
	if len(p.VersionFiles) == 0 {
		return false
	}
	for _, vf := range p.VersionFiles {
		if vf == "tag-only" {
			return true
		}
	}
	return false
}

// HelmOptions contains Helm-specific package options
type HelmOptions struct {
	AppDependency string // Package name to use for appVersion
}

// GetHelmOptions extracts Helm-specific options from package options
func (p *Package) GetHelmOptions() *HelmOptions {
	if p.Options == nil {
		return &HelmOptions{}
	}

	opts := &HelmOptions{}
	if appDep, ok := p.Options["appDependency"].(string); ok {
		opts.AppDependency = appDep
	}
	return opts
}

// Dependency represents a package dependency
type Dependency struct {
	Package     string            `yaml:"package"`
	Strategy    string            `yaml:"strategy,omitempty"` // "fixed" or "linked"
	BumpMapping map[string]string `yaml:"bumpMapping,omitempty"`
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if len(c.Packages) == 0 {
		return fmt.Errorf("at least one package must be defined")
	}

	// Check for duplicate package names
	names := make(map[string]bool)
	for _, pkg := range c.Packages {
		if names[pkg.Name] {
			return fmt.Errorf("duplicate package name: %s", pkg.Name)
		}
		names[pkg.Name] = true

		// Validate each package
		if err := pkg.Validate(); err != nil {
			return fmt.Errorf("invalid package %s: %w", pkg.Name, err)
		}
	}

	// Validate package options (requires all packages to be known)
	for _, pkg := range c.Packages {
		if err := pkg.ValidateOptions(c.Packages); err != nil {
			return fmt.Errorf("invalid options for package %s: %w", pkg.Name, err)
		}
	}

	return nil
}

// Validate checks if a package is valid
func (p *Package) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("package name is required")
	}
	if p.Path == "" {
		return fmt.Errorf("package path is required")
	}
	return nil
}

// ValidateOptions validates package options against the configuration
// This is called by Config.Validate after all packages are known
func (p *Package) ValidateOptions(allPackages []Package) error {
	if p.Ecosystem == EcosystemHelm {
		helmOpts := p.GetHelmOptions()
		if helmOpts.AppDependency != "" {
			// Check that appDependency references a valid package
			found := false
			for _, pkg := range allPackages {
				if pkg.Name == helmOpts.AppDependency {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("helm package %q has appDependency %q but no such package exists",
					p.Name, helmOpts.AppDependency)
			}
		}
	}
	return nil
}

// GetPackage retrieves a package by name
func (c *Config) GetPackage(name string) (Package, bool) {
	for _, pkg := range c.Packages {
		if pkg.Name == name {
			return pkg, true
		}
	}
	return Package{}, false
}

// Merge merges this config with another, with the overlay taking precedence
func (c *Config) Merge(overlay *Config) *Config {
	merged := &Config{
		Packages:     append([]Package{}, c.Packages...),
		Extends:      append([]RemoteConfig{}, c.Extends...),
		Templates:    c.Templates,
		Metadata:     c.Metadata,
		Consignments: c.Consignments,
		History:      c.History,
		GitHub:       c.GitHub,
		PreRelease:   c.PreRelease,
	}
	
	// Append overlay packages
	merged.Packages = append(merged.Packages, overlay.Packages...)
	
	// Overlay takes precedence for other fields
	if len(overlay.Extends) > 0 {
		merged.Extends = overlay.Extends
	}
	if overlay.Templates.Changelog != nil || overlay.Templates.TagName != nil || overlay.Templates.ReleaseNotes != nil || overlay.Templates.CommitMessage != nil {
		merged.Templates = overlay.Templates
	}
	if len(overlay.Metadata.Fields) > 0 {
		merged.Metadata = overlay.Metadata
	}
	if overlay.Consignments.Path != "" {
		merged.Consignments = overlay.Consignments
	}
	if overlay.History.Path != "" {
		merged.History = overlay.History
	}
	if overlay.GitHub.Owner != "" || overlay.GitHub.Repo != "" {
		merged.GitHub = overlay.GitHub
	}
	if len(overlay.PreRelease.Stages) > 0 || overlay.PreRelease.SnapshotTagTemplate != "" {
		merged.PreRelease = overlay.PreRelease
	}

	return merged
}

// WithDefaults returns a config with default values applied.
// Performs a deep copy so the original config is not modified.
func (c *Config) WithDefaults() *Config {
	result := Config{
		Templates:    c.Templates,
		Consignments: c.Consignments,
		History:      c.History,
		GitHub:       c.GitHub,
	}

	// Deep copy Extends
	if len(c.Extends) > 0 {
		result.Extends = make([]RemoteConfig, len(c.Extends))
		copy(result.Extends, c.Extends)
	}

	// Deep copy Packages (and nested Dependencies/BumpMapping)
	result.Packages = make([]Package, len(c.Packages))
	for i, pkg := range c.Packages {
		result.Packages[i] = pkg
		if len(pkg.Dependencies) > 0 {
			result.Packages[i].Dependencies = make([]Dependency, len(pkg.Dependencies))
			for j, dep := range pkg.Dependencies {
				result.Packages[i].Dependencies[j] = dep
				if dep.BumpMapping != nil {
					result.Packages[i].Dependencies[j].BumpMapping = make(map[string]string)
					for k, v := range dep.BumpMapping {
						result.Packages[i].Dependencies[j].BumpMapping[k] = v
					}
				}
			}
		}
	}

	// Deep copy Metadata.Fields
	if len(c.Metadata.Fields) > 0 {
		result.Metadata.Fields = make([]MetadataField, len(c.Metadata.Fields))
		copy(result.Metadata.Fields, c.Metadata.Fields)
	}

	// Deep copy PreRelease.Stages
	result.PreRelease = c.PreRelease
	if len(c.PreRelease.Stages) > 0 {
		result.PreRelease.Stages = make([]StageConfig, len(c.PreRelease.Stages))
		copy(result.PreRelease.Stages, c.PreRelease.Stages)
	}

	// Apply defaults
	if result.Consignments.Path == "" {
		result.Consignments.Path = ".shipyard/consignments"
	}
	if result.History.Path == "" {
		result.History.Path = ".shipyard/history.json"
	}
	for i := range result.Packages {
		for j := range result.Packages[i].Dependencies {
			if result.Packages[i].Dependencies[j].Strategy == "" {
				result.Packages[i].Dependencies[j].Strategy = "linked"
			}
		}
	}

	return &result
}

// UnmarshalYAML implements custom unmarshaling for RemoteConfig
func (rc *RemoteConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try string format first (implied type detection)
	var str string
	if err := unmarshal(&str); err == nil {
		rc.parseImplied(str)
		return nil
	}
	
	// Fall back to object format
	type rawConfig RemoteConfig
	var raw rawConfig
	if err := unmarshal(&raw); err != nil {
		return err
	}
	*rc = RemoteConfig(raw)
	return nil
}

// parseImplied detects type from string format
func (rc *RemoteConfig) parseImplied(value string) {
	// Git URL patterns
	if strings.HasPrefix(value, "git@") ||
		strings.HasPrefix(value, "ssh://") ||
		strings.Contains(value, ".git") {
		rc.parseGitURL(value)
		return
	}
	
	// HTTP(S) URL
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		if strings.Contains(value, ".git") || strings.Contains(value, "#") {
			rc.parseGitURL(value)
		} else {
			rc.URL = value
		}
		return
	}
	
	// Default: treat as direct URL
	rc.URL = value
}

// parseGitURL extracts git URL, path, and ref from fragment syntax
func (rc *RemoteConfig) parseGitURL(value string) {
	parts := strings.SplitN(value, "#", 2)
	rc.Git = parts[0]
	
	if len(parts) == 2 {
		fragment := parts[1]
		if idx := strings.LastIndex(fragment, "@"); idx != -1 {
			rc.Path = fragment[:idx]
			rc.Ref = fragment[idx+1:]
		} else {
			rc.Path = fragment
		}
	}
	
	if rc.Path == "" {
		rc.Path = "shipyard.yaml"
	}
	if rc.Ref == "" {
		rc.Ref = "main"
	}
}
