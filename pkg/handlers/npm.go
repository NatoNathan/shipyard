package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/NatoNathan/shipyard/pkg/config"
)

// NPMHandler handles NPM package ecosystems.
// It manages NPM packages by parsing package.json files and updating versions.
type NPMHandler struct{}

// GetEcosystem returns the ecosystem type for NPM packages
func (h *NPMHandler) GetEcosystem() config.PackageEcosystem {
	return config.EcosystemNPM
}

// GetManifestFile returns the manifest file name for NPM packages
func (h *NPMHandler) GetManifestFile() string {
	return "package.json"
}

// PackageJSON represents the structure of a package.json file.
// We use a minimal structure with a rest field to preserve unknown fields.
type PackageJSON struct {
	Name    string                 `json:"name"`
	Version string                 `json:"version"`
	rest    map[string]interface{} `json:"-"` // To capture any additional fields, to write back if needed
}

// save writes the PackageJSON back to the manifest file with proper formatting
func (p *PackageJSON) save(manifestPath string) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal package JSON: %w", err)
	}
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write package manifest %s: %w", manifestPath, err)
	}
	return nil
}

// readPackageJSON reads and parses a package.json file
func readPackageJSON(manifestPath string) (*PackageJSON, error) {
	packageJson, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package manifest %s: %w", manifestPath, err)
	}

	var pkgJSON PackageJSON
	if err := json.Unmarshal(packageJson, &pkgJSON); err != nil {
		return nil, fmt.Errorf("failed to parse package manifest %s: %w", manifestPath, err)
	}
	return &pkgJSON, nil
}

// LoadPackage loads NPM package information from the given path.
// It parses the package.json file to extract package name and version.
func (h *NPMHandler) LoadPackage(path string) (*EcosystemPackage, error) {
	// Handle both direct package.json file paths and directory paths
	var manifestPath string
	if filepath.Ext(path) == ".json" {
		manifestPath = path
	} else {
		manifestPath = filepath.Join(path, h.GetManifestFile())
	}

	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("manifest file %s does not exist for ecosystem %s", manifestPath, h.GetEcosystem())
	}

	pkgJSON, err := readPackageJSON(manifestPath)
	if err != nil {
		return nil, err
	}

	return &EcosystemPackage{
		Name:      pkgJSON.Name,
		Path:      filepath.Dir(manifestPath),
		Manifest:  manifestPath,
		Ecosystem: h.GetEcosystem(),
		Version:   pkgJSON.Version,
	}, nil
}

// UpdateVersion updates the version in an NPM package's package.json file
func (h *NPMHandler) UpdateVersion(path string, version string) error {
	manifestPath := filepath.Join(path, h.GetManifestFile())
	pkgJSON, err := readPackageJSON(manifestPath)
	if err != nil {
		return err
	}

	// Update the version field
	pkgJSON.Version = version
	return pkgJSON.save(manifestPath)
}
