package ecosystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHelmEcosystem_ReadVersion tests reading version from Chart.yaml
func TestHelmEcosystem_ReadVersion(t *testing.T) {
	t.Run("reads version from Chart.yaml", func(t *testing.T) {
		// Setup: Create Chart.yaml
		tempDir := t.TempDir()
		chartPath := filepath.Join(tempDir, "Chart.yaml")
		chartContent := `apiVersion: v2
name: my-chart
description: A Helm chart for Kubernetes
type: application
version: 1.2.3
appVersion: "1.0.0"
`
		require.NoError(t, os.WriteFile(chartPath, []byte(chartContent), 0644))

		// Test: Read version
		helm := NewHelmEcosystem(tempDir)
		version, err := helm.ReadVersion()

		// Verify: Version read correctly
		require.NoError(t, err)
		assert.Equal(t, "1.2.3", version.String())
	})

	t.Run("returns error for missing Chart.yaml", func(t *testing.T) {
		// Setup: Empty directory
		tempDir := t.TempDir()

		// Test: Read version
		helm := NewHelmEcosystem(tempDir)
		_, err := helm.ReadVersion()

		// Verify: Error returned
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Chart.yaml")
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		// Setup: Create invalid Chart.yaml
		tempDir := t.TempDir()
		chartPath := filepath.Join(tempDir, "Chart.yaml")
		require.NoError(t, os.WriteFile(chartPath, []byte("invalid: yaml: syntax"), 0644))

		// Test: Read version
		helm := NewHelmEcosystem(tempDir)
		_, err := helm.ReadVersion()

		// Verify: Error returned
		assert.Error(t, err)
	})

	t.Run("returns error for missing version field", func(t *testing.T) {
		// Setup: Create Chart.yaml without version
		tempDir := t.TempDir()
		chartPath := filepath.Join(tempDir, "Chart.yaml")
		chartContent := `apiVersion: v2
name: my-chart
description: A Helm chart
`
		require.NoError(t, os.WriteFile(chartPath, []byte(chartContent), 0644))

		// Test: Read version
		helm := NewHelmEcosystem(tempDir)
		_, err := helm.ReadVersion()

		// Verify: Error returned
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "version")
	})
}

// TestHelmEcosystem_UpdateVersion tests updating version in Chart.yaml
func TestHelmEcosystem_UpdateVersion(t *testing.T) {
	t.Run("updates version in Chart.yaml", func(t *testing.T) {
		// Setup: Create Chart.yaml
		tempDir := t.TempDir()
		chartPath := filepath.Join(tempDir, "Chart.yaml")
		chartContent := `apiVersion: v2
name: my-chart
description: A Helm chart for Kubernetes
type: application
version: 1.2.3
appVersion: "1.0.0"
`
		require.NoError(t, os.WriteFile(chartPath, []byte(chartContent), 0644))

		// Test: Update version
		helm := NewHelmEcosystem(tempDir)
		newVersion := semver.MustParse("2.0.0")
		err := helm.UpdateVersion(newVersion)

		// Verify: Version updated
		require.NoError(t, err)

		// Read back and verify
		updatedVersion, err := helm.ReadVersion()
		require.NoError(t, err)
		assert.Equal(t, "2.0.0", updatedVersion.String())

		// Verify other fields preserved
		content, _ := os.ReadFile(chartPath)
		contentStr := string(content)
		assert.Contains(t, contentStr, "apiVersion: v2")
		assert.Contains(t, contentStr, "name: my-chart")
		assert.Contains(t, contentStr, "appVersion")
		assert.Contains(t, contentStr, "1.0.0")
	})

	t.Run("returns error for missing Chart.yaml", func(t *testing.T) {
		// Setup: Empty directory
		tempDir := t.TempDir()

		// Test: Update version
		helm := NewHelmEcosystem(tempDir)
		newVersion := semver.MustParse("2.0.0")
		err := helm.UpdateVersion(newVersion)

		// Verify: Error returned
		assert.Error(t, err)
	})
}

// TestHelmEcosystem_GetVersionFiles tests getting version file paths
func TestHelmEcosystem_GetVersionFiles(t *testing.T) {
	t.Run("returns Chart.yaml path", func(t *testing.T) {
		// Setup: Create Chart.yaml
		tempDir := t.TempDir()
		chartPath := filepath.Join(tempDir, "Chart.yaml")
		require.NoError(t, os.WriteFile(chartPath, []byte("version: 1.0.0\n"), 0644))

		// Test: Get version files
		helm := NewHelmEcosystem(tempDir)
		files := helm.GetVersionFiles()

		// Verify: Chart.yaml returned
		require.Len(t, files, 1)
		assert.Equal(t, chartPath, files[0])
	})

	t.Run("returns empty for missing Chart.yaml", func(t *testing.T) {
		// Setup: Empty directory
		tempDir := t.TempDir()

		// Test: Get version files
		helm := NewHelmEcosystem(tempDir)
		files := helm.GetVersionFiles()

		// Verify: Empty slice returned
		assert.Len(t, files, 0)
	})
}

// TestDetectHelmEcosystem tests Helm chart detection
func TestDetectHelmEcosystem(t *testing.T) {
	t.Run("detects Chart.yaml", func(t *testing.T) {
		// Setup: Create Chart.yaml
		tempDir := t.TempDir()
		chartPath := filepath.Join(tempDir, "Chart.yaml")
		require.NoError(t, os.WriteFile(chartPath, []byte("version: 1.0.0\n"), 0644))

		// Test: Detect Helm
		detected := DetectHelmEcosystem(tempDir)

		// Verify: Detected
		assert.True(t, detected)
	})

	t.Run("detects charts directory", func(t *testing.T) {
		// Setup: Create charts directory
		tempDir := t.TempDir()
		chartsDir := filepath.Join(tempDir, "charts")
		require.NoError(t, os.MkdirAll(chartsDir, 0755))

		// Test: Detect Helm
		detected := DetectHelmEcosystem(tempDir)

		// Verify: Detected
		assert.True(t, detected)
	})

	t.Run("returns false for no Helm files", func(t *testing.T) {
		// Setup: Empty directory
		tempDir := t.TempDir()

		// Test: Detect Helm
		detected := DetectHelmEcosystem(tempDir)

		// Verify: Not detected
		assert.False(t, detected)
	})
}
