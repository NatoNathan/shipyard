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

	Changelog ChangelogConfig `mapstructure:"changelog" json:"changelog" yaml:"changelog"`

	// For monorepo projects
	Packages []Package `mapstructure:"packages" json:"packages,omitempty" yaml:"packages,omitempty"`
	// For single-repo projects
	Package Package `mapstructure:"package" json:"package,omitempty" yaml:"package,omitempty"`
}

// ChangelogConfig represents the changelog configuration
type ChangelogConfig struct {
	Template       string `mapstructure:"template" json:"template" yaml:"template"`
	HeaderTemplate string `mapstructure:"header_template" json:"header_template,omitempty" yaml:"header_template,omitempty"`
	FooterTemplate string `mapstructure:"footer_template" json:"footer_template,omitempty" yaml:"footer_template,omitempty"`
}

// Package represents a package configuration within a project
type Package struct {
	Name      string `mapstructure:"name" json:"name" yaml:"name"`                // e.g., "api", "frontend"
	Path      string `mapstructure:"path" json:"path" yaml:"path"`                // e.g., "packages/api", "packages/frontend"
	Manifest  string `mapstructure:"manifest" json:"manifest" yaml:"manifest"`    // e.g., "packages/api/package.json"
	Ecosystem string `mapstructure:"ecosystem" json:"ecosystem" yaml:"ecosystem"` // e.g., "npm", "go", "python"
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
	if c.Changelog.HeaderTemplate != "" {
		changelogMap["header_template"] = c.Changelog.HeaderTemplate
	}
	if c.Changelog.FooterTemplate != "" {
		changelogMap["footer_template"] = c.Changelog.FooterTemplate
	}
	if len(changelogMap) > 0 {
		configMap["changelog"] = changelogMap
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

// IsValid performs basic validation on the package configuration
func (p *Package) IsValid() error {
	if p.Name == "" {
		return &ValidationError{Field: "name", Message: "package name is required"}
	}

	if p.Path == "" {
		return &ValidationError{Field: "path", Message: "package path is required"}
	}

	if p.Ecosystem == "" {
		return &ValidationError{Field: "ecosystem", Message: "package ecosystem is required"}
	}

	return nil
}

// ToMap converts the Package to a map[string]interface{} suitable for serialization
func (p *Package) ToMap() map[string]interface{} {
	packageMap := map[string]interface{}{
		"name":      p.Name,
		"path":      p.Path,
		"ecosystem": p.Ecosystem,
	}
	if p.Manifest != "" {
		packageMap["manifest"] = p.Manifest
	}
	return packageMap
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

	// Handle remote configs (not implemented yet)
	if strings.HasPrefix(extends, "github:") {
		return nil, errors.New("github: remote config loading not implemented yet")
	} else if strings.HasPrefix(extends, "https://") {
		return nil, errors.New("https: remote config loading not implemented yet")
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
