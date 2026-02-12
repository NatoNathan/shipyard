package ecosystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
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

		// Verify other fields preserved and appVersion updated
		content, _ := os.ReadFile(chartPath)
		contentStr := string(content)
		assert.Contains(t, contentStr, "apiVersion: v2")
		assert.Contains(t, contentStr, "name: my-chart")
		assert.Contains(t, contentStr, `appVersion: "2.0.0"`)
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

		// Verify: Chart.yaml returned (relative path)
		require.Len(t, files, 1)
		assert.Equal(t, "Chart.yaml", files[0])
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

// TestHelmEcosystem_UpdateVersionWithAppDependency tests the appDependency feature
func TestHelmEcosystem_UpdateVersionWithAppDependency(t *testing.T) {
	tests := []struct {
		name          string
		initialChart  string
		chartVersion  string
		appDependency string
		depVersion    string
		expectedChart string
		description   string
	}{
		{
			name: "appVersion syncs to dependency",
			initialChart: `apiVersion: v2
name: my-chart
version: 1.0.0
appVersion: "0.5.0"
`,
			chartVersion:  "1.1.0",
			appDependency: "my-app",
			depVersion:    "2.3.4",
			expectedChart: `apiVersion: v2
name: my-chart
version: 1.1.0
appVersion: "2.3.4"
`,
			description: "When appDependency is set, appVersion should sync to the dependency's version",
		},
		{
			name: "BACKWARD COMPATIBLE: no context provided",
			initialChart: `apiVersion: v2
name: my-chart
version: 1.0.0
appVersion: "1.0.0"
`,
			chartVersion: "2.0.0",
			expectedChart: `apiVersion: v2
name: my-chart
version: 2.0.0
appVersion: "2.0.0"
`,
			description: "Without context, appVersion should match chart version (backward compatible)",
		},
		{
			name: "BACKWARD COMPATIBLE: no options in config",
			initialChart: `apiVersion: v2
name: my-chart
version: 1.0.0
appVersion: "1.0.0"
`,
			chartVersion:  "2.0.0",
			appDependency: "", // Empty = no appDependency option
			expectedChart: `apiVersion: v2
name: my-chart
version: 2.0.0
appVersion: "2.0.0"
`,
			description: "Without appDependency option, appVersion should match chart version",
		},
		{
			name: "appVersion fallback when dependency not in version map",
			initialChart: `apiVersion: v2
name: my-chart
version: 1.0.0
appVersion: "1.0.0"
`,
			chartVersion:  "1.1.0",
			appDependency: "non-existent",
			depVersion:    "", // Not in version map
			expectedChart: `apiVersion: v2
name: my-chart
version: 1.1.0
appVersion: "1.1.0"
`,
			description: "When dependency not in version map, fallback to chart version",
		},
		{
			name: "preserves YAML formatting and structure",
			initialChart: `# My Chart
apiVersion: v2
name: my-chart  # Chart name
version: 1.0.0
appVersion: "0.5.0"
description: A chart
`,
			chartVersion:  "1.1.0",
			appDependency: "my-app",
			depVersion:    "2.0.0",
			expectedChart: `# My Chart
apiVersion: v2
name: my-chart  # Chart name
version: 1.1.0
appVersion: "2.0.0"
description: A chart
`,
			description: "Should preserve YAML structure and comments on separate lines",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Create temp directory with Chart.yaml
			tmpDir := t.TempDir()
			chartPath := filepath.Join(tmpDir, "Chart.yaml")
			err := os.WriteFile(chartPath, []byte(tt.initialChart), 0644)
			require.NoError(t, err)

			// Create ecosystem handler
			h := NewHelmEcosystem(tmpDir)

			// Set context if appDependency specified or if we're testing with context
			if tt.appDependency != "" || tt.name == "BACKWARD COMPATIBLE: no options in config" {
				pkg := &config.Package{
					Name:      "my-chart",
					Ecosystem: "helm",
				}

				// Only set options if appDependency is not empty
				if tt.appDependency != "" {
					pkg.Options = map[string]interface{}{
						"appDependency": tt.appDependency,
					}
				}

				allVersions := make(map[string]semver.Version)
				if tt.depVersion != "" {
					allVersions[tt.appDependency] = semver.MustParse(tt.depVersion)
				}

				ctx := &HandlerContext{
					AllVersions:   allVersions,
					PackageConfig: pkg,
				}
				h.SetContext(ctx)
			}
			// else: no context set (testing backward compatibility without context)

			// Update version
			ver := semver.MustParse(tt.chartVersion)
			err = h.UpdateVersion(ver)
			require.NoError(t, err, tt.description)

			// Read result
			result, err := os.ReadFile(chartPath)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedChart, string(result), tt.description)
		})
	}
}
