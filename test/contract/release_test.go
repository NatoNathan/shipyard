package contract

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// filterEnv returns a copy of env with entries starting with key+"=" removed.
func filterEnv(env []string, key string) []string {
	prefix := key + "="
	filtered := make([]string, 0, len(env))
	for _, e := range env {
		if !strings.HasPrefix(e, prefix) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// TestReleaseContract_MissingGitHubConfig tests that release fails when no github section is configured
func TestReleaseContract_MissingGitHubConfig(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	// Default config from init has no github section
	cmd := exec.Command(shipyardBin, "release")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	outputStr := string(output)
	assert.Error(t, err, "release should fail when GitHub is not configured")
	assert.True(t,
		strings.Contains(strings.ToLower(outputStr), "github") &&
			(strings.Contains(strings.ToLower(outputStr), "configured") || strings.Contains(strings.ToLower(outputStr), "config")),
		"Output should mention GitHub configuration issue, got: %s", outputStr)
}

// TestReleaseContract_MissingGitHubToken tests that release fails when GITHUB_TOKEN is not set
func TestReleaseContract_MissingGitHubToken(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	writeConfig(t, tempDir, `packages:
  - name: core
    path: ./core
    ecosystem: go
github:
  owner: testowner
  repo: testrepo
`)

	cmd := exec.Command(shipyardBin, "release")
	cmd.Dir = tempDir
	cmd.Env = filterEnv(os.Environ(), "GITHUB_TOKEN")
	output, err := cmd.CombinedOutput()

	outputStr := string(output)
	assert.Error(t, err, "release should fail when GITHUB_TOKEN is not set")
	assert.True(t,
		strings.Contains(outputStr, "GITHUB_TOKEN") || strings.Contains(strings.ToLower(outputStr), "token"),
		"Output should mention GITHUB_TOKEN or token, got: %s", outputStr)
}

// TestReleaseContract_NoHistory tests that release fails when there is no release history for the package
func TestReleaseContract_NoHistory(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	writeConfig(t, tempDir, `packages:
  - name: core
    path: ./core
    ecosystem: go
github:
  owner: testowner
  repo: testrepo
`)

	cmd := exec.Command(shipyardBin, "release", "--package", "core")
	cmd.Dir = tempDir
	cmd.Env = append(filterEnv(os.Environ(), "GITHUB_TOKEN"), "GITHUB_TOKEN=fake")
	output, err := cmd.CombinedOutput()

	outputStr := strings.ToLower(string(output))
	assert.Error(t, err, "release should fail when there is no history")
	assert.True(t,
		strings.Contains(outputStr, "no releases found") ||
			strings.Contains(outputStr, "history") ||
			strings.Contains(outputStr, "tag"),
		"Output should mention missing releases, history, or tag, got: %s", string(output))
}

// TestReleaseContract_MultiPackageRequiresFlag tests that release fails without --package in multi-package repos
func TestReleaseContract_MultiPackageRequiresFlag(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepoMultiPackage(t, shipyardBin, tempDir)

	writeConfig(t, tempDir, `packages:
  - name: core
    path: ./core
    ecosystem: go
  - name: api
    path: ./api
    ecosystem: go
  - name: web
    path: ./web
    ecosystem: go
github:
  owner: testowner
  repo: testrepo
`)

	cmd := exec.Command(shipyardBin, "release")
	cmd.Dir = tempDir
	cmd.Env = append(filterEnv(os.Environ(), "GITHUB_TOKEN"), "GITHUB_TOKEN=fake")
	output, err := cmd.CombinedOutput()

	outputStr := strings.ToLower(string(output))
	assert.Error(t, err, "release should fail without --package in multi-package repo")
	assert.True(t,
		strings.Contains(outputStr, "--package") || strings.Contains(outputStr, "package is required"),
		"Output should mention --package flag or that package is required, got: %s", string(output))
}

// TestReleaseContract_HelpFlag tests that release --help exits 0 and shows usage
func TestReleaseContract_HelpFlag(t *testing.T) {
	shipyardBin := buildShipyard(t)

	cmd := exec.Command(shipyardBin, "release", "--help")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "release --help should exit 0")
	assert.Contains(t, string(output), "Publish a version release to GitHub",
		"Help output should contain command description")
}
