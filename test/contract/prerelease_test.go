package contract

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const prereleaseConfigYAML = `packages:
  - name: core
    path: ./core
    ecosystem: go
prerelease:
  stages:
    - name: alpha
      order: 1
    - name: beta
      order: 2
    - name: rc
      order: 3
`

// TestPrereleaseContract_FirstPrerelease tests creating the first pre-release version
func TestPrereleaseContract_FirstPrerelease(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)
	writeConfig(t, tempDir, prereleaseConfigYAML)
	createInitialCommit(t, tempDir)
	createTestConsignment(t, tempDir, "c1", "core", "minor", "New feature")

	cmd := exec.Command(shipyardBin, "version", "prerelease", "--no-commit", "--no-tag")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "prerelease should exit 0, output: %s", string(output))

	// Version file should contain alpha.1
	versionContent := readFileContent(t, filepath.Join(tempDir, "core", "version.go"))
	assert.Contains(t, versionContent, "alpha.1", "version file should contain alpha.1 prerelease identifier")

	// Prerelease state file should exist and contain stage info
	stateContent := readFileContent(t, filepath.Join(tempDir, ".shipyard", "prerelease.yml"))
	assert.Contains(t, stateContent, "alpha", "prerelease state should contain alpha stage")
	assert.Contains(t, stateContent, "counter", "prerelease state should contain counter")
}

// TestPrereleaseContract_IncrementCounter tests that running prerelease again increments the counter
func TestPrereleaseContract_IncrementCounter(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)
	writeConfig(t, tempDir, prereleaseConfigYAML)

	// Write existing prerelease state (alpha.1 already done)
	writePrereleaseState(t, tempDir, `packages:
  core:
    stage: alpha
    counter: 1
    targetVersion: "1.1.0"
`)

	createInitialCommit(t, tempDir)
	createTestConsignment(t, tempDir, "c1", "core", "minor", "New feature")

	cmd := exec.Command(shipyardBin, "version", "prerelease", "--no-commit", "--no-tag")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "prerelease should exit 0, output: %s", string(output))

	// Version file should contain alpha.2
	versionContent := readFileContent(t, filepath.Join(tempDir, "core", "version.go"))
	assert.Contains(t, versionContent, "alpha.2", "version file should contain alpha.2 after increment")

	// Prerelease state should show counter: 2
	stateContent := readFileContent(t, filepath.Join(tempDir, ".shipyard", "prerelease.yml"))
	assert.True(t,
		strings.Contains(stateContent, "counter: 2") || strings.Contains(stateContent, "counter:2"),
		"prerelease state should contain counter: 2, got: %s", stateContent)
}

// TestPrereleaseContract_PreviewMode tests that preview mode does not modify files
func TestPrereleaseContract_PreviewMode(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)
	writeConfig(t, tempDir, prereleaseConfigYAML)
	createInitialCommit(t, tempDir)
	createTestConsignment(t, tempDir, "c1", "core", "minor", "New feature")

	cmd := exec.Command(shipyardBin, "version", "prerelease", "--preview")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "prerelease --preview should exit 0, output: %s", string(output))

	// Output should mention preview
	outputStr := strings.ToLower(string(output))
	assert.Contains(t, outputStr, "preview", "output should mention preview mode")

	// Version file should still contain the original version
	versionContent := readFileContent(t, filepath.Join(tempDir, "core", "version.go"))
	assert.Contains(t, versionContent, "1.0.0", "version file should still contain 1.0.0 in preview mode")
}

// TestPrereleaseContract_NoConsignments tests error when no consignments exist
func TestPrereleaseContract_NoConsignments(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)
	writeConfig(t, tempDir, prereleaseConfigYAML)
	createInitialCommit(t, tempDir)
	// No consignment created

	cmd := exec.Command(shipyardBin, "version", "prerelease")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	assert.Error(t, err, "prerelease should fail with no consignments")
	exitCode := getExitCode(err)
	assert.NotEqual(t, 0, exitCode, "exit code should be non-zero")

	outputStr := strings.ToLower(string(output))
	assert.Contains(t, outputStr, "no pending consignments", "output should mention no pending consignments")
}

// TestPrereleaseContract_NoStagesConfigured tests error when no prerelease stages are configured
func TestPrereleaseContract_NoStagesConfigured(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)
	// Default config from init (no prerelease stages)
	createInitialCommit(t, tempDir)
	createTestConsignment(t, tempDir, "c1", "core", "minor", "New feature")

	cmd := exec.Command(shipyardBin, "version", "prerelease")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	assert.Error(t, err, "prerelease should fail with no stages configured")
	exitCode := getExitCode(err)
	assert.NotEqual(t, 0, exitCode, "exit code should be non-zero")

	outputStr := strings.ToLower(string(output))
	assert.True(t,
		strings.Contains(outputStr, "no pre-release stages") || strings.Contains(outputStr, "stages"),
		"output should mention missing stages, got: %s", outputStr)
}

// TestPrereleaseContract_JSONOutput tests JSON output with global --json flag
func TestPrereleaseContract_JSONOutput(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)
	writeConfig(t, tempDir, prereleaseConfigYAML)
	createInitialCommit(t, tempDir)
	createTestConsignment(t, tempDir, "c1", "core", "minor", "Feature")

	cmd := exec.Command(shipyardBin, "--json", "version", "prerelease", "--no-commit", "--no-tag")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "prerelease with --json should exit 0, output: %s", string(output))

	outputStr := string(output)
	assert.Contains(t, outputStr, `"stage"`, "JSON output should contain stage key")
	assert.Contains(t, outputStr, `"alpha"`, "JSON output should contain alpha value")
}

// TestPrereleaseContract_HelpFlag tests the --help flag output
func TestPrereleaseContract_HelpFlag(t *testing.T) {
	shipyardBin := buildShipyard(t)

	cmd := exec.Command(shipyardBin, "version", "prerelease", "--help")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "prerelease --help should exit 0")

	outputStr := string(output)
	assert.Contains(t, outputStr, "Create or increment a pre-release version", "help output should contain command description")
}
