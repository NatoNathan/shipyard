package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NatoNathan/shipyard/pkg/config"
)

// GoHandler handles Go module ecosystems.
// It manages Go modules by parsing go.mod files and handling version through .version files.
type GoHandler struct{}

// GetEcosystem returns the ecosystem type for Go modules
func (h *GoHandler) GetEcosystem() config.PackageEcosystem {
	return config.EcosystemGo
}

// GetManifestFile returns the manifest file name for Go modules
func (h *GoHandler) GetManifestFile() string {
	return "go.mod"
}

// getVersion retrieves the version from a .version file or returns "latest" as fallback.
// Go modules don't have versions in go.mod, so we use a separate version file.
func (h *GoHandler) getVersion(path string) string {
	versionFile := filepath.Join(path, ".version")
	if _, err := os.Stat(versionFile); err == nil {
		versionBytes, err := os.ReadFile(versionFile)
		if err == nil {
			return strings.TrimSpace(string(versionBytes))
		}
	}
	return "latest"
}

// LoadPackage loads Go module information from the given path.
// It parses the go.mod file to extract the module name and retrieves version information.
func (h *GoHandler) LoadPackage(path string) (*EcosystemPackage, error) {
	// Handle both direct go.mod file paths and directory paths
	var manifestPath string
	if filepath.Ext(path) == ".mod" {
		manifestPath = path
	} else {
		manifestPath = filepath.Join(path, h.GetManifestFile())
	}

	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("manifest file %s does not exist for ecosystem %s", manifestPath, h.GetEcosystem())
	}

	// Read and parse the go.mod file to extract module name
	goMod, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read go.mod file %s: %w", manifestPath, err)
	}

	// Parse the module name from go.mod (first line: "module <name>")
	lines := strings.Split(string(goMod), "\n")
	var moduleName string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			moduleName = strings.TrimSpace(strings.TrimPrefix(line, "module"))
			break
		}
	}

	if moduleName == "" {
		return nil, fmt.Errorf("could not find module name in go.mod file %s", manifestPath)
	}

	packagePath := filepath.Dir(manifestPath)
	return &EcosystemPackage{
		Name:      moduleName,
		Path:      packagePath,
		Manifest:  manifestPath,
		Ecosystem: h.GetEcosystem(),
		Version:   h.getVersion(packagePath),
	}, nil
}

// UpdateVersion updates the version for a Go module by writing to a .version file.
// Since Go modules don't store version in go.mod, we use a separate version file.
func (h *GoHandler) UpdateVersion(path string, version string) error {
	versionFile := filepath.Join(path, ".version")
	if err := os.WriteFile(versionFile, []byte(version), 0644); err != nil {
		return fmt.Errorf("failed to write version file %s: %w", versionFile, err)
	}
	return nil
}
