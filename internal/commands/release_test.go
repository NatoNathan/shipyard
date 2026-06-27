package commands

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/history"
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReleaseCommand_MissingGitHubConfig(t *testing.T) {
	// Setup: Create a temp directory with shipyard initialized but no GitHub config
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Create config without GitHub settings
	configPath := filepath.Join(tempDir, ".shipyard", "shipyard.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0755))

	cfg := &config.Config{
		Packages: []config.Package{
			{Name: "core", Path: ".", Ecosystem: config.EcosystemGo},
		},
	}
	require.NoError(t, config.WriteConfig(cfg, configPath))

	// Create history with a release
	historyDir := filepath.Join(tempDir, ".shipyard")
	require.NoError(t, os.MkdirAll(historyDir, 0755))
	historyPath := filepath.Join(historyDir, "history.json")
	// Create empty history file first (required by AppendToHistory)
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))
	entries := []history.Entry{
		{
			Version: "1.0.0",
			Package: "core",
			Tag:     "v1.0.0",
		},
	}
	require.NoError(t, history.AppendToHistory(historyPath, entries))

	// Change to test directory
	cleanup := changeToDir(t, tempDir)
	defer cleanup()

	// Run release command - should fail due to missing GitHub config
	opts := &ReleaseOptions{
		Package: "core",
	}
	err := runRelease(opts)

	// Verify error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GitHub not configured")
}

func TestReleaseCommand_MissingToken(t *testing.T) {
	// Setup: Create a temp directory with shipyard initialized and GitHub config
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Create config with GitHub settings
	configPath := filepath.Join(tempDir, ".shipyard", "shipyard.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0755))

	cfg := &config.Config{
		Packages: []config.Package{
			{Name: "core", Path: ".", Ecosystem: config.EcosystemGo},
		},
		GitHub: config.GitHubConfig{
			Owner: "testowner",
			Repo:  "testrepo",
		},
	}
	require.NoError(t, config.WriteConfig(cfg, configPath))

	// Create history with a release
	historyDir := filepath.Join(tempDir, ".shipyard")
	require.NoError(t, os.MkdirAll(historyDir, 0755))
	historyPath := filepath.Join(historyDir, "history.json")
	// Create empty history file first (required by AppendToHistory)
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))
	entries := []history.Entry{
		{
			Version: "1.0.0",
			Package: "core",
			Tag:     "v1.0.0",
		},
	}
	require.NoError(t, history.AppendToHistory(historyPath, entries))

	// Change to test directory
	cleanup := changeToDir(t, tempDir)
	defer cleanup()

	// Ensure GITHUB_TOKEN is not set
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", originalToken)
	os.Unsetenv("GITHUB_TOKEN")

	// Run release command - should fail due to missing token
	opts := &ReleaseOptions{
		Package: "core",
	}
	err := runRelease(opts)

	// Verify error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GITHUB_TOKEN")
}

func TestReleaseCommand_NoHistory(t *testing.T) {
	// Setup: Create a temp directory with shipyard initialized
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Create config with GitHub settings
	configPath := filepath.Join(tempDir, ".shipyard", "shipyard.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0755))

	cfg := &config.Config{
		Packages: []config.Package{
			{Name: "core", Path: ".", Ecosystem: config.EcosystemGo},
		},
		GitHub: config.GitHubConfig{
			Owner: "testowner",
			Repo:  "testrepo",
		},
	}
	require.NoError(t, config.WriteConfig(cfg, configPath))

	// Create empty history file
	historyPath := filepath.Join(tempDir, ".shipyard", "history.json")
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))

	// Change to test directory
	cleanup := changeToDir(t, tempDir)
	defer cleanup()

	// Set fake token
	os.Setenv("GITHUB_TOKEN", "fake-token-for-test")
	defer os.Unsetenv("GITHUB_TOKEN")

	// Run release command - should fail due to no history
	opts := &ReleaseOptions{
		Package: "core",
	}
	err := runRelease(opts)

	// Verify error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no releases found")
}

func TestReleaseCommand_MultiPackageRequiresFlag(t *testing.T) {
	// Setup: Create a temp directory with multi-package repo
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Create config with multiple packages
	configPath := filepath.Join(tempDir, ".shipyard", "shipyard.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0755))

	cfg := &config.Config{
		Packages: []config.Package{
			{Name: "core", Path: ".", Ecosystem: config.EcosystemGo},
			{Name: "api", Path: "./api", Ecosystem: config.EcosystemGo},
		},
		GitHub: config.GitHubConfig{
			Owner: "testowner",
			Repo:  "testrepo",
		},
	}
	require.NoError(t, config.WriteConfig(cfg, configPath))

	// Create history with releases for both packages
	historyDir := filepath.Join(tempDir, ".shipyard")
	require.NoError(t, os.MkdirAll(historyDir, 0755))
	historyPath := filepath.Join(historyDir, "history.json")
	// Create empty history file first (required by AppendToHistory)
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))
	entries := []history.Entry{
		{Version: "1.0.0", Package: "core", Tag: "core/v1.0.0"},
		{Version: "1.0.0", Package: "api", Tag: "api/v1.0.0"},
	}
	require.NoError(t, history.AppendToHistory(historyPath, entries))

	// Change to test directory
	cleanup := changeToDir(t, tempDir)
	defer cleanup()

	// Set fake token
	os.Setenv("GITHUB_TOKEN", "fake-token-for-test")
	defer os.Unsetenv("GITHUB_TOKEN")

	// Run release command without --package flag
	opts := &ReleaseOptions{
		Package: "", // Empty - should require flag
	}
	err := runRelease(opts)

	// Verify error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--package is required")
}

func TestReleaseCommand_JSONOutput(t *testing.T) {
	tempDir := setupReleaseCommandProject(t, []history.Entry{{Version: "1.2.3", Package: "core", Tag: "v1.2.3"}})
	cleanup := changeToDir(t, tempDir)
	defer cleanup()
	t.Setenv("GITHUB_TOKEN", "fake-token-for-test")

	fake := &fakeReleasePublisher{}
	withFakeReleasePublisher(t, fake)

	output := captureStdout(t, func() {
		err := runRelease(&ReleaseOptions{Package: "core", JSON: true})
		require.NoError(t, err)
	})

	require.Len(t, fake.calls, 1)
	assert.Contains(t, output, `"success": true`)
	assert.Contains(t, output, `"package": "core"`)
	assert.Contains(t, output, `"version": "1.2.3"`)
	assert.Contains(t, output, `"tag": "v1.2.3"`)
}

func TestReleaseCommand_QuietMode(t *testing.T) {
	tempDir := setupReleaseCommandProject(t, []history.Entry{{Version: "1.2.3", Package: "core", Tag: "v1.2.3"}})
	cleanup := changeToDir(t, tempDir)
	defer cleanup()
	t.Setenv("GITHUB_TOKEN", "fake-token-for-test")

	fake := &fakeReleasePublisher{}
	withFakeReleasePublisher(t, fake)

	output := captureStdout(t, func() {
		err := runRelease(&ReleaseOptions{Package: "core", Quiet: true})
		require.NoError(t, err)
	})

	require.Len(t, fake.calls, 1)
	assert.Empty(t, strings.TrimSpace(output))
}

func TestReleaseCommand_DraftFlag(t *testing.T) {
	tempDir := setupReleaseCommandProject(t, []history.Entry{{Version: "1.2.3", Package: "core", Tag: "v1.2.3"}})
	cleanup := changeToDir(t, tempDir)
	defer cleanup()
	t.Setenv("GITHUB_TOKEN", "fake-token-for-test")

	fake := &fakeReleasePublisher{}
	withFakeReleasePublisher(t, fake)

	err := runRelease(&ReleaseOptions{Package: "core", Draft: true, Quiet: true})
	require.NoError(t, err)
	require.Len(t, fake.calls, 1)
	assert.True(t, fake.calls[0].draft)
	assert.False(t, fake.calls[0].prerelease)
}

func TestReleaseCommand_PrereleaseFlag(t *testing.T) {
	tempDir := setupReleaseCommandProject(t, []history.Entry{{Version: "1.2.3", Package: "core", Tag: "v1.2.3-beta.1"}})
	cleanup := changeToDir(t, tempDir)
	defer cleanup()
	t.Setenv("GITHUB_TOKEN", "fake-token-for-test")

	fake := &fakeReleasePublisher{}
	withFakeReleasePublisher(t, fake)

	err := runRelease(&ReleaseOptions{Package: "core", Prerelease: true, Quiet: true})
	require.NoError(t, err)
	require.Len(t, fake.calls, 1)
	assert.False(t, fake.calls[0].draft)
	assert.True(t, fake.calls[0].prerelease)
}

func TestReleaseCommand_SpecificTag(t *testing.T) {
	// Setup: Create a temp directory with multiple releases
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Create config with GitHub settings
	configPath := filepath.Join(tempDir, ".shipyard", "shipyard.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0755))

	cfg := &config.Config{
		Packages: []config.Package{
			{Name: "core", Path: ".", Ecosystem: config.EcosystemGo},
		},
		GitHub: config.GitHubConfig{
			Owner: "testowner",
			Repo:  "testrepo",
		},
	}
	require.NoError(t, config.WriteConfig(cfg, configPath))

	// Create history with multiple releases
	historyDir := filepath.Join(tempDir, ".shipyard")
	require.NoError(t, os.MkdirAll(historyDir, 0755))
	historyPath := filepath.Join(historyDir, "history.json")
	// Create empty history file first (required by AppendToHistory)
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))
	entries := []history.Entry{
		{Version: "1.0.0", Package: "core", Tag: "v1.0.0"},
		{Version: "1.1.0", Package: "core", Tag: "v1.1.0"},
		{Version: "1.2.0", Package: "core", Tag: "v1.2.0"},
	}
	require.NoError(t, history.AppendToHistory(historyPath, entries))

	// Change to test directory
	cleanup := changeToDir(t, tempDir)
	defer cleanup()

	// Set fake token
	os.Setenv("GITHUB_TOKEN", "fake-token-for-test")
	defer os.Unsetenv("GITHUB_TOKEN")

	// Run release command with specific tag
	opts := &ReleaseOptions{
		Package: "core",
		Tag:     "v1.0.0",
	}

	// This will fail at GitHub publishing, but we can verify it selected the right tag
	// by checking the error doesn't say "tag not found"
	err := runRelease(opts)

	// Should fail at GitHub API call, not tag selection
	if err != nil {
		assert.NotContains(t, err.Error(), "tag v1.0.0 not found in history")
	}
}

func TestReleaseCommand_InvalidTag(t *testing.T) {
	// Setup: Create a temp directory with releases
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Create config with GitHub settings
	configPath := filepath.Join(tempDir, ".shipyard", "shipyard.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0755))

	cfg := &config.Config{
		Packages: []config.Package{
			{Name: "core", Path: ".", Ecosystem: config.EcosystemGo},
		},
		GitHub: config.GitHubConfig{
			Owner: "testowner",
			Repo:  "testrepo",
		},
	}
	require.NoError(t, config.WriteConfig(cfg, configPath))

	// Create history with releases
	historyDir := filepath.Join(tempDir, ".shipyard")
	require.NoError(t, os.MkdirAll(historyDir, 0755))
	historyPath := filepath.Join(historyDir, "history.json")
	// Create empty history file first (required by AppendToHistory)
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))
	entries := []history.Entry{
		{Version: "1.0.0", Package: "core", Tag: "v1.0.0"},
	}
	require.NoError(t, history.AppendToHistory(historyPath, entries))

	// Change to test directory
	cleanup := changeToDir(t, tempDir)
	defer cleanup()

	// Set fake token
	os.Setenv("GITHUB_TOKEN", "fake-token-for-test")
	defer os.Unsetenv("GITHUB_TOKEN")

	// Run release command with non-existent tag
	opts := &ReleaseOptions{
		Package: "core",
		Tag:     "v9.9.9", // Doesn't exist
	}
	err := runRelease(opts)

	// Verify error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tag v9.9.9 not found in history")
}

type fakeReleaseCall struct {
	packageName string
	version     semver.Version
	tagName     string
	notes       string
	draft       bool
	prerelease  bool
}

type fakeReleasePublisher struct {
	calls []fakeReleaseCall
	err   error
}

func (p *fakeReleasePublisher) PublishRelease(ctx context.Context, packageName string, version semver.Version, tagName string, releaseNotes string, draft bool, prerelease bool) error {
	p.calls = append(p.calls, fakeReleaseCall{
		packageName: packageName,
		version:     version,
		tagName:     tagName,
		notes:       releaseNotes,
		draft:       draft,
		prerelease:  prerelease,
	})
	return p.err
}

func withFakeReleasePublisher(t *testing.T, publisher *fakeReleasePublisher) {
	t.Helper()
	original := newReleasePublisher
	newReleasePublisher = func(repoPath string, cfg *config.Config) ReleasePublisher {
		return publisher
	}
	t.Cleanup(func() {
		newReleasePublisher = original
	})
}

func setupReleaseCommandProject(t *testing.T, entries []history.Entry) string {
	t.Helper()

	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	shipyardDir := filepath.Join(tempDir, ".shipyard")
	require.NoError(t, os.MkdirAll(shipyardDir, 0755))

	cfg := &config.Config{
		Packages: []config.Package{
			{Name: "core", Path: ".", Ecosystem: config.EcosystemGo},
		},
		GitHub: config.GitHubConfig{
			Owner: "testowner",
			Repo:  "testrepo",
		},
	}
	require.NoError(t, config.WriteConfig(cfg, filepath.Join(shipyardDir, "shipyard.yaml")))

	historyPath := filepath.Join(shipyardDir, "history.json")
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))
	require.NoError(t, history.AppendToHistory(historyPath, entries))

	return tempDir
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = writer
	defer func() {
		os.Stdout = original
		_ = reader.Close()
		_ = writer.Close()
	}()

	fn()

	require.NoError(t, writer.Close())

	output, err := io.ReadAll(reader)
	require.NoError(t, err)

	return string(output)
}
