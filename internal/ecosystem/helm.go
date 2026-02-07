package ecosystem

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/NatoNathan/shipyard/pkg/semver"
	"gopkg.in/yaml.v3"
)

var _ Handler = (*HelmEcosystem)(nil)

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

// UpdateVersion updates the version and appVersion in Chart.yaml using regex
// replacement to preserve YAML comments and formatting.
func (h *HelmEcosystem) UpdateVersion(version semver.Version) error {
	chartPath := filepath.Join(h.path, "Chart.yaml")

	// Read existing content
	content, err := os.ReadFile(chartPath)
	if err != nil {
		return fmt.Errorf("failed to read Chart.yaml: %w", err)
	}

	versionStr := version.String()

	// Replace version field using regex to preserve formatting/comments
	versionRe := regexp.MustCompile(`(?m)^(version:\s*)(.+)$`)
	newContent := versionRe.ReplaceAll(content, []byte(fmt.Sprintf(`${1}%s`, versionStr)))

	// Also update appVersion if it exists
	appVersionRe := regexp.MustCompile(`(?m)^(appVersion:\s*)(.+)$`)
	newContent = appVersionRe.ReplaceAll(newContent, []byte(fmt.Sprintf(`${1}"%s"`, versionStr)))

	if string(newContent) == string(content) {
		return fmt.Errorf("no version field found in Chart.yaml")
	}

	return os.WriteFile(chartPath, newContent, 0644)
}

// GetVersionFiles returns paths to all version-containing files
func (h *HelmEcosystem) GetVersionFiles() []string {
	chartPath := filepath.Join(h.path, "Chart.yaml")
	if _, err := os.Stat(chartPath); err == nil {
		return []string{"Chart.yaml"}
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
