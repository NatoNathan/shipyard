package handlers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NatoNathan/shipyard/pkg/config"
)

// EcosystemPackage represents a package in a specific ecosystem with metadata
type EcosystemPackage struct {
	Name      string                  `json:"name"`
	Path      string                  `json:"path"`
	Ecosystem config.PackageEcosystem `json:"ecosystem"`
	Manifest  string                  `json:"manifest"`
	Version   string                  `json:"version,omitempty"` // Optional version field
}

// ToConfigPackage converts an EcosystemPackage to a config.Package
func (p *EcosystemPackage) ToConfigPackage() *config.Package {
	return &config.Package{
		Name:      p.Name,
		Path:      p.Path,
		Ecosystem: p.Ecosystem,
		Manifest:  p.Manifest,
	}
}

// EcosystemHandler defines the interface for handling different package ecosystems.
// Each ecosystem (NPM, Go, Helm, etc.) should implement this interface to provide
// standardized package management operations.
type EcosystemHandler interface {
	// GetManifestFile returns the expected manifest file name for this ecosystem
	// (e.g., "package.json" for NPM, "go.mod" for Go, "Chart.yaml" for Helm)
	GetManifestFile() string

	// GetEcosystem returns the ecosystem type this handler supports
	GetEcosystem() config.PackageEcosystem

	// LoadPackage loads package information from the given path
	// The path should point to a directory containing the manifest file
	LoadPackage(path string) (*EcosystemPackage, error)

	// UpdateVersion updates the package version in the manifest file
	// This is used during release operations to bump versions
	UpdateVersion(path string, version string) error
}

// Handler registry - stores all registered ecosystem handlers
var ecosystemHandlers = map[config.PackageEcosystem]EcosystemHandler{
	config.EcosystemNPM:  &NPMHandler{},
	config.EcosystemGo:   &GoHandler{},
	config.EcosystemHelm: &HelmHandler{},
}

// RegisterEcosystemHandler registers a new ecosystem handler.
// This allows for dynamic registration of new ecosystem types at runtime.
func RegisterEcosystemHandler(handler EcosystemHandler) {
	ecosystemHandlers[handler.GetEcosystem()] = handler
}

// GetRegisteredEcosystems returns all registered ecosystem types.
// This is useful for validation and UI display purposes.
func GetRegisteredEcosystems() []config.PackageEcosystem {
	ecosystems := make([]config.PackageEcosystem, 0, len(ecosystemHandlers))
	for ecosystem := range ecosystemHandlers {
		ecosystems = append(ecosystems, ecosystem)
	}
	return ecosystems
}

// GetHandler returns the handler for the given ecosystem.
// Returns the handler and a boolean indicating if the ecosystem is supported.
func GetHandler(ecosystem config.PackageEcosystem) (EcosystemHandler, bool) {
	handler, ok := ecosystemHandlers[ecosystem]
	return handler, ok
}

// LoadPackage loads package information using the appropriate handler.
// This is a convenience function that looks up the handler by ecosystem type
// and delegates to the handler's LoadPackage method.
func LoadPackage(ecosystem config.PackageEcosystem, path string) (*config.Package, error) {
	if ecosystem == "" {
		return nil, fmt.Errorf("ecosystem cannot be empty")
	}

	handler, ok := ecosystemHandlers[ecosystem]
	if !ok {
		return nil, fmt.Errorf("unknown ecosystem: %s", ecosystem)
	}

	pkg, err := handler.LoadPackage(path)
	if err != nil {
		return nil, err
	}
	return pkg.ToConfigPackage(), nil
}

// ScanForPackages scans the given root directory for package configurations.
// It walks the directory tree and identifies packages based on manifest files.
// This is useful for auto-discovery of packages in a monorepo structure.
func ScanForPackages(root string) ([]*config.Package, error) {
	var packages []*config.Package

	// Walk the directory tree
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Check if the file is a known manifest file
		for _, handler := range ecosystemHandlers {
			if info.Name() == handler.GetManifestFile() {
				pkg, err := handler.LoadPackage(filepath.Dir(path))
				if err != nil {
					// Log error but continue scanning - some manifest files might be invalid
					continue
				}
				packages = append(packages, pkg.ToConfigPackage())
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory %s: %w", root, err)
	}
	return packages, nil
}

// UpdatePackageVersion updates the version of a package using its ecosystem handler.
// This is a convenience function that looks up the handler and delegates to UpdateVersion.
func UpdatePackageVersion(pkg *config.Package, version string) error {
	handler, ok := GetHandler(pkg.Ecosystem)
	if !ok {
		return fmt.Errorf("unsupported ecosystem: %s", pkg.Ecosystem)
	}
	return handler.UpdateVersion(pkg.Path, version)
}

// ValidatePackage validates that a package configuration is valid for its ecosystem.
// It checks if the manifest file exists and can be loaded successfully.
func ValidatePackage(pkg *config.Package) error {
	handler, ok := GetHandler(pkg.Ecosystem)
	if !ok {
		return fmt.Errorf("unsupported ecosystem: %s", pkg.Ecosystem)
	}

	_, err := handler.LoadPackage(pkg.Path)
	return err
}
