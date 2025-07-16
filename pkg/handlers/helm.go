package handlers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NatoNathan/shipyard/pkg/config"
	"gopkg.in/yaml.v3"
)

// HelmHandler handles Helm chart ecosystems.
// It manages Helm charts by parsing Chart.yaml files and updating versions.
type HelmHandler struct{}

// GetEcosystem returns the ecosystem type for Helm charts
func (h *HelmHandler) GetEcosystem() config.PackageEcosystem {
	return config.EcosystemHelm
}

// GetManifestFile returns the manifest file name for Helm charts
func (h *HelmHandler) GetManifestFile() string {
	return "Chart.yaml"
}

// ChartYAML represents the structure of a Chart.yaml file.
// We use a minimal structure with a rest field to preserve unknown fields.
type ChartYAML struct {
	Name    string                 `yaml:"name"`
	Version string                 `yaml:"version"`
	rest    map[string]interface{} `yaml:"-"` // To capture any additional fields, to write back if needed
}

// save writes the ChartYAML back to the manifest file with proper formatting
func (c *ChartYAML) save(manifestPath string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal Chart.yaml: %w", err)
	}
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write Chart.yaml %s: %w", manifestPath, err)
	}
	return nil
}

// readManifest reads and parses a Chart.yaml file
func readManifest(manifestPath string) (*ChartYAML, error) {
	chartYaml, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Chart.yaml file %s: %w", manifestPath, err)
	}

	var chart ChartYAML
	if err := yaml.Unmarshal(chartYaml, &chart); err != nil {
		return nil, fmt.Errorf("failed to parse Chart.yaml file %s: %w", manifestPath, err)
	}

	return &chart, nil
}

// LoadPackage loads Helm chart information from the given path.
// It parses the Chart.yaml file to extract chart name and version.
func (h *HelmHandler) LoadPackage(path string) (*EcosystemPackage, error) {
	// Handle both direct Chart.yaml file paths and directory paths
	var manifestPath string
	if filepath.Ext(path) == ".yaml" {
		manifestPath = path
	} else {
		manifestPath = filepath.Join(path, h.GetManifestFile())
	}

	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("manifest file %s does not exist for ecosystem %s", manifestPath, h.GetEcosystem())
	}

	chart, err := readManifest(manifestPath)
	if err != nil {
		return nil, err
	}

	return &EcosystemPackage{
		Name:      chart.Name,
		Path:      filepath.Dir(manifestPath),
		Manifest:  manifestPath,
		Ecosystem: h.GetEcosystem(),
		Version:   chart.Version,
	}, nil
}

// UpdateVersion updates the version in a Helm chart's Chart.yaml file
func (h *HelmHandler) UpdateVersion(path string, version string) error {
	manifestPath := filepath.Join(path, h.GetManifestFile())
	chart, err := readManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read Chart.yaml file %s: %w", manifestPath, err)
	}

	// Update the version field
	chart.Version = version
	if err := chart.save(manifestPath); err != nil {
		return fmt.Errorf("failed to save Chart.yaml file %s: %w", manifestPath, err)
	}
	return nil
}
