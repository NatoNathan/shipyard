package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/NatoNathan/shipyard/pkg/config"
)

// NPMHandler handles NPM package ecosystems
type NPMHandler struct{}

func (h *NPMHandler) GetEcosystem() config.PackageEcosystem {
	return config.EcosystemNPM
}

func (h *NPMHandler) GetManifestFile() string {
	return "package.json"
}

type PackageJSON struct {
	Name    string                 `json:"name"`
	Version string                 `json:"version"`
	rest    map[string]interface{} `json:"-"` // To capture any additional fields, to write back if needed
}

func (p *PackageJSON) save(manifestPath string) error {
	// Marshal the package JSON back to a file
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal package JSON: %w", err)
	}
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write package manifest %s: %w", manifestPath, err)
	}
	return nil
}

func readPackageJSON(manifestPath string) (*PackageJSON, error) {
	// read the package.json file
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

func (h *NPMHandler) LoadPackage(path string) (*EcosystemPackage, error) {
	// if path has a package.json, use that
	// otherwise, add the default manifest to the path
	var manifestPath string
	if filepath.Ext(path) == ".json" {
		manifestPath = path
	} else {
		manifestPath = filepath.Join(path, h.GetManifestFile())
	}

	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("manifest file %s does not exist for ecosystem %s", manifestPath, h.GetEcosystem())
	}

	// read the package.json file
	packageJson, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package manifest %s: %w", manifestPath, err)
	}

	var pkgJSON PackageJSON
	if err := json.Unmarshal(packageJson, &pkgJSON); err != nil {
		return nil, fmt.Errorf("failed to parse package manifest %s: %w", manifestPath, err)
	}

	return &EcosystemPackage{
		Name:      pkgJSON.Name,
		Path:      filepath.Dir(manifestPath),
		Manifest:  manifestPath,
		Ecosystem: h.GetEcosystem(),
		Version:   pkgJSON.Version,
	}, nil
}

func (h *NPMHandler) UpdateVersion(path string, version string) error {
	// This function is a placeholder for future implementation
	// NPM packages typically have a version in the package.json file
	// The version is managed within the package.json itself
	manifestPath := filepath.Join(path, h.GetManifestFile())
	pkgJSON, err := readPackageJSON(manifestPath)
	if err != nil {
		return err
	}

	// Update the version field
	pkgJSON.Version = version
	return pkgJSON.save(manifestPath)
}
