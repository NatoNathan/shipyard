package detect

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/NatoNathan/shipyard/internal/config"
)

// DetectPackages scans a directory tree and detects packages based on ecosystem markers
func DetectPackages(rootPath string) ([]config.Package, error) {
	var packages []config.Package
	seen := make(map[string]bool) // Track seen paths to avoid duplicates

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories except .git
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != ".git" {
			return filepath.SkipDir
		}

		// Skip common directories that shouldn't be scanned
		if info.IsDir() {
			dirName := info.Name()
			if dirName == "node_modules" || dirName == "vendor" || dirName == "__pycache__" ||
			   dirName == ".git" || dirName == "dist" || dirName == "build" || dirName == "target" {
				return filepath.SkipDir
			}
		}

		if info.IsDir() {
			return nil
		}

		// Get the directory containing this file
		dir := filepath.Dir(path)

		// Skip if we've already processed this directory
		if seen[dir] {
			return nil
		}

		// Detect package based on file markers
		var pkg *config.Package
		var detectErr error

		switch info.Name() {
		case "go.mod":
			pkg, detectErr = detectGoPackage(rootPath, dir, path)
		case "package.json":
			pkg, detectErr = detectNPMPackage(rootPath, dir, path)
		case "pyproject.toml":
			pkg, detectErr = detectPythonPackage(rootPath, dir, path)
		case "setup.py":
			if !seen[dir] { // Only detect if no pyproject.toml was found
				pkg, detectErr = detectPythonSetupPackage(rootPath, dir, path)
			}
		case "Chart.yaml":
			pkg, detectErr = detectHelmPackage(rootPath, dir, path)
		case "Cargo.toml":
			pkg, detectErr = detectCargoPackage(rootPath, dir, path)
		case "deno.json", "deno.jsonc":
			pkg, detectErr = detectDenoPackage(rootPath, dir, path)
		}

		if detectErr != nil {
			// Log error but continue scanning
			return nil
		}

		if pkg != nil {
			seen[dir] = true
			packages = append(packages, *pkg)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory tree: %w", err)
	}

	return packages, nil
}

// detectGoPackage detects a Go package from go.mod
func detectGoPackage(rootPath, dir, goModPath string) (*config.Package, error) {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, err
	}

	// Parse module name from go.mod
	// Format: module github.com/owner/repo
	lines := strings.Split(string(content), "\n")
	var moduleName string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			moduleName = strings.TrimSpace(strings.TrimPrefix(line, "module"))
			break
		}
	}

	if moduleName == "" {
		return nil, fmt.Errorf("no module name found in go.mod")
	}

	// Extract package name from module path
	parts := strings.Split(moduleName, "/")
	packageName := parts[len(parts)-1]

	return &config.Package{
		Name:      packageName,
		Path:      NormalizePackagePath(rootPath, dir),
		Ecosystem: config.EcosystemGo,
	}, nil
}

// detectNPMPackage detects an NPM package from package.json
func detectNPMPackage(rootPath, dir, packageJSONPath string) (*config.Package, error) {
	content, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return nil, err
	}

	var pkgJSON struct {
		Name string `json:"name"`
	}

	if err := json.Unmarshal(content, &pkgJSON); err != nil {
		return nil, fmt.Errorf("failed to parse package.json: %w", err)
	}

	if pkgJSON.Name == "" {
		return nil, fmt.Errorf("no name found in package.json")
	}

	return &config.Package{
		Name:      pkgJSON.Name,
		Path:      NormalizePackagePath(rootPath, dir),
		Ecosystem: config.EcosystemNPM,
	}, nil
}

// detectPythonPackage detects a Python package from pyproject.toml
func detectPythonPackage(rootPath, dir, pyprojectPath string) (*config.Package, error) {
	content, err := os.ReadFile(pyprojectPath)
	if err != nil {
		return nil, err
	}

	var pyproject struct {
		Project struct {
			Name string `toml:"name"`
		} `toml:"project"`
	}

	if err := toml.Unmarshal(content, &pyproject); err != nil {
		return nil, fmt.Errorf("failed to parse pyproject.toml: %w", err)
	}

	if pyproject.Project.Name == "" {
		return nil, fmt.Errorf("no project name found in pyproject.toml")
	}

	return &config.Package{
		Name:      pyproject.Project.Name,
		Path:      NormalizePackagePath(rootPath, dir),
		Ecosystem: config.EcosystemPython,
	}, nil
}

// detectPythonSetupPackage detects a Python package from setup.py
func detectPythonSetupPackage(rootPath, dir, setupPath string) (*config.Package, error) {
	content, err := os.ReadFile(setupPath)
	if err != nil {
		return nil, err
	}

	// Try to extract name from setup() call
	// This is a simple regex-based approach, not a full Python parser
	nameRegex := regexp.MustCompile(`name\s*=\s*["']([^"']+)["']`)
	matches := nameRegex.FindStringSubmatch(string(content))

	var packageName string
	if len(matches) > 1 {
		packageName = matches[1]
	} else {
		// Fallback to directory name
		packageName = filepath.Base(dir)
	}

	return &config.Package{
		Name:      packageName,
		Path:      NormalizePackagePath(rootPath, dir),
		Ecosystem: config.EcosystemPython,
	}, nil
}

// NormalizePackagePath normalizes a package path relative to root
func NormalizePackagePath(rootPath, pkgPath string) string {
	relPath, err := filepath.Rel(rootPath, pkgPath)
	if err != nil {
		return "./"
	}

	if relPath == "." {
		return "./"
	}

	return "./" + relPath
}
// detectHelmPackage detects a Helm chart from Chart.yaml
func detectHelmPackage(rootPath, dir, chartPath string) (*config.Package, error) {
	content, err := os.ReadFile(chartPath)
	if err != nil {
		return nil, err
	}

	var chart struct {
		Name string `json:"name"`
	}

	// Try to parse as YAML (Chart.yaml is YAML format)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name:") {
			chart.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			break
		}
	}

	if chart.Name == "" {
		return nil, fmt.Errorf("no name found in Chart.yaml")
	}

	return &config.Package{
		Name:      chart.Name,
		Path:      NormalizePackagePath(rootPath, dir),
		Ecosystem: config.EcosystemHelm,
	}, nil
}

// detectCargoPackage detects a Rust/Cargo package from Cargo.toml
func detectCargoPackage(rootPath, dir, cargoPath string) (*config.Package, error) {
	content, err := os.ReadFile(cargoPath)
	if err != nil {
		return nil, err
	}

	var cargo struct {
		Package struct {
			Name string `toml:"name"`
		} `toml:"package"`
	}

	if err := toml.Unmarshal(content, &cargo); err != nil {
		return nil, fmt.Errorf("failed to parse Cargo.toml: %w", err)
	}

	if cargo.Package.Name == "" {
		return nil, fmt.Errorf("no package name found in Cargo.toml")
	}

	return &config.Package{
		Name:      cargo.Package.Name,
		Path:      NormalizePackagePath(rootPath, dir),
		Ecosystem: config.EcosystemCargo,
	}, nil
}

// detectDenoPackage detects a Deno project from deno.json or deno.jsonc
func detectDenoPackage(rootPath, dir, denoPath string) (*config.Package, error) {
	content, err := os.ReadFile(denoPath)
	if err != nil {
		return nil, err
	}

	var denoConfig struct {
		Name string `json:"name"`
	}

	if err := json.Unmarshal(content, &denoConfig); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filepath.Base(denoPath), err)
	}

	packageName := denoConfig.Name
	if packageName == "" {
		// Fallback to directory name if no name field
		packageName = filepath.Base(dir)
	}

	return &config.Package{
		Name:      packageName,
		Path:      NormalizePackagePath(rootPath, dir),
		Ecosystem: config.EcosystemDeno,
	}, nil
}
