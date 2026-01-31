package detect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDetectPackages_GoModule tests detection of Go modules
func TestDetectPackages_GoModule(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod file
	goModContent := `module github.com/example/myproject

go 1.21

require (
	github.com/spf13/cobra v1.8.0
)
`
	goModPath := filepath.Join(tempDir, "go.mod")
	require.NoError(t, os.WriteFile(goModPath, []byte(goModContent), 0644))

	// Detect packages
	packages, err := DetectPackages(tempDir)
	require.NoError(t, err)

	// Should detect one Go package
	assert.Len(t, packages, 1)
	assert.Equal(t, "myproject", packages[0].Name)
	assert.Equal(t, "./", packages[0].Path)
	assert.Equal(t, config.EcosystemGo, packages[0].Ecosystem)
}

// TestDetectPackages_NPMPackage tests detection of NPM packages
func TestDetectPackages_NPMPackage(t *testing.T) {
	tempDir := t.TempDir()

	// Create package.json file
	packageJSON := `{
  "name": "my-npm-package",
  "version": "1.0.0",
  "description": "Test package"
}
`
	packageJSONPath := filepath.Join(tempDir, "package.json")
	require.NoError(t, os.WriteFile(packageJSONPath, []byte(packageJSON), 0644))

	// Detect packages
	packages, err := DetectPackages(tempDir)
	require.NoError(t, err)

	// Should detect one NPM package
	assert.Len(t, packages, 1)
	assert.Equal(t, "my-npm-package", packages[0].Name)
	assert.Equal(t, "./", packages[0].Path)
	assert.Equal(t, config.EcosystemNPM, packages[0].Ecosystem)
}

// TestDetectPackages_PythonPackage tests detection of Python packages
func TestDetectPackages_PythonPackage(t *testing.T) {
	tests := []struct {
		name         string
		filename     string
		content      string
		expectedName string
	}{
		{
			name:     "pyproject.toml",
			filename: "pyproject.toml",
			content: `[project]
name = "my-python-package"
version = "0.1.0"
description = "Test package"
`,
			expectedName: "my-python-package",
		},
		{
			name:     "setup.py",
			filename: "setup.py",
			content: `from setuptools import setup

setup(
    name="my-setup-package",
    version="0.1.0",
)
`,
			expectedName: "my-setup-package",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create Python package file
			filePath := filepath.Join(tempDir, tt.filename)
			require.NoError(t, os.WriteFile(filePath, []byte(tt.content), 0644))

			// Detect packages
			packages, err := DetectPackages(tempDir)
			require.NoError(t, err)

			// Should detect one Python package
			assert.Len(t, packages, 1)
			assert.Equal(t, tt.expectedName, packages[0].Name)
			assert.Equal(t, "./", packages[0].Path)
			assert.Equal(t, config.EcosystemPython, packages[0].Ecosystem)
		})
	}
}

// TestDetectPackages_Monorepo tests detection of multiple packages in a monorepo
func TestDetectPackages_Monorepo(t *testing.T) {
	tempDir := t.TempDir()

	// Create multiple packages
	packages := map[string]struct {
		file    string
		content string
	}{
		"packages/core": {
			file:    "go.mod",
			content: "module github.com/example/core\n\ngo 1.21\n",
		},
		"packages/api": {
			file:    "go.mod",
			content: "module github.com/example/api\n\ngo 1.21\n",
		},
		"clients/web": {
			file:    "package.json",
			content: `{"name": "@example/web", "version": "1.0.0"}`,
		},
		"services/worker": {
			file:    "pyproject.toml",
			content: "[project]\nname = \"worker\"\nversion = \"1.0.0\"\n",
		},
	}

	for pkgPath, pkgData := range packages {
		fullPath := filepath.Join(tempDir, pkgPath)
		require.NoError(t, os.MkdirAll(fullPath, 0755))
		filePath := filepath.Join(fullPath, pkgData.file)
		require.NoError(t, os.WriteFile(filePath, []byte(pkgData.content), 0644))
	}

	// Detect packages
	detectedPackages, err := DetectPackages(tempDir)
	require.NoError(t, err)

	// Should detect 4 packages
	assert.Len(t, detectedPackages, 4)

	// Verify package names and paths
	packageNames := make(map[string]bool)
	for _, pkg := range detectedPackages {
		packageNames[pkg.Name] = true
	}

	assert.True(t, packageNames["core"], "Should detect core package")
	assert.True(t, packageNames["api"], "Should detect api package")
	assert.True(t, packageNames["@example/web"], "Should detect web package")
	assert.True(t, packageNames["worker"], "Should detect worker package")
}

// TestDetectPackages_NoPackages tests behavior when no packages are found
func TestDetectPackages_NoPackages(t *testing.T) {
	tempDir := t.TempDir()

	// Create some non-package files
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("# Test"), 0644))

	// Detect packages
	packages, err := DetectPackages(tempDir)
	require.NoError(t, err)

	// Should return empty slice when no packages found
	assert.Empty(t, packages)
}

// TestDetectPackages_NestedStructure tests detection with various nesting levels
func TestDetectPackages_NestedStructure(t *testing.T) {
	tempDir := t.TempDir()

	// Create nested structure
	// Root level package
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module root\n\ngo 1.21\n"), 0644))

	// Nested packages should be detected separately
	nestedDir := filepath.Join(tempDir, "nested", "deep", "package")
	require.NoError(t, os.MkdirAll(nestedDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(nestedDir, "package.json"), []byte(`{"name": "nested-package", "version": "1.0.0"}`), 0644))

	// Detect packages
	packages, err := DetectPackages(tempDir)
	require.NoError(t, err)

	// Should detect both packages
	assert.GreaterOrEqual(t, len(packages), 2, "Should detect at least 2 packages")
}

// TestDetectPackages_MultipleEcosystemsSameDir tests behavior when multiple ecosystem markers exist in same directory
func TestDetectPackages_MultipleEcosystemsSameDir(t *testing.T) {
	tempDir := t.TempDir()

	// Create multiple ecosystem files in the same directory
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(`{"name": "test", "version": "1.0.0"}`), 0644))

	// Detect packages
	packages, err := DetectPackages(tempDir)
	require.NoError(t, err)

	// Should detect as separate packages or prioritize one ecosystem
	// Current implementation should detect both
	assert.GreaterOrEqual(t, len(packages), 1, "Should detect at least one package")
}
