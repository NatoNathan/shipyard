package contract

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromoteContract_AlphaToBeta(t *testing.T) {
	shipyardBin := buildShipyard(t)

	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	writeConfig(t, tempDir, `packages:
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
`)

	createTestConsignment(t, tempDir, "c1", "core", "minor", "Feature")

	writePrereleaseState(t, tempDir, `packages:
  core:
    stage: alpha
    counter: 1
    targetVersion: "1.1.0"
`)

	createInitialCommit(t, tempDir)

	cmd := exec.Command(shipyardBin, "version", "promote", "--no-commit", "--no-tag")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "promote should exit 0: %s", string(output))

	versionContent := readFileContent(t, filepath.Join(tempDir, "core", "version.go"))
	assert.Contains(t, versionContent, "beta.1", "version file should contain beta.1")

	prereleaseContent := readFileContent(t, filepath.Join(tempDir, ".shipyard", "prerelease.yml"))
	assert.Contains(t, prereleaseContent, "beta", "prerelease.yml should contain beta")
}

func TestPromoteContract_PreviewMode(t *testing.T) {
	shipyardBin := buildShipyard(t)

	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	writeConfig(t, tempDir, `packages:
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
`)

	createTestConsignment(t, tempDir, "c1", "core", "minor", "Feature")

	writePrereleaseState(t, tempDir, `packages:
  core:
    stage: alpha
    counter: 1
    targetVersion: "1.1.0"
`)

	createInitialCommit(t, tempDir)

	cmd := exec.Command(shipyardBin, "version", "promote", "--preview")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "promote --preview should exit 0: %s", string(output))
	assert.Contains(t, strings.ToLower(string(output)), "preview", "output should contain 'preview'")

	versionContent := readFileContent(t, filepath.Join(tempDir, "core", "version.go"))
	assert.True(t,
		strings.Contains(versionContent, "1.0.0") || strings.Contains(versionContent, "alpha"),
		"version file should be unchanged (still 1.0.0 or alpha)")
}

func TestPromoteContract_NoState(t *testing.T) {
	shipyardBin := buildShipyard(t)

	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	writeConfig(t, tempDir, `packages:
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
`)

	createInitialCommit(t, tempDir)

	cmd := exec.Command(shipyardBin, "version", "promote")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	exitCode := getExitCode(err)
	assert.NotEqual(t, 0, exitCode, "promote with no state should exit non-zero")

	outputLower := strings.ToLower(string(output))
	assert.True(t,
		strings.Contains(outputLower, "no pre-release state") || strings.Contains(outputLower, "prerelease"),
		"output should mention missing pre-release state, got: %s", string(output))
}

func TestPromoteContract_AtHighestStage(t *testing.T) {
	shipyardBin := buildShipyard(t)

	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	writeConfig(t, tempDir, `packages:
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
`)

	createTestConsignment(t, tempDir, "c1", "core", "minor", "Feature")

	writePrereleaseState(t, tempDir, `packages:
  core:
    stage: rc
    counter: 1
    targetVersion: "1.1.0"
`)

	createInitialCommit(t, tempDir)

	cmd := exec.Command(shipyardBin, "version", "promote")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	exitCode := getExitCode(err)
	assert.NotEqual(t, 0, exitCode, "promote at highest stage should exit non-zero")

	outputLower := strings.ToLower(string(output))
	assert.True(t,
		strings.Contains(outputLower, "already at highest") || strings.Contains(outputLower, "highest"),
		"output should mention already at highest stage, got: %s", string(output))
}

func TestPromoteContract_JSONOutput(t *testing.T) {
	shipyardBin := buildShipyard(t)

	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	writeConfig(t, tempDir, `packages:
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
`)

	createTestConsignment(t, tempDir, "c1", "core", "minor", "Feature")

	writePrereleaseState(t, tempDir, `packages:
  core:
    stage: alpha
    counter: 1
    targetVersion: "1.1.0"
`)

	createInitialCommit(t, tempDir)

	cmd := exec.Command(shipyardBin, "--json", "version", "promote", "--no-commit", "--no-tag")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "promote with --json should exit 0: %s", string(output))

	outputStr := string(output)
	hasOldStage := strings.Contains(outputStr, `"oldStage"`) || strings.Contains(outputStr, `"old_stage"`) || strings.Contains(outputStr, `"alpha"`)
	hasBeta := strings.Contains(outputStr, `"beta"`)
	assert.True(t, hasOldStage, "JSON output should contain oldStage or alpha, got: %s", outputStr)
	assert.True(t, hasBeta, "JSON output should contain beta, got: %s", outputStr)
}

func TestPromoteContract_HelpFlag(t *testing.T) {
	shipyardBin := buildShipyard(t)

	cmd := exec.Command(shipyardBin, "version", "promote", "--help")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "promote --help should exit 0: %s", string(output))
	assert.Contains(t, string(output), "Promote a pre-release to the next stage", "help output should describe the command")
}
