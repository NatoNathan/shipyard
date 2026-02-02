package ecosystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoEcosystem_ReadVersion_VersionGo(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectedVersion string
		wantErr         bool
	}{
		{
			name: "simple version const",
			content: `package main

const Version = "1.2.3"
`,
			expectedVersion: "1.2.3",
			wantErr:         false,
		},
		{
			name: "version with package comment",
			content: `// Package provides CLI functionality
package myapp

// Version is the application version
const Version = "2.5.10"
`,
			expectedVersion: "2.5.10",
			wantErr:         false,
		},
		{
			name: "version with var instead of const",
			content: `package main

var Version = "3.0.0"
`,
			expectedVersion: "3.0.0",
			wantErr:         false,
		},
		{
			name: "version with single quotes",
			content: `package main

const Version = '0.1.0'
`,
			expectedVersion: "0.1.0",
			wantErr:         false,
		},
		{
			name: "no version found",
			content: `package main

const AppName = "myapp"
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			versionFile := filepath.Join(tmpDir, "version.go")
			require.NoError(t, os.WriteFile(versionFile, []byte(tt.content), 0644))

			eco := NewGoEcosystem(tmpDir)
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

func TestGoEcosystem_ReadVersion_GoMod(t *testing.T) {
	tests := []struct {
		name            string
		versionGoExists bool
		goModContent    string
		expectedVersion string
		wantErr         bool
	}{
		{
			name:            "version from go.mod comment",
			versionGoExists: false,
			goModContent: `module github.com/example/project

go 1.21

// version: 1.5.0
`,
			expectedVersion: "1.5.0",
			wantErr:         false,
		},
		{
			name:            "prefer version.go over go.mod",
			versionGoExists: true,
			goModContent: `module github.com/example/project

// version: 2.0.0
`,
			expectedVersion: "1.0.0", // From version.go
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if tt.versionGoExists {
				versionGo := `package main

const Version = "1.0.0"
`
				require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "version.go"), []byte(versionGo), 0644))
			}

			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(tt.goModContent), 0644))

			eco := NewGoEcosystem(tmpDir)
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

func TestGoEcosystem_UpdateVersion_VersionGo(t *testing.T) {
	tests := []struct {
		name           string
		originalContent string
		newVersion     string
		expectedContent string
		description    string
	}{
		{
			name: "update simple const",
			originalContent: `package main

const Version = "1.0.0"
`,
			newVersion: "2.0.0",
			expectedContent: `package main

const Version = "2.0.0"
`,
			description: "should update version string",
		},
		{
			name: "update with comments",
			originalContent: `package main

// Version is the application version
const Version = "1.5.0"

func main() {
	// some code
}
`,
			newVersion: "1.6.0",
			expectedContent: `package main

// Version is the application version
const Version = "1.6.0"

func main() {
	// some code
}
`,
			description: "should preserve surrounding code",
		},
		{
			name: "update var declaration",
			originalContent: `package main

var Version = "3.0.0"
`,
			newVersion: "3.1.0",
			expectedContent: `package main

var Version = "3.1.0"
`,
			description: "should update var declaration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			versionFile := filepath.Join(tmpDir, "version.go")
			require.NoError(t, os.WriteFile(versionFile, []byte(tt.originalContent), 0644))

			eco := NewGoEcosystem(tmpDir)
			newVer, err := semver.Parse(tt.newVersion)
			require.NoError(t, err)

			err = eco.UpdateVersion(newVer)
			require.NoError(t, err)

			// Verify file was updated
			content, err := os.ReadFile(versionFile)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedContent, string(content), tt.description)
		})
	}
}

func TestGoEcosystem_UpdateVersion_GoMod(t *testing.T) {
	t.Run("update go.mod comment", func(t *testing.T) {
		tmpDir := t.TempDir()

		originalContent := `module github.com/example/project

go 1.21

// version: 1.0.0

require (
	github.com/spf13/cobra v1.7.0
)
`

		expectedContent := `module github.com/example/project

go 1.21

// version: 2.0.0

require (
	github.com/spf13/cobra v1.7.0
)
`

		goModFile := filepath.Join(tmpDir, "go.mod")
		require.NoError(t, os.WriteFile(goModFile, []byte(originalContent), 0644))

		eco := NewGoEcosystem(tmpDir)
		newVer, err := semver.Parse("2.0.0")
		require.NoError(t, err)

		err = eco.UpdateVersion(newVer)
		require.NoError(t, err)

		// Verify file was updated
		content, err := os.ReadFile(goModFile)
		require.NoError(t, err)
		assert.Equal(t, expectedContent, string(content))
	})
}

func TestGoEcosystem_GetVersionFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create both version.go and go.mod
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "version.go"), []byte("package main\nconst Version = \"1.0.0\""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n// version: 1.0.0"), 0644))

	eco := NewGoEcosystem(tmpDir)
	files := eco.GetVersionFiles()

	// Should return at least version.go (relative path)
	assert.Contains(t, files, "version.go")
}

func TestGoEcosystem_Detect(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, dir string)
		expected bool
	}{
		{
			name: "go.mod present",
			setup: func(t *testing.T, dir string) {
				os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)
			},
			expected: true,
		},
		{
			name: "version.go present",
			setup: func(t *testing.T, dir string) {
				os.WriteFile(filepath.Join(dir, "version.go"), []byte("package main"), 0644)
			},
			expected: true,
		},
		{
			name: "no go files",
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

			result := DetectGoEcosystem(tmpDir)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGoEcosystem_ErrorHandling(t *testing.T) {
	t.Run("read version from nonexistent directory", func(t *testing.T) {
		eco := NewGoEcosystem("/nonexistent/path")
		_, err := eco.ReadVersion()
		assert.Error(t, err)
	})

	t.Run("update version in nonexistent directory", func(t *testing.T) {
		eco := NewGoEcosystem("/nonexistent/path")
		ver, _ := semver.Parse("1.0.0")
		err := eco.UpdateVersion(ver)
		assert.Error(t, err)
	})
}
// TestGoEcosystem_TagOnly tests tag-only mode for Go projects
func TestGoEcosystem_TagOnly(t *testing.T) {
	t.Run("tag-only mode skips file updates", func(t *testing.T) {
		// Setup: Create version.go with original version
		tempDir := t.TempDir()
		versionPath := filepath.Join(tempDir, "version.go")
		originalContent := `package main

const Version = "1.0.0"
`
		require.NoError(t, os.WriteFile(versionPath, []byte(originalContent), 0644))

		// Test: Update with tag-only mode
		eco := NewGoEcosystemWithOptions(tempDir, &GoEcosystemOptions{TagOnly: true})
		newVersion := semver.MustParse("2.0.0")
		err := eco.UpdateVersion(newVersion)

		// Verify: No error and file unchanged
		require.NoError(t, err)
		
		content, err := os.ReadFile(versionPath)
		require.NoError(t, err)
		assert.Equal(t, originalContent, string(content), "File should not be modified in tag-only mode")
	})

	t.Run("tag-only mode reads version from files", func(t *testing.T) {
		// Setup: Create version.go
		tempDir := t.TempDir()
		versionPath := filepath.Join(tempDir, "version.go")
		content := `package main

const Version = "1.5.0"
`
		require.NoError(t, os.WriteFile(versionPath, []byte(content), 0644))

		// Test: Read version with tag-only mode
		eco := NewGoEcosystemWithOptions(tempDir, &GoEcosystemOptions{TagOnly: true})
		version, err := eco.ReadVersion()

		// Verify: Version read correctly
		require.NoError(t, err)
		assert.Equal(t, "1.5.0", version.String())
	})

	t.Run("tag-only returns empty version files", func(t *testing.T) {
		// Setup: Create version.go
		tempDir := t.TempDir()
		versionPath := filepath.Join(tempDir, "version.go")
		require.NoError(t, os.WriteFile(versionPath, []byte(`package main
const Version = "1.0.0"`), 0644))

		// Test: Get version files with tag-only mode
		eco := NewGoEcosystemWithOptions(tempDir, &GoEcosystemOptions{TagOnly: true})
		files := eco.GetVersionFiles()

		// Verify: Empty list (no files to update)
		assert.Len(t, files, 0, "Tag-only mode should return empty version files list")
	})

	t.Run("default mode is not tag-only", func(t *testing.T) {
		// Setup: Create version.go
		tempDir := t.TempDir()
		versionPath := filepath.Join(tempDir, "version.go")
		originalContent := `package main

const Version = "1.0.0"
`
		require.NoError(t, os.WriteFile(versionPath, []byte(originalContent), 0644))

		// Test: Update with default (non tag-only) mode
		eco := NewGoEcosystem(tempDir)
		newVersion := semver.MustParse("2.0.0")
		err := eco.UpdateVersion(newVersion)

		// Verify: File updated
		require.NoError(t, err)
		
		content, err := os.ReadFile(versionPath)
		require.NoError(t, err)
		assert.NotEqual(t, originalContent, string(content), "File should be modified in normal mode")
		assert.Contains(t, string(content), "2.0.0")
	})
}
