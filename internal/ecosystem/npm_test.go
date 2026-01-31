package ecosystem

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNPMEcosystem_ReadVersion(t *testing.T) {
	tests := []struct {
		name            string
		packageJSON     map[string]interface{}
		expectedVersion string
		wantErr         bool
	}{
		{
			name: "valid package.json",
			packageJSON: map[string]interface{}{
				"name":    "test-package",
				"version": "1.2.3",
			},
			expectedVersion: "1.2.3",
			wantErr:         false,
		},
		{
			name: "package.json with dependencies",
			packageJSON: map[string]interface{}{
				"name":    "my-app",
				"version": "2.5.10",
				"dependencies": map[string]interface{}{
					"express": "^4.18.0",
				},
			},
			expectedVersion: "2.5.10",
			wantErr:         false,
		},
		{
			name: "missing version field",
			packageJSON: map[string]interface{}{
				"name": "test-package",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			packageJSONPath := filepath.Join(tmpDir, "package.json")

			jsonData, err := json.MarshalIndent(tt.packageJSON, "", "  ")
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(packageJSONPath, jsonData, 0644))

			eco := NewNPMEcosystem(tmpDir)
			version, err := eco.ReadVersion()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedVersion, version.String())
		})
	}
}

func TestNPMEcosystem_UpdateVersion(t *testing.T) {
	tests := []struct {
		name           string
		originalJSON   map[string]interface{}
		newVersion     string
		expectedJSON   map[string]interface{}
		description    string
	}{
		{
			name: "update simple package.json",
			originalJSON: map[string]interface{}{
				"name":    "test-package",
				"version": "1.0.0",
			},
			newVersion: "2.0.0",
			expectedJSON: map[string]interface{}{
				"name":    "test-package",
				"version": "2.0.0",
			},
			description: "should update version field",
		},
		{
			name: "preserve other fields",
			originalJSON: map[string]interface{}{
				"name":    "my-app",
				"version": "1.5.0",
				"main":    "index.js",
				"scripts": map[string]interface{}{
					"start": "node index.js",
				},
				"dependencies": map[string]interface{}{
					"express": "^4.18.0",
				},
			},
			newVersion: "1.6.0",
			expectedJSON: map[string]interface{}{
				"name":    "my-app",
				"version": "1.6.0",
				"main":    "index.js",
				"scripts": map[string]interface{}{
					"start": "node index.js",
				},
				"dependencies": map[string]interface{}{
					"express": "^4.18.0",
				},
			},
			description: "should preserve all other fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			packageJSONPath := filepath.Join(tmpDir, "package.json")

			// Write original
			jsonData, err := json.MarshalIndent(tt.originalJSON, "", "  ")
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(packageJSONPath, jsonData, 0644))

			// Update version
			eco := NewNPMEcosystem(tmpDir)
			newVer, err := semver.Parse(tt.newVersion)
			require.NoError(t, err)

			err = eco.UpdateVersion(newVer)
			require.NoError(t, err)

			// Read and verify
			content, err := os.ReadFile(packageJSONPath)
			require.NoError(t, err)

			var result map[string]interface{}
			err = json.Unmarshal(content, &result)
			require.NoError(t, err)

			// Compare version
			assert.Equal(t, tt.expectedJSON["version"], result["version"], tt.description)

			// Verify other fields preserved
			for key, expectedValue := range tt.expectedJSON {
				if key != "version" {
					assert.Equal(t, expectedValue, result[key], "field %s should be preserved", key)
				}
			}
		})
	}
}

func TestNPMEcosystem_GetVersionFiles(t *testing.T) {
	tmpDir := t.TempDir()
	packageJSONPath := filepath.Join(tmpDir, "package.json")

	packageJSON := map[string]interface{}{
		"name":    "test",
		"version": "1.0.0",
	}
	jsonData, _ := json.Marshal(packageJSON)
	require.NoError(t, os.WriteFile(packageJSONPath, jsonData, 0644))

	eco := NewNPMEcosystem(tmpDir)
	files := eco.GetVersionFiles()

	assert.Contains(t, files, packageJSONPath)
}

func TestNPMEcosystem_Detect(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, dir string)
		expected bool
	}{
		{
			name: "package.json present",
			setup: func(t *testing.T, dir string) {
				packageJSON := map[string]interface{}{"name": "test"}
				jsonData, _ := json.Marshal(packageJSON)
				os.WriteFile(filepath.Join(dir, "package.json"), jsonData, 0644)
			},
			expected: true,
		},
		{
			name: "package-lock.json present",
			setup: func(t *testing.T, dir string) {
				os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte("{}"), 0644)
			},
			expected: true,
		},
		{
			name: "no npm files",
			setup: func(t *testing.T, dir string) {
				os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tt.setup(t, tmpDir)

			result := DetectNPMEcosystem(tmpDir)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNPMEcosystem_JSONFormatting(t *testing.T) {
	t.Run("preserves JSON formatting", func(t *testing.T) {
		tmpDir := t.TempDir()
		packageJSONPath := filepath.Join(tmpDir, "package.json")

		// Write with specific formatting
		originalContent := `{
  "name": "test-package",
  "version": "1.0.0",
  "description": "A test package"
}`

		require.NoError(t, os.WriteFile(packageJSONPath, []byte(originalContent), 0644))

		// Update version
		eco := NewNPMEcosystem(tmpDir)
		newVer, _ := semver.Parse("2.0.0")
		err := eco.UpdateVersion(newVer)
		require.NoError(t, err)

		// Verify formatting preserved (2-space indent)
		content, _ := os.ReadFile(packageJSONPath)
		assert.Contains(t, string(content), "  \"version\": \"2.0.0\"")
		assert.Contains(t, string(content), "  \"name\": \"test-package\"")
	})
}

func TestNPMEcosystem_ErrorHandling(t *testing.T) {
	t.Run("read version from nonexistent package.json", func(t *testing.T) {
		eco := NewNPMEcosystem("/nonexistent/path")
		_, err := eco.ReadVersion()
		assert.Error(t, err)
	})

	t.Run("update version in nonexistent package.json", func(t *testing.T) {
		eco := NewNPMEcosystem("/nonexistent/path")
		ver, _ := semver.Parse("1.0.0")
		err := eco.UpdateVersion(ver)
		assert.Error(t, err)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		packageJSONPath := filepath.Join(tmpDir, "package.json")
		require.NoError(t, os.WriteFile(packageJSONPath, []byte("invalid json"), 0644))

		eco := NewNPMEcosystem(tmpDir)
		_, err := eco.ReadVersion()
		assert.Error(t, err)
	})
}
