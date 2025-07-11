package handlers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NatoNathan/shipyard/pkg/config"
	"gopkg.in/yaml.v3"
)

// HelmHandler handles Helm chart ecosystems
type HelmHandler struct{}

func (h *HelmHandler) GetEcosystem() config.PackageEcosystem {
	return config.EcosystemHelm
}

func (h *HelmHandler) GetManifestFile() string {
	return "Chart.yaml"
}

type ChartYAML struct {
	Name    string                 `yaml:"name"`
	Version string                 `yaml:"version"`
	rest    map[string]interface{} `yaml:"-"` // To capture any additional fields, to write back if needed
}

func (c *ChartYAML) save(manifestPath string) error {
	// Marshal the chart YAML back to a file
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal Chart.yaml: %w", err)
	}
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write Chart.yaml %s: %w", manifestPath, err)
	}
	return nil
}

func readManifest(manifestPath string) (*ChartYAML, error) {
	// read the Chart.yaml file
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

func (h *HelmHandler) LoadPackage(path string) (*EcosystemPackage, error) {
	// if path has a Chart.yaml, use that
	// otherwise, add the default manifest to the path
	var manifestPath string
	if filepath.Ext(path) == ".yaml" {
		manifestPath = path
	} else {
		manifestPath = filepath.Join(path, h.GetManifestFile())
	}
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("manifest file %s does not exist for ecosystem %s", manifestPath, h.GetEcosystem())
	}

	// read the Chart.yaml file
	chartYaml, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Chart.yaml file %s: %w", manifestPath, err)
	}

	// Parse the chart name from Chart.yaml

	var chart ChartYAML
	if err := yaml.Unmarshal(chartYaml, &chart); err != nil {
		return nil, fmt.Errorf("failed to parse Chart.yaml file %s: %w", manifestPath, err)
	}

	return &EcosystemPackage{
		Name:      chart.Name,
		Path:      filepath.Dir(manifestPath),
		Manifest:  manifestPath,
		Ecosystem: h.GetEcosystem(),
		Version:   chart.Version,
	}, nil
}

func (h *HelmHandler) UpdateVersion(path string, version string) error {
	// This function is a placeholder for future implementation
	// Helm charts typically have a version in the Chart.yaml file
	// The version is managed within the Chart.yaml itself
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
