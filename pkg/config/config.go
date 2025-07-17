// Package config provides public configuration types and utilities for Shipyard projects.
// This package contains the core configuration structures that can be used by external
// tools, MCP servers, and other integrations.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// RepoType represents the type of repository structure
type RepoType string

// Repository types supported by Shipyard
const (
	// RepositoryTypeMonorepo represents a monorepo with multiple packages
	RepositoryTypeMonorepo RepoType = "monorepo"
	// RepositoryTypeSingleRepo represents a single-package repository
	RepositoryTypeSingleRepo RepoType = "single-repo"
)

// ProjectConfig represents the complete configuration for a Shipyard project
type ProjectConfig struct {
	Type RepoType `mapstructure:"type" json:"type" yaml:"type"` // "monorepo" or "single-repo"
	Repo string   `mapstructure:"repo" json:"repo" yaml:"repo"` // e.g., "github.com/NatoNathan/shipyard"

	Changelog   ChangelogConfig    `mapstructure:"changelog" json:"changelog" yaml:"changelog"`
	Git         GitConfig          `mapstructure:"git" json:"git,omitempty" yaml:"git,omitempty"`
	ChangeTypes []ChangeTypeConfig `mapstructure:"change_types" json:"change_types,omitempty" yaml:"change_types,omitempty"`

	// For monorepo projects
	Packages []Package `mapstructure:"packages" json:"packages,omitempty" yaml:"packages,omitempty"`
	// For single-repo projects
	Package Package `mapstructure:"package" json:"package,omitempty" yaml:"package,omitempty"`
}

// ChangelogConfig represents the changelog configuration
type ChangelogConfig struct {
	Template    string `mapstructure:"template" json:"template" yaml:"template"`                                 // template name, file path, or URL
	OutputPath  string `mapstructure:"output_path" json:"output_path,omitempty" yaml:"output_path,omitempty"`    // default changelog filename (default: "CHANGELOG.md")
	PackagePath *bool  `mapstructure:"package_path" json:"package_path,omitempty" yaml:"package_path,omitempty"` // place changelog in package path for monorepo (default: true for monorepo)
}

// GitConfig represents the git integration configuration
type GitConfig struct {
	TagTemplate    string `mapstructure:"tag_template" json:"tag_template,omitempty" yaml:"tag_template,omitempty"`
	CommitTemplate string `mapstructure:"commit_template" json:"commit_template,omitempty" yaml:"commit_template,omitempty"`
}

// ChangeTypeConfig represents a custom change type configuration
type ChangeTypeConfig struct {
	Name        string `mapstructure:"name" json:"name" yaml:"name"`                              // e.g., "feat", "fix", "docs"
	DisplayName string `mapstructure:"display_name" json:"display_name" yaml:"display_name"`      // e.g., "✨ Feature", "🔧 Bug Fix"
	SemverBump  string `mapstructure:"semver_bump" json:"semver_bump" yaml:"semver_bump"`         // "major", "minor", "patch"
	Section     string `mapstructure:"section" json:"section,omitempty" yaml:"section,omitempty"` // changelog section name
}

// GetPackages returns all packages in the project configuration.
// For monorepo projects, it returns the Packages slice.
// For single-repo projects, it returns a slice containing the single Package.
func (c *ProjectConfig) GetPackages() []Package {
	if c.Type == RepositoryTypeMonorepo {
		return c.Packages
	}
	return []Package{c.Package}
}

// GetPackageByName returns the package with the specified name, or nil if not found
func (c *ProjectConfig) GetPackageByName(name string) *Package {
	packages := c.GetPackages()
	for i := range packages {
		if packages[i].Name == name {
			return &packages[i]
		}
	}
	return nil
}

// GetChangelogOutputPath returns the changelog output path configuration
func (c *ProjectConfig) GetChangelogOutputPath() string {
	if c.Changelog.OutputPath != "" {
		return c.Changelog.OutputPath
	}
	return "CHANGELOG.md" // default
}

// ShouldUsePackagePaths returns true if changelogs should be placed in package directories for monorepo
func (c *ProjectConfig) ShouldUsePackagePaths() bool {
	// Default to true for monorepo projects unless explicitly disabled
	if c.Type == RepositoryTypeMonorepo {
		// If not explicitly set, default to true
		if c.Changelog.PackagePath == nil {
			return true
		}
		return *c.Changelog.PackagePath
	}
	return false
}

// HasPackage returns true if the project has a package with the specified name
func (c *ProjectConfig) HasPackage(name string) bool {
	return c.GetPackageByName(name) != nil
}

// GetPackageNames returns the names of all packages in the project
func (c *ProjectConfig) GetPackageNames() []string {
	packages := c.GetPackages()
	names := make([]string, len(packages))
	for i, pkg := range packages {
		names[i] = pkg.Name
	}
	return names
}

// GetChangeTypes returns the custom change types or default ones if none are configured
func (c *ProjectConfig) GetChangeTypes() []ChangeTypeConfig {
	if len(c.ChangeTypes) > 0 {
		return c.ChangeTypes
	}
	return DefaultChangeTypes()
}

// GetChangeTypeByName returns the change type configuration by name
func (c *ProjectConfig) GetChangeTypeByName(name string) *ChangeTypeConfig {
	changeTypes := c.GetChangeTypes()
	for i := range changeTypes {
		if changeTypes[i].Name == name {
			return &changeTypes[i]
		}
	}
	return nil
}

// GetChangeTypeNames returns the names of all available change types
func (c *ProjectConfig) GetChangeTypeNames() []string {
	changeTypes := c.GetChangeTypes()
	names := make([]string, len(changeTypes))
	for i, ct := range changeTypes {
		names[i] = ct.Name
	}
	return names
}

// DefaultChangeTypes returns the default change type configurations
func DefaultChangeTypes() []ChangeTypeConfig {
	return []ChangeTypeConfig{
		{
			Name:        "patch",
			DisplayName: "🔧 Patch - Bug fixes and minor updates",
			SemverBump:  "patch",
			Section:     "Fixed",
		},
		{
			Name:        "minor",
			DisplayName: "✨ Minor - New features (backward compatible)",
			SemverBump:  "minor",
			Section:     "Added",
		},
		{
			Name:        "major",
			DisplayName: "💥 Major - Breaking changes",
			SemverBump:  "major",
			Section:     "Changed",
		},
	}
}

// IsValid performs basic validation on the project configuration
func (c *ProjectConfig) IsValid() error {
	if c.Type == "" {
		return &ValidationError{Field: "type", Message: "repository type is required"}
	}

	if c.Type != RepositoryTypeMonorepo && c.Type != RepositoryTypeSingleRepo {
		return &ValidationError{Field: "type", Message: "repository type must be 'monorepo' or 'single-repo'"}
	}

	if c.Repo == "" {
		return &ValidationError{Field: "repo", Message: "repository URL is required"}
	}

	if c.Type == RepositoryTypeMonorepo {
		if len(c.Packages) == 0 {
			return &ValidationError{Field: "packages", Message: "monorepo must have at least one package"}
		}
		for i, pkg := range c.Packages {
			if err := pkg.IsValid(); err != nil {
				return &ValidationError{Field: "packages[" + string(rune(i)) + "]", Message: err.Error()}
			}
		}
	} else {
		if err := c.Package.IsValid(); err != nil {
			return &ValidationError{Field: "package", Message: err.Error()}
		}
	}

	return nil
}

// ToMap converts the ProjectConfig to a map[string]interface{} suitable for serialization
func (c *ProjectConfig) ToMap() map[string]interface{} {
	configMap := make(map[string]interface{})

	configMap["type"] = c.Type
	configMap["repo"] = c.Repo

	// Add changelog config
	changelogMap := make(map[string]interface{})
	if c.Changelog.Template != "" {
		changelogMap["template"] = c.Changelog.Template
	}
	if len(changelogMap) > 0 {
		configMap["changelog"] = changelogMap
	}

	// Add git config
	gitMap := make(map[string]interface{})
	if c.Git.TagTemplate != "" {
		gitMap["tag_template"] = c.Git.TagTemplate
	}
	if c.Git.CommitTemplate != "" {
		gitMap["commit_template"] = c.Git.CommitTemplate
	}
	if len(gitMap) > 0 {
		configMap["git"] = gitMap
	}

	// Add change types if custom ones are defined
	if len(c.ChangeTypes) > 0 {
		changeTypesSlice := make([]map[string]interface{}, len(c.ChangeTypes))
		for i, ct := range c.ChangeTypes {
			ctMap := map[string]interface{}{
				"name":         ct.Name,
				"display_name": ct.DisplayName,
				"semver_bump":  ct.SemverBump,
			}
			if ct.Section != "" {
				ctMap["section"] = ct.Section
			}
			changeTypesSlice[i] = ctMap
		}
		configMap["change_types"] = changeTypesSlice
	}

	// Add packages based on type
	if c.Type == RepositoryTypeMonorepo {
		if len(c.Packages) > 0 {
			packagesMap := make([]map[string]interface{}, len(c.Packages))
			for i, pkg := range c.Packages {
				packagesMap[i] = pkg.ToMap()
			}
			configMap["packages"] = packagesMap
		}
	} else {
		// Single repo
		configMap["package"] = c.Package.ToMap()
	}

	return configMap
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// DefaultProjectConfig returns a default project configuration
func DefaultProjectConfig() *ProjectConfig {
	return &ProjectConfig{
		Type: RepositoryTypeMonorepo,
		Repo: "github.com/example/example-repo",
		Changelog: ChangelogConfig{
			Template: "keepachangelog",
		},
		Packages: []Package{
			{
				Name:      "api",
				Path:      "packages/api",
				Ecosystem: "go",
				Manifest:  "packages/api/go.mod",
			},
		},
	}
}

// NewMonorepoConfig creates a new monorepo configuration
func NewMonorepoConfig(repo string, packages []Package) *ProjectConfig {
	return &ProjectConfig{
		Type:      RepositoryTypeMonorepo,
		Repo:      repo,
		Packages:  packages,
		Changelog: ChangelogConfig{Template: "keepachangelog"},
	}
}

// NewSingleRepoConfig creates a new single-repo configuration
func NewSingleRepoConfig(repo string, pkg Package) *ProjectConfig {
	return &ProjectConfig{
		Type:      RepositoryTypeSingleRepo,
		Repo:      repo,
		Package:   pkg,
		Changelog: ChangelogConfig{Template: "keepachangelog"},
	}
}

// LoadFromFile loads a project configuration from a file
func LoadFromFile(configPath string) (*ProjectConfig, error) {
	if configPath == "" {
		return nil, &ValidationError{Field: "configPath", Message: "config path cannot be empty"}
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", configPath)
	}

	// Create a new viper instance for this load operation
	v := viper.New()
	setViperConfigFromPath(v, configPath)

	// Set defaults
	v.SetDefault("type", RepositoryTypeMonorepo)
	v.SetDefault("repo", "github.com/example/example-repo")
	v.SetDefault("changelog.template", "keepachangelog")

	// Read the config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Handle config inheritance (extends)
	if extends := v.GetString("extends"); extends != "" {
		baseConfig, err := loadBaseConfig(extends, configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load base config: %w", err)
		}
		v.MergeConfigMap(baseConfig.AllSettings())
	}

	// Unmarshal into our struct
	var config ProjectConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate the loaded config
	if err := config.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// SaveToFile saves a project configuration to a file
func SaveToFile(config *ProjectConfig, configPath string) error {
	if config == nil {
		return &ValidationError{Field: "config", Message: "config cannot be nil"}
	}

	if configPath == "" {
		return &ValidationError{Field: "configPath", Message: "config path cannot be empty"}
	}

	// Validate before saving
	if err := config.IsValid(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create a new viper instance for this save operation
	v := viper.New()
	setViperConfigFromPath(v, configPath)
	// Set values directly - viper can handle struct serialization
	v.Set("type", config.Type)
	v.Set("repo", config.Repo)
	v.Set("changelog", config.Changelog)

	if config.Git.TagTemplate != "" || config.Git.CommitTemplate != "" {
		v.Set("git", config.Git)
	}

	if len(config.ChangeTypes) > 0 {
		v.Set("change_types", config.ChangeTypes)
	}

	if config.Type == RepositoryTypeMonorepo {
		if len(config.Packages) > 0 {
			v.Set("packages", config.Packages)
		}
	} else {
		v.Set("package", config.Package)
	}

	// Write the config file
	if err := v.SafeWriteConfig(); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadFromDefaultPath loads configuration from the default Shipyard config path
func LoadFromDefaultPath() (*ProjectConfig, error) {
	return LoadFromFile(".shipyard/config.yaml")
}

// SaveToDefaultPath saves configuration to the default Shipyard config path
func SaveToDefaultPath(config *ProjectConfig) error {
	return SaveToFile(config, ".shipyard/config.yaml")
}

// LoadRemoteConfig loads a remote configuration from various sources
func LoadRemoteConfig(remoteURL string, forceFresh bool) (*ProjectConfig, error) {
	if !IsValidRemoteURL(remoteURL) {
		return nil, fmt.Errorf("invalid remote URL format: %s", remoteURL)
	}

	fetcher := NewRemoteConfigFetcher("")
	v, err := fetcher.FetchRemoteConfig(remoteURL, forceFresh)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch remote config: %w", err)
	}

	// Unmarshal into our struct
	var config ProjectConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal remote config: %w", err)
	}

	// Validate the loaded config (but allow configs without packages for base configs)
	if err := validateRemoteConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid remote configuration: %w", err)
	}

	return &config, nil
}

// validateRemoteConfig validates a remote config (more lenient than local configs)
func validateRemoteConfig(config *ProjectConfig) error {
	// Basic validation - remote configs don't need to have packages
	if config.Type == "" {
		return &ValidationError{Field: "type", Message: "repository type is required"}
	}

	if config.Type != RepositoryTypeMonorepo && config.Type != RepositoryTypeSingleRepo {
		return &ValidationError{Field: "type", Message: "repository type must be 'monorepo' or 'single-repo'"}
	}

	// Remote configs should NOT contain packages - they should be base configs
	if config.Type == RepositoryTypeMonorepo && len(config.Packages) > 0 {
		return &ValidationError{Field: "packages", Message: "remote base configurations should not contain packages - packages should be defined in the extending project"}
	}

	if config.Type == RepositoryTypeSingleRepo && config.Package.Name != "" {
		return &ValidationError{Field: "package", Message: "remote base configurations should not contain package definition - package should be defined in the extending project"}
	}

	// Don't require repo URL in remote configs as they're meant to be extended

	return nil
}

// LoadRemoteTemplate loads a remote template from various sources
func LoadRemoteTemplate(templateURL string, forceFresh bool) (string, error) {
	if !IsValidRemoteURL(templateURL) {
		return "", fmt.Errorf("invalid remote template URL format: %s", templateURL)
	}

	fetcher := NewRemoteConfigFetcher("")
	return fetcher.FetchRemoteTemplate(templateURL, forceFresh)
}

// ClearRemoteConfigCache clears the remote configuration cache
func ClearRemoteConfigCache() error {
	fetcher := NewRemoteConfigFetcher("")
	return fetcher.ClearCache()
}

// ListCachedRemoteConfigs returns a list of cached remote configurations
func ListCachedRemoteConfigs() ([]RemoteConfigCache, error) {
	fetcher := NewRemoteConfigFetcher("")
	return fetcher.ListCachedConfigs()
}

// setViperConfigFromPath sets up viper configuration from a file path
func setViperConfigFromPath(v *viper.Viper, path string) {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	ext = strings.TrimPrefix(ext, ".")

	v.SetConfigName(name)
	v.SetConfigType(ext)
	v.AddConfigPath(dir)
}

// loadBaseConfig loads a base configuration for inheritance
func loadBaseConfig(extends, currentConfigPath string) (*viper.Viper, error) {
	if extends == "" {
		return nil, errors.New("no base config to fetch")
	}

	// Handle remote configs
	if IsValidRemoteURL(extends) {
		fetcher := NewRemoteConfigFetcher("")
		return fetcher.FetchRemoteConfig(extends, false)
	}

	// Handle relative paths
	var basePath string
	if !filepath.IsAbs(extends) {
		// Make relative to current config file
		basePath = filepath.Join(filepath.Dir(currentConfigPath), extends)
	} else {
		basePath = extends
	}

	// Check if base config exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("base config file does not exist: %s", basePath)
	}

	// Load base config
	v := viper.New()
	setViperConfigFromPath(v, basePath)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read base config: %w", err)
	}

	return v, nil
}
