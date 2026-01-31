package commands

import (
	"os"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// initGitRepo initializes a git repository in the given directory
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	_, err := gogit.PlainInit(dir, false)
	require.NoError(t, err, "Failed to initialize git repository")
}

// TestInitCommand_NewProject tests initialization in a project without existing Shipyard configuration
func TestInitCommand_NewProject(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Initialize git repository
	initGitRepo(t, tempDir)

	// Run init command
	err := runInit(tempDir, InitOptions{
		Yes:    true,
		Force:  false,
		Remote: "",
	})
	require.NoError(t, err, "Init command should succeed")

	// Verify configuration file was created
	configPath := filepath.Join(tempDir, ".shipyard", "shipyard.yaml")
	assert.FileExists(t, configPath, "Configuration file should be created")

	// Verify consignments directory was created
	consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
	assert.DirExists(t, consignmentsDir, "Consignments directory should be created")

	// Verify history file was created
	historyPath := filepath.Join(tempDir, ".shipyard", "history.json")
	assert.FileExists(t, historyPath, "History file should be created")

	// Verify history file has empty array
	historyContent, err := os.ReadFile(historyPath)
	require.NoError(t, err, "Should be able to read history file")
	assert.Equal(t, "[]", string(historyContent), "History file should contain empty JSON array")
}

// TestInitCommand_ExistingConfiguration tests initialization when Shipyard is already configured
func TestInitCommand_ExistingConfiguration(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Initialize git repository
	initGitRepo(t, tempDir)

	// Create existing configuration
	shipyardDir := filepath.Join(tempDir, ".shipyard")
	require.NoError(t, os.Mkdir(shipyardDir, 0755), "Failed to create .shipyard directory")

	configPath := filepath.Join(shipyardDir, "shipyard.yaml")
	existingConfig := []byte("packages:\n  - name: existing\n    path: ./\n    ecosystem: go\n")
	require.NoError(t, os.WriteFile(configPath, existingConfig, 0644), "Failed to write existing config")

	// Run init command without force flag (should fail)
	err := runInit(tempDir, InitOptions{
		Yes:    true,
		Force:  false,
		Remote: "",
	})
	assert.Error(t, err, "Init command should fail when configuration exists without force flag")
	assert.Contains(t, err.Error(), "already initialized", "Error should indicate existing configuration")

	// Verify config was not overwritten
	configContent, err := os.ReadFile(configPath)
	require.NoError(t, err, "Should be able to read config file")
	assert.Equal(t, existingConfig, configContent, "Existing config should not be modified")
}

// TestInitCommand_ForceReinitialize tests re-initialization with --force flag
func TestInitCommand_ForceReinitialize(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Initialize git repository
	initGitRepo(t, tempDir)

	// Create existing configuration
	shipyardDir := filepath.Join(tempDir, ".shipyard")
	require.NoError(t, os.Mkdir(shipyardDir, 0755), "Failed to create .shipyard directory")

	configPath := filepath.Join(shipyardDir, "shipyard.yaml")
	existingConfig := []byte("packages:\n  - name: existing\n    path: ./\n    ecosystem: go\n")
	require.NoError(t, os.WriteFile(configPath, existingConfig, 0644), "Failed to write existing config")

	// Run init command with force flag (should succeed)
	err := runInit(tempDir, InitOptions{
		Yes:    true,
		Force:  true,
		Remote: "",
	})
	require.NoError(t, err, "Init command should succeed with force flag")

	// Verify config was recreated (will have different content)
	configContent, err := os.ReadFile(configPath)
	require.NoError(t, err, "Should be able to read config file")
	assert.NotEqual(t, existingConfig, configContent, "Config should be regenerated")
	assert.Contains(t, string(configContent), "packages:", "New config should have packages section")
}

// TestInitCommand_NotGitRepository tests initialization in a non-git directory
func TestInitCommand_NotGitRepository(t *testing.T) {
	// Create temporary directory without .git
	tempDir := t.TempDir()

	// Run init command (should fail)
	err := runInit(tempDir, InitOptions{
		Yes:    true,
		Force:  false,
		Remote: "",
	})
	assert.Error(t, err, "Init command should fail when not in git repository")
	assert.Contains(t, err.Error(), "git repository", "Error should mention git repository requirement")

	// Verify no Shipyard files were created
	shipyardDir := filepath.Join(tempDir, ".shipyard")
	assert.NoDirExists(t, shipyardDir, "Shipyard directory should not be created")
}

// TestInitCommand_PackageAutoDetection tests automatic package detection
func TestInitCommand_PackageAutoDetection(t *testing.T) {
	tests := []struct {
		name              string
		setupFiles        map[string]string
		expectedEcosystem string
		expectedPackages  []string
	}{
		{
			name: "Go module",
			setupFiles: map[string]string{
				"go.mod": "module github.com/test/project\n\ngo 1.21\n",
			},
			expectedEcosystem: "go",
			expectedPackages:  []string{"project"},
		},
		{
			name: "NPM package",
			setupFiles: map[string]string{
				"package.json": `{"name": "test-package", "version": "1.0.0"}`,
			},
			expectedEcosystem: "npm",
			expectedPackages:  []string{"test-package"},
		},
		{
			name: "Python package",
			setupFiles: map[string]string{
				"pyproject.toml": "[project]\nname = \"test-package\"\nversion = \"1.0.0\"\n",
			},
			expectedEcosystem: "python",
			expectedPackages:  []string{"test-package"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir := t.TempDir()

			// Initialize git repository
			// Initialize git repository
			initGitRepo(t, tempDir)

			// Create test files
			for filename, content := range tt.setupFiles {
				filePath := filepath.Join(tempDir, filename)
				require.NoError(t, os.WriteFile(filePath, []byte(content), 0644), "Failed to write test file")
			}

			// Run init command
			err := runInit(tempDir, InitOptions{
				Yes:    true,
				Force:  false,
				Remote: "",
			})
			require.NoError(t, err, "Init command should succeed")

			// Verify configuration contains detected package
			configPath := filepath.Join(tempDir, ".shipyard", "shipyard.yaml")
			configContent, err := os.ReadFile(configPath)
			require.NoError(t, err, "Should be able to read config file")

			// Check ecosystem detection
			assert.Contains(t, string(configContent), "ecosystem: "+tt.expectedEcosystem, "Config should contain detected ecosystem")

			// Check package name detection (for some ecosystems)
			if len(tt.expectedPackages) > 0 {
				assert.Contains(t, string(configContent), "name:", "Config should contain package name")
			}
		})
	}
}

// TestInitCommand_MonorepoDetection tests detection of multiple packages in monorepo
func TestInitCommand_MonorepoDetection(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Initialize git repository
	initGitRepo(t, tempDir)

	// Create monorepo structure with multiple packages
	packages := map[string]string{
		"packages/core/go.mod":           "module github.com/test/core\n\ngo 1.21\n",
		"packages/api/go.mod":            "module github.com/test/api\n\ngo 1.21\n",
		"clients/web/package.json":       `{"name": "@test/web", "version": "1.0.0"}`,
		"services/worker/pyproject.toml": "[project]\nname = \"worker\"\nversion = \"1.0.0\"\n",
	}

	for filePath, content := range packages {
		fullPath := filepath.Join(tempDir, filePath)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755), "Failed to create package directory")
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644), "Failed to write package file")
	}

	// Run init command
	err := runInit(tempDir, InitOptions{
		Yes:    true,
		Force:  false,
		Remote: "",
	})
	require.NoError(t, err, "Init command should succeed")

	// Verify configuration contains all detected packages
	configPath := filepath.Join(tempDir, ".shipyard", "shipyard.yaml")
	configContent, err := os.ReadFile(configPath)
	require.NoError(t, err, "Should be able to read config file")

	// Check that multiple packages were detected
	assert.Contains(t, string(configContent), "packages:", "Config should have packages section")

	// Count package entries (each package should have a "- name:" entry)
	packageCount := 0
	for i := 0; i < len(configContent)-7; i++ {
		if string(configContent[i:i+7]) == "- name:" {
			packageCount++
		}
	}
	assert.GreaterOrEqual(t, packageCount, 4, "Config should contain at least 4 detected packages")
}

// TestInitCommand_RemoteConfig tests initialization with remote configuration
func TestInitCommand_RemoteConfig(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Initialize git repository
	initGitRepo(t, tempDir)

	// Create a mock remote config file
	remoteConfigDir := t.TempDir()
	remoteConfigPath := filepath.Join(remoteConfigDir, "remote-config.yaml")
	remoteConfig := []byte(`templates:
  changelog:
    source: "builtin:corporate"
  tagName:
    inline: "{{ .Package.Name }}-v{{ .Version }}"
`)
	require.NoError(t, os.WriteFile(remoteConfigPath, remoteConfig, 0644), "Failed to write remote config")

	// Run init command with remote config (using file:// URL for testing)
	err := runInit(tempDir, InitOptions{
		Yes:    true,
		Force:  false,
		Remote: "file://" + remoteConfigPath,
	})
	require.NoError(t, err, "Init command should succeed with remote config")

	// Verify configuration includes remote config reference
	configPath := filepath.Join(tempDir, ".shipyard", "shipyard.yaml")
	configContent, err := os.ReadFile(configPath)
	require.NoError(t, err, "Should be able to read config file")

	assert.Contains(t, string(configContent), "extends:", "Config should contain extends section")
	assert.Contains(t, string(configContent), remoteConfigPath, "Config should reference remote config URL")
}

