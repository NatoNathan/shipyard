package config

import (
	"fmt"
	"path/filepath"
	"slices"
)

// PackageEcosystem represents the type of package ecosystem
type PackageEcosystem string

// Constants for supported ecosystems
const (
	// EcosystemNPM represents the NPM ecosystem
	EcosystemNPM PackageEcosystem = "npm"
	// EcosystemGo represents the Go ecosystem
	EcosystemGo PackageEcosystem = "go"
	// EcosystemHelm represents the Helm ecosystem
	EcosystemHelm PackageEcosystem = "helm"
)

// Package represents a package configuration within a project
type Package struct {
	Name          string           `mapstructure:"name" json:"name" yaml:"name"`                                                   // e.g., "api", "frontend"
	Path          string           `mapstructure:"path" json:"path" yaml:"path"`                                                   // e.g., "packages/api", "packages/frontend"
	Manifest      string           `mapstructure:"manifest" json:"manifest" yaml:"manifest"`                                       // e.g., "packages/api/package.json"
	Ecosystem     PackageEcosystem `mapstructure:"ecosystem" json:"ecosystem" yaml:"ecosystem"`                                    // e.g., "npm", "go", "python"
	ChangelogPath string           `mapstructure:"changelog_path" json:"changelog_path,omitempty" yaml:"changelog_path,omitempty"` // e.g., "CHANGELOG.md", "docs/CHANGELOG.md"
}

// GetSupportedEcosystems returns all supported ecosystems
func GetSupportedEcosystems() []PackageEcosystem {
	return SupportedEcosystems
}

// SupportedEcosystems lists all supported package ecosystems
var SupportedEcosystems = []PackageEcosystem{
	EcosystemNPM,
	EcosystemGo,
	EcosystemHelm,
}

// NewPackage creates a new Package instance
func NewPackage(name, path string, ecosystem PackageEcosystem) *Package {
	return &Package{
		Name:      name,
		Path:      path,
		Ecosystem: ecosystem,
	}
}

// NewPackageFromMap creates a Package from a map
func NewPackageFromMap(data map[string]interface{}) (*Package, error) {
	pkg := &Package{}
	var ok bool
	pkg.Name, ok = data["name"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid or missing package name")
	}
	pkg.Path, ok = data["path"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid or missing package path")
	}
	pkg.Ecosystem, ok = data["ecosystem"].(PackageEcosystem)
	if !ok {
		return nil, fmt.Errorf("invalid or missing package ecosystem")
	}
	pkg.Manifest, ok = data["manifest"].(string)
	if !ok {
		pkg.Manifest = "" // Default to empty if not provided
	}
	if err := pkg.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid package configuration: %w", err)
	}
	return pkg, nil
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
	if !slices.Contains(GetSupportedEcosystems(), p.Ecosystem) {
		return &ValidationError{Field: "ecosystem", Message: fmt.Sprintf("unsupported ecosystem: %s", p.Ecosystem)}
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
	if p.ChangelogPath != "" {
		packageMap["changelog_path"] = p.ChangelogPath
	}
	return packageMap
}

// GetChangelogPath returns the full path to the changelog file for this package
// It combines the package path with the changelog filename
func (p *Package) GetChangelogPath(defaultFilename string) string {
	changelogFilename := p.ChangelogPath
	if changelogFilename == "" {
		changelogFilename = defaultFilename
	}

	if p.Path == "" {
		return changelogFilename
	}

	return filepath.Join(p.Path, changelogFilename)
}
