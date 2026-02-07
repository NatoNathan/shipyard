package contract

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVersionContract_PreviewMode tests that --preview shows changes without modifying files
func TestVersionContract_PreviewMode(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)
	createTestConsignment(t, tempDir, "c1", "core", "patch", "Fix bug")

	// Run version --preview
	cmd := exec.Command(shipyardBin, "version", "--preview")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify: exit 0
	require.NoError(t, err, "version --preview should exit 0, output: %s", string(output))

	// Verify: output contains "Preview" (case insensitive)
	assert.True(t, strings.Contains(strings.ToLower(string(output)), "preview"),
		"Output should contain 'Preview', got: %s", string(output))

	// Verify: version file still contains 1.0.0 (not modified)
	versionContent := readFileContent(t, filepath.Join(tempDir, "core", "version.go"))
	assert.Contains(t, versionContent, "1.0.0", "Version file should still contain 1.0.0 in preview mode")

	// Verify: consignment file still exists
	consignmentPath := filepath.Join(tempDir, ".shipyard", "consignments", "c1.md")
	_, err = os.Stat(consignmentPath)
	assert.NoError(t, err, "Consignment file should still exist after preview")
}

// TestVersionContract_NoConsignmentsNoop tests that version with no consignments is a no-op
func TestVersionContract_NoConsignmentsNoop(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	// Run version --verbose (no consignments present)
	cmd := exec.Command(shipyardBin, "version", "--verbose")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify: exit 0
	require.NoError(t, err, "version with no consignments should exit 0, output: %s", string(output))

	// Verify: output contains "No pending consignments" (case insensitive)
	assert.True(t, strings.Contains(strings.ToLower(string(output)), "no pending consignments"),
		"Output should contain 'No pending consignments', got: %s", string(output))
}

// TestVersionContract_ApplyWithNoCommitNoTag tests that version applies changes without git operations
func TestVersionContract_ApplyWithNoCommitNoTag(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)
	createTestConsignment(t, tempDir, "c1", "core", "patch", "Fix bug")

	// Run version --no-commit --no-tag
	cmd := exec.Command(shipyardBin, "version", "--no-commit", "--no-tag")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify: exit 0
	require.NoError(t, err, "version --no-commit --no-tag should exit 0, output: %s", string(output))

	// Verify: version file contains 1.0.1 (bumped from 1.0.0)
	versionContent := readFileContent(t, filepath.Join(tempDir, "core", "version.go"))
	assert.Contains(t, versionContent, "1.0.1", "Version file should contain bumped version 1.0.1")
	assert.NotContains(t, versionContent, "1.0.0", "Version file should no longer contain 1.0.0")

	// Verify: consignment file is deleted
	consignmentPath := filepath.Join(tempDir, ".shipyard", "consignments", "c1.md")
	_, err = os.Stat(consignmentPath)
	assert.True(t, os.IsNotExist(err), "Consignment file should be deleted after version apply")

	// Verify: history.json contains the new version
	historyContent := readFileContent(t, filepath.Join(tempDir, ".shipyard", "history.json"))
	assert.Contains(t, historyContent, "1.0.1", "History should contain the new version 1.0.1")
}

// TestVersionContract_ApplyWithCommitAndTag tests that version creates git commit and tag
func TestVersionContract_ApplyWithCommitAndTag(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)
	createTestConsignment(t, tempDir, "c1", "core", "patch", "Fix bug")
	// Create initial commit with all files including the consignment
	createInitialCommit(t, tempDir)

	// Run version (default: commit + tag)
	cmd := exec.Command(shipyardBin, "version")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify: exit 0
	require.NoError(t, err, "version should exit 0, output: %s", string(output))

	// Verify: git tag exists
	tagCmd := exec.Command("git", "tag", "-l")
	tagCmd.Dir = tempDir
	tagOutput, err := tagCmd.CombinedOutput()
	require.NoError(t, err, "git tag -l should succeed")
	assert.NotEmpty(t, strings.TrimSpace(string(tagOutput)), "Should have at least one git tag after version")
}

// TestVersionContract_PackageFilter tests that --package filters version bumps to specific packages
func TestVersionContract_PackageFilter(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepoMultiPackage(t, shipyardBin, tempDir)
	createTestConsignment(t, tempDir, "c1", "core", "patch", "Core fix")
	createTestConsignment(t, tempDir, "c2", "api", "minor", "API feature")

	// Run version --package core --no-commit --no-tag
	cmd := exec.Command(shipyardBin, "version", "--package", "core", "--no-commit", "--no-tag")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify: exit 0
	require.NoError(t, err, "version --package core should exit 0, output: %s", string(output))

	// Verify: core/version.go contains 1.0.1 (bumped)
	coreVersion := readFileContent(t, filepath.Join(tempDir, "core", "version.go"))
	assert.Contains(t, coreVersion, "1.0.1", "core/version.go should contain bumped version 1.0.1")

	// Verify: api/version.go still contains 1.0.0 (not bumped)
	apiVersion := readFileContent(t, filepath.Join(tempDir, "api", "version.go"))
	assert.Contains(t, apiVersion, "1.0.0", "api/version.go should still contain 1.0.0")
}

// TestVersionContract_VerboseOutput tests that --verbose produces detailed output
func TestVersionContract_VerboseOutput(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)
	createTestConsignment(t, tempDir, "c1", "core", "patch", "Fix bug")

	// Run version --no-commit --no-tag --verbose
	cmd := exec.Command(shipyardBin, "version", "--no-commit", "--no-tag", "--verbose")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify: exit 0
	require.NoError(t, err, "version --verbose should exit 0, output: %s", string(output))

	// Verify: output contains package name and version info
	outputStr := string(output)
	assert.Contains(t, outputStr, "core", "Verbose output should contain package name 'core'")
	assert.Contains(t, outputStr, "1.0.1", "Verbose output should contain the new version")
}

// TestVersionContract_HelpFlag tests that --help displays usage information
func TestVersionContract_HelpFlag(t *testing.T) {
	shipyardBin := buildShipyard(t)

	// Run version --help (no repo setup needed)
	cmd := exec.Command(shipyardBin, "version", "--help")
	output, err := cmd.CombinedOutput()

	// Verify: exit 0
	require.NoError(t, err, "version --help should exit 0")

	// Verify: output contains expected help content
	outputStr := string(output)
	assert.Contains(t, outputStr, "Set sail with your cargo", "Help should contain command description")
	assert.Contains(t, outputStr, "--preview", "Help should mention --preview flag")
	assert.Contains(t, outputStr, "--no-commit", "Help should mention --no-commit flag")
}
