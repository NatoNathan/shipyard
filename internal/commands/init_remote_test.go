package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInitCommand_RemoteConfigHTTP tests initialization with HTTP(S) remote config
func TestInitCommand_RemoteConfigHTTP(t *testing.T) {
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Create a mock remote config file
	remoteConfigDir := t.TempDir()
	remoteConfigPath := filepath.Join(remoteConfigDir, "remote-config.yaml")
	remoteConfig := []byte(`templates:
  changelog:
    source: "builtin:corporate"
  tagName:
    inline: "{{ .Package.Name }}-v{{ .Version }}"

packages:
  - name: "shared-lib"
    path: "./lib"
    ecosystem: "go"
`)
	require.NoError(t, os.WriteFile(remoteConfigPath, remoteConfig, 0644))

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

// TestInitCommand_RemoteConfigGit tests initialization with git remote config
func TestInitCommand_RemoteConfigGit(t *testing.T) {
	// This test is more complex as it requires setting up a git repository
	// For now, we'll skip it and rely on HTTP/file-based testing
	t.Skip("Git remote config testing requires more complex setup")
}

// TestInitCommand_RemoteConfigMerge tests that remote and local configs are properly merged
func TestInitCommand_RemoteConfigMerge(t *testing.T) {
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Create a package file to be detected locally
	goModPath := filepath.Join(tempDir, "go.mod")
	require.NoError(t, os.WriteFile(goModPath, []byte("module github.com/test/local\n\ngo 1.21\n"), 0644))

	// Create a mock remote config
	remoteConfigDir := t.TempDir()
	remoteConfigPath := filepath.Join(remoteConfigDir, "remote.yaml")
	remoteConfig := []byte(`templates:
  changelog:
    source: "builtin:remote-template"
`)
	require.NoError(t, os.WriteFile(remoteConfigPath, remoteConfig, 0644))

	// Run init with remote config
	err := runInit(tempDir, InitOptions{
		Yes:    true,
		Force:  false,
		Remote: "file://" + remoteConfigPath,
	})
	require.NoError(t, err, "Init command should succeed")

	// Verify config contains both remote extends and local packages
	configPath := filepath.Join(tempDir, ".shipyard", "shipyard.yaml")
	configContent, err := os.ReadFile(configPath)
	require.NoError(t, err, "Should be able to read config file")

	assert.Contains(t, string(configContent), "extends:", "Config should contain extends section")
	assert.Contains(t, string(configContent), "local", "Config should contain locally detected package")
}

// TestInitCommand_RemoteConfigInvalidURL tests handling of invalid remote URLs
func TestInitCommand_RemoteConfigInvalidURL(t *testing.T) {
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Run init with invalid remote URL
	err := runInit(tempDir, InitOptions{
		Yes:    true,
		Force:  false,
		Remote: "https://invalid-domain-that-does-not-exist.example.com/config.yaml",
	})

	// Should succeed but may log a warning about unreachable remote
	// The init should still create local config with the extends reference
	require.NoError(t, err, "Init command should succeed even if remote is unreachable")

	// Verify config was still created
	configPath := filepath.Join(tempDir, ".shipyard", "shipyard.yaml")
	assert.FileExists(t, configPath, "Config should be created even if remote is unreachable")
}
