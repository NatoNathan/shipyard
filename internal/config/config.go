package config

import (
	"fmt"
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
	Enabled bool   `yaml:"enabled,omitempty"`
	Owner   string `yaml:"owner,omitempty"`
	Repo    string `yaml:"repo,omitempty"`
	Token   string `yaml:"token,omitempty"` // Format: "env:VAR_NAME"
}

// Package represents a versionable package
type Package struct {
	Name         string          `yaml:"name"`
	Path         string          `yaml:"path"`
	Ecosystem    string          `yaml:"ecosystem,omitempty"`
	VersionFiles []string        `yaml:"versionFiles,omitempty"` // Use ["tag-only"] for tag-only mode
	Dependencies []Dependency    `yaml:"dependencies,omitempty"`
	Templates    *TemplateConfig `yaml:"templates,omitempty"`
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
	
	return merged
}

// WithDefaults returns a config with default values applied
func (c *Config) WithDefaults() *Config {
	result := *c
	
	if result.Consignments.Path == "" {
		result.Consignments.Path = ".shipyard/consignments"
	}
	
	if result.History.Path == "" {
		result.History.Path = ".shipyard/history.json"
	}
	
	// Apply default strategy to dependencies
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
