package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
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
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	remoteRepoDir := filepath.Join(t.TempDir(), "remote-config.git")
	repo, err := gogit.PlainInit(remoteRepoDir, false)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(remoteRepoDir, "shipyard.yaml"), []byte("packages: []\n"), 0644))
	worktree, err := repo.Worktree()
	require.NoError(t, err)
	_, err = worktree.Add("shipyard.yaml")
	require.NoError(t, err)
	_, err = worktree.Commit("add remote config", &gogit.CommitOptions{
		Author: &object.Signature{Name: "Test", Email: "test@example.com", When: time.Now()},
	})
	require.NoError(t, err)

	remote := "file://" + remoteRepoDir + "#shipyard.yaml@master"
	err = runInit(tempDir, InitOptions{Yes: true, Force: false, Remote: remote})
	require.NoError(t, err)

	configPath := filepath.Join(tempDir, ".shipyard", "shipyard.yaml")
	configContent, err := os.ReadFile(configPath)
	require.NoError(t, err)
	content := string(configContent)

	assert.Contains(t, content, "extends:")
	assert.Contains(t, content, "git: file://"+remoteRepoDir)
	assert.Contains(t, content, "path: shipyard.yaml")
	assert.Contains(t, content, "ref: master")
	assert.False(t, strings.Contains(content, "url: "+remote), "git remotes should be stored as structured git/path/ref fields")
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
