package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NatoNathan/shipyard/pkg/config"
)

// GoHandler handles Go module ecosystems
type GoHandler struct{}

func (h *GoHandler) GetEcosystem() config.PackageEcosystem {
	return config.EcosystemGo
}

func (h *GoHandler) GetManifestFile() string {
	return "go.mod"
}

func (h *GoHandler) getVersion(path string) string {
	// Go modules do not have a version in the manifest file
	// The version is typically managed with git tags and we can fallback to a `.version` file
	versionFile := filepath.Join(path, ".version")
	if _, err := os.Stat(versionFile); err == nil {
		versionBytes, _ := os.ReadFile(versionFile)
		return strings.TrimSpace(string(versionBytes))
	}
	return "latest"
}

func (h *GoHandler) LoadPackage(path string) (*EcosystemPackage, error) {
	// if path has a go.mod, use that
	// otherwise, add the default manifest to the path
	var manifestPath string
	if filepath.Ext(path) == ".mod" {
		manifestPath = path
	} else {
		manifestPath = filepath.Join(path, h.GetManifestFile())
	}
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("manifest file %s does not exist for ecosystem %s", manifestPath, h.GetEcosystem())
	}

	// read the go.mod file to extract module name
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

	return &EcosystemPackage{
		Name:      moduleName,
		Path:      filepath.Dir(manifestPath),
		Manifest:  manifestPath,
		Ecosystem: h.GetEcosystem(),
		Version:   h.getVersion(filepath.Dir(manifestPath)),
	}, nil
}

func (h *GoHandler) UpdateVersion(path string, version string) error {
	// This function is a placeholder for future implementation
	// Go modules typically do not have a version in the manifest file
	// The version is managed with git tags or a separate .version file
	_, err := os.ReadFile(filepath.Join(filepath.Dir(path), ".version"))
	if err != nil {
		return nil // No version file, nothing to update
	}

	// Update the version field
	if err := os.WriteFile(filepath.Join(filepath.Dir(path), ".version"), []byte(version), 0644); err != nil {
		return fmt.Errorf("failed to write .version file: %w", err)
	}
	return nil
}
