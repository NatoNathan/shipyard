package ecosystem

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/BurntSushi/toml"
	"github.com/NatoNathan/shipyard/pkg/semver"
)

// PythonEcosystem handles version management for Python projects
type PythonEcosystem struct {
	path string
}

// NewPythonEcosystem creates a new Python ecosystem handler
func NewPythonEcosystem(path string) *PythonEcosystem {
	return &PythonEcosystem{path: path}
}

// ReadVersion reads the current version from Python version files
// Checks in order: pyproject.toml, __version__.py, setup.py
func (p *PythonEcosystem) ReadVersion() (semver.Version, error) {
	// Try pyproject.toml first
	pyprojectPath := filepath.Join(p.path, "pyproject.toml")
	if _, err := os.Stat(pyprojectPath); err == nil {
		version, err := p.readVersionFromPyproject(pyprojectPath)
		if err == nil {
			return version, nil
		}
	}

	// Try __version__.py
	versionPyPath := filepath.Join(p.path, "__version__.py")
	if _, err := os.Stat(versionPyPath); err == nil {
		return p.readVersionFromVersionPy(versionPyPath)
	}

	// Try setup.py
	setupPyPath := filepath.Join(p.path, "setup.py")
	if _, err := os.Stat(setupPyPath); err == nil {
		return p.readVersionFromSetupPy(setupPyPath)
	}

	return semver.Version{}, fmt.Errorf("no version file found in Python project at %s", p.path)
}

// UpdateVersion updates the version in Python version files
func (p *PythonEcosystem) UpdateVersion(version semver.Version) error {
	updated := false

	// Update pyproject.toml if it exists
	pyprojectPath := filepath.Join(p.path, "pyproject.toml")
	if _, err := os.Stat(pyprojectPath); err == nil {
		if err := p.updatePyproject(pyprojectPath, version); err != nil {
			return err
		}
		updated = true
	}

	// Update __version__.py if it exists
	versionPyPath := filepath.Join(p.path, "__version__.py")
	if _, err := os.Stat(versionPyPath); err == nil {
		if err := p.updateVersionPy(versionPyPath, version); err != nil {
			return err
		}
		updated = true
	}

	// Update setup.py if it exists
	setupPyPath := filepath.Join(p.path, "setup.py")
	if _, err := os.Stat(setupPyPath); err == nil {
		if err := p.updateSetupPy(setupPyPath, version); err != nil {
			return err
		}
		updated = true
	}

	if !updated {
		return fmt.Errorf("no version files to update in Python project at %s", p.path)
	}

	return nil
}

// GetVersionFiles returns paths to all version-containing files
func (p *PythonEcosystem) GetVersionFiles() []string {
	var files []string

	candidates := []string{
		filepath.Join(p.path, "pyproject.toml"),
		filepath.Join(p.path, "__version__.py"),
		filepath.Join(p.path, "setup.py"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			files = append(files, path)
		}
	}

	return files
}

// readVersionFromPyproject extracts version from pyproject.toml
func (p *PythonEcosystem) readVersionFromPyproject(path string) (semver.Version, error) {
	var config struct {
		Tool struct {
			Poetry struct {
				Version string `toml:"version"`
			} `toml:"poetry"`
		} `toml:"tool"`
		Project struct {
			Version string `toml:"version"`
		} `toml:"project"`
	}

	if _, err := toml.DecodeFile(path, &config); err != nil {
		return semver.Version{}, fmt.Errorf("failed to parse pyproject.toml: %w", err)
	}

	// Check [tool.poetry] format
	if config.Tool.Poetry.Version != "" {
		return semver.Parse(config.Tool.Poetry.Version)
	}

	// Check [project] format
	if config.Project.Version != "" {
		return semver.Parse(config.Project.Version)
	}

	return semver.Version{}, fmt.Errorf("no version found in pyproject.toml")
}

// readVersionFromVersionPy extracts version from __version__.py
func (p *PythonEcosystem) readVersionFromVersionPy(path string) (semver.Version, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return semver.Version{}, fmt.Errorf("failed to read __version__.py: %w", err)
	}

	// Match: __version__ = "1.2.3" or __version__ = '1.2.3'
	re := regexp.MustCompile(`(?m)^__version__\s*=\s*["']([0-9]+\.[0-9]+\.[0-9]+)["']`)
	matches := re.FindSubmatch(content)

	if len(matches) < 2 {
		return semver.Version{}, fmt.Errorf("no __version__ found in %s", path)
	}

	return semver.Parse(string(matches[1]))
}

// readVersionFromSetupPy extracts version from setup.py
func (p *PythonEcosystem) readVersionFromSetupPy(path string) (semver.Version, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return semver.Version{}, fmt.Errorf("failed to read setup.py: %w", err)
	}

	// Match: version="1.2.3" or version='1.2.3'
	re := regexp.MustCompile(`version\s*=\s*["']([0-9]+\.[0-9]+\.[0-9]+)["']`)
	matches := re.FindSubmatch(content)

	if len(matches) < 2 {
		return semver.Version{}, fmt.Errorf("no version found in setup.py")
	}

	return semver.Parse(string(matches[1]))
}

// updatePyproject updates version in pyproject.toml
func (p *PythonEcosystem) updatePyproject(path string, version semver.Version) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read pyproject.toml: %w", err)
	}

	// Try both [tool.poetry] and [project] formats
	re := regexp.MustCompile(`(version\s*=\s*)["']([0-9]+\.[0-9]+\.[0-9]+)["']`)
	newContent := re.ReplaceAll(content, []byte(fmt.Sprintf(`${1}"%s"`, version.String())))

	if string(newContent) == string(content) {
		return fmt.Errorf("no version field found in pyproject.toml")
	}

	return os.WriteFile(path, newContent, 0644)
}

// updateVersionPy updates version in __version__.py
func (p *PythonEcosystem) updateVersionPy(path string, version semver.Version) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read __version__.py: %w", err)
	}

	re := regexp.MustCompile(`(__version__\s*=\s*)["']([0-9]+\.[0-9]+\.[0-9]+)["']`)
	newContent := re.ReplaceAll(content, []byte(fmt.Sprintf(`${1}"%s"`, version.String())))

	if string(newContent) == string(content) {
		return fmt.Errorf("no __version__ found in %s", path)
	}

	return os.WriteFile(path, newContent, 0644)
}

// updateSetupPy updates version in setup.py
func (p *PythonEcosystem) updateSetupPy(path string, version semver.Version) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read setup.py: %w", err)
	}

	re := regexp.MustCompile(`(version\s*=\s*)["']([0-9]+\.[0-9]+\.[0-9]+)["']`)
	newContent := re.ReplaceAll(content, []byte(fmt.Sprintf(`${1}"%s"`, version.String())))

	if string(newContent) == string(content) {
		return fmt.Errorf("no version field found in setup.py")
	}

	return os.WriteFile(path, newContent, 0644)
}

// DetectPythonEcosystem checks if a directory contains a Python project
func DetectPythonEcosystem(path string) bool {
	// Check for pyproject.toml
	if _, err := os.Stat(filepath.Join(path, "pyproject.toml")); err == nil {
		return true
	}

	// Check for setup.py
	if _, err := os.Stat(filepath.Join(path, "setup.py")); err == nil {
		return true
	}

	// Check for __version__.py
	if _, err := os.Stat(filepath.Join(path, "__version__.py")); err == nil {
		return true
	}

	// Check for requirements.txt
	if _, err := os.Stat(filepath.Join(path, "requirements.txt")); err == nil {
		return true
	}

	return false
}
