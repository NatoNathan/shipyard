package handlers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NatoNathan/shipyard/pkg/config"
)

type EcosystemPackage struct {
	Name      string                  `json:"name"`
	Path      string                  `json:"path"`
	Ecosystem config.PackageEcosystem `json:"ecosystem"`
	Manifest  string                  `json:"manifest"`
	Version   string                  `json:"version,omitempty"` // Optional version field
}

func (p *EcosystemPackage) ToConfigPackage() *config.Package {
	return &config.Package{
		Name:      p.Name,
		Path:      p.Path,
		Ecosystem: p.Ecosystem,
		Manifest:  p.Manifest,
	}
}

// EcosystemHandler defines the interface for handling different package ecosystems
type EcosystemHandler interface {
	// GetManifestFile returns the expected manifest file name for this ecosystem
	GetManifestFile() string
	// GetEcosystem returns the ecosystem type this handler supports
	GetEcosystem() config.PackageEcosystem

	// LoadPackage loads package information from the given path
	LoadPackage(path string) (*EcosystemPackage, error)
	// UpdateVersion updates the package version in the manifest file
	UpdateVersion(path string, version string) error
}

// Handler registry
var ecosystemHandlers = map[config.PackageEcosystem]EcosystemHandler{
	config.EcosystemNPM:  &NPMHandler{},
	config.EcosystemGo:   &GoHandler{},
	config.EcosystemHelm: &HelmHandler{},
}

// RegisterEcosystemHandler registers a new ecosystem handler
func RegisterEcosystemHandler(handler EcosystemHandler) {
	ecosystemHandlers[handler.GetEcosystem()] = handler
}

// GetRegisteredEcosystems returns all registered ecosystems
func GetRegisteredEcosystems() []config.PackageEcosystem {
	ecosystems := make([]config.PackageEcosystem, 0, len(ecosystemHandlers))
	for ecosystem := range ecosystemHandlers {
		ecosystems = append(ecosystems, ecosystem)
	}
	return ecosystems
}

// GetHandler returns the handler for the given ecosystem
func GetHandler(ecosystem config.PackageEcosystem) (EcosystemHandler, bool) {
	handler, ok := ecosystemHandlers[ecosystem]
	return handler, ok
}

// LoadPackage loads package information using the appropriate handler
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

// ScanForPackages scans the given root directory for package configurations
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
					// Log error but continue scanning
					continue
				}
				packages = append(packages, pkg.ToConfigPackage())
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return packages, nil
}
