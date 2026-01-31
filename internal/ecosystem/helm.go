package ecosystem

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NatoNathan/shipyard/pkg/semver"
	"gopkg.in/yaml.v3"
)

// HelmEcosystem handles version management for Helm charts
type HelmEcosystem struct {
	path string
}

// NewHelmEcosystem creates a new Helm ecosystem handler
func NewHelmEcosystem(path string) *HelmEcosystem {
	return &HelmEcosystem{path: path}
}

// HelmChart represents the structure of Chart.yaml
type HelmChart struct {
	APIVersion  string `yaml:"apiVersion"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Type        string `yaml:"type,omitempty"`
	Version     string `yaml:"version"`
	AppVersion  string `yaml:"appVersion,omitempty"`
}

// ReadVersion reads the current version from Chart.yaml
func (h *HelmEcosystem) ReadVersion() (semver.Version, error) {
	chartPath := filepath.Join(h.path, "Chart.yaml")

	content, err := os.ReadFile(chartPath)
	if err != nil {
		return semver.Version{}, fmt.Errorf("failed to read Chart.yaml: %w", err)
	}

	var chart HelmChart
	if err := yaml.Unmarshal(content, &chart); err != nil {
		return semver.Version{}, fmt.Errorf("failed to parse Chart.yaml: %w", err)
	}

	if chart.Version == "" {
		return semver.Version{}, fmt.Errorf("no version field found in Chart.yaml")
	}

	return semver.Parse(chart.Version)
}

// UpdateVersion updates the version in Chart.yaml
func (h *HelmEcosystem) UpdateVersion(version semver.Version) error {
	chartPath := filepath.Join(h.path, "Chart.yaml")

	// Read existing content
	content, err := os.ReadFile(chartPath)
	if err != nil {
		return fmt.Errorf("failed to read Chart.yaml: %w", err)
	}

	// Parse as generic map to preserve structure and comments
	var chartData map[string]interface{}
	if err := yaml.Unmarshal(content, &chartData); err != nil {
		return fmt.Errorf("failed to parse Chart.yaml: %w", err)
	}

	// Update version
	chartData["version"] = version.String()

	// Write back
	newContent, err := yaml.Marshal(chartData)
	if err != nil {
		return fmt.Errorf("failed to marshal Chart.yaml: %w", err)
	}

	return os.WriteFile(chartPath, newContent, 0644)
}

// GetVersionFiles returns paths to all version-containing files
func (h *HelmEcosystem) GetVersionFiles() []string {
	chartPath := filepath.Join(h.path, "Chart.yaml")
	if _, err := os.Stat(chartPath); err == nil {
		return []string{chartPath}
	}
	return []string{}
}

// DetectHelmEcosystem checks if a directory contains a Helm chart
func DetectHelmEcosystem(path string) bool {
	// Check for Chart.yaml
	if _, err := os.Stat(filepath.Join(path, "Chart.yaml")); err == nil {
		return true
	}

	// Check for charts directory (common in Helm projects)
	if _, err := os.Stat(filepath.Join(path, "charts")); err == nil {
		return true
	}

	return false
}
