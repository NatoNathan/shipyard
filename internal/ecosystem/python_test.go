package ecosystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPythonEcosystem_ReadVersion_PyprojectToml(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectedVersion string
		wantErr         bool
	}{
		{
			name: "valid pyproject.toml",
			content: `[tool.poetry]
name = "myproject"
version = "1.2.3"
description = "A test project"
`,
			expectedVersion: "1.2.3",
			wantErr:         false,
		},
		{
			name: "project section format",
			content: `[project]
name = "myproject"
version = "2.5.10"
`,
			expectedVersion: "2.5.10",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			pyprojectPath := filepath.Join(tmpDir, "pyproject.toml")
			require.NoError(t, os.WriteFile(pyprojectPath, []byte(tt.content), 0644))

			eco := NewPythonEcosystem(tmpDir)
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

func TestPythonEcosystem_ReadVersion_VersionPy(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectedVersion string
		wantErr         bool
	}{
		{
			name: "__version__ string",
			content: `# Module version
__version__ = "1.2.3"
`,
			expectedVersion: "1.2.3",
			wantErr:         false,
		},
		{
			name: "__version__ with single quotes",
			content: `__version__ = '2.0.0'
`,
			expectedVersion: "2.0.0",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			versionPyPath := filepath.Join(tmpDir, "__version__.py")
			require.NoError(t, os.WriteFile(versionPyPath, []byte(tt.content), 0644))

			eco := NewPythonEcosystem(tmpDir)
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

func TestPythonEcosystem_UpdateVersion(t *testing.T) {
	t.Run("update pyproject.toml", func(t *testing.T) {
		tmpDir := t.TempDir()
		pyprojectPath := filepath.Join(tmpDir, "pyproject.toml")

		original := `[tool.poetry]
name = "myproject"
version = "1.0.0"
description = "A test project"
`

		expected := `[tool.poetry]
name = "myproject"
version = "2.0.0"
description = "A test project"
`

		require.NoError(t, os.WriteFile(pyprojectPath, []byte(original), 0644))

		eco := NewPythonEcosystem(tmpDir)
		newVer, _ := semver.Parse("2.0.0")
		err := eco.UpdateVersion(newVer)
		require.NoError(t, err)

		content, _ := os.ReadFile(pyprojectPath)
		assert.Equal(t, expected, string(content))
	})

	t.Run("update __version__.py", func(t *testing.T) {
		tmpDir := t.TempDir()
		versionPyPath := filepath.Join(tmpDir, "__version__.py")

		original := `__version__ = "1.0.0"
`

		expected := `__version__ = "2.0.0"
`

		require.NoError(t, os.WriteFile(versionPyPath, []byte(original), 0644))

		eco := NewPythonEcosystem(tmpDir)
		newVer, _ := semver.Parse("2.0.0")
		err := eco.UpdateVersion(newVer)
		require.NoError(t, err)

		content, _ := os.ReadFile(versionPyPath)
		assert.Equal(t, expected, string(content))
	})
}

func TestDetectPythonEcosystem(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, dir string)
		expected bool
	}{
		{
			name: "pyproject.toml present",
			setup: func(t *testing.T, dir string) {
				os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[project]"), 0644)
			},
			expected: true,
		},
		{
			name: "__version__.py present",
			setup: func(t *testing.T, dir string) {
				os.WriteFile(filepath.Join(dir, "__version__.py"), []byte("__version__ = '1.0.0'"), 0644)
			},
			expected: true,
		},
		{
			name: "setup.py present",
			setup: func(t *testing.T, dir string) {
				os.WriteFile(filepath.Join(dir, "setup.py"), []byte("from setuptools import setup"), 0644)
			},
			expected: true,
		},
		{
			name: "no python files",
			setup: func(t *testing.T, dir string) {
				os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644)
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tt.setup(t, tmpDir)

			result := DetectPythonEcosystem(tmpDir)
			assert.Equal(t, tt.expected, result)
		})
	}
}
