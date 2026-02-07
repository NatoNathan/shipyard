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

// TestRemoveContract_RemoveById tests removing a single consignment by its ID
func TestRemoveContract_RemoveById(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)
	createTestConsignment(t, tempDir, "c1", "core", "patch", "Fix1")
	createTestConsignment(t, tempDir, "c2", "core", "minor", "Feature")

	// Run remove --id c1
	cmd := exec.Command(shipyardBin, "remove", "--id", "c1")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify: exit 0
	require.NoError(t, err, "remove --id should exit 0, got: %s", string(output))

	outputStr := string(output)
	assert.Contains(t, outputStr, "Removed", "Output should contain 'Removed'")
	assert.Contains(t, outputStr, "1", "Output should contain count '1'")

	// Verify c1.md is gone
	c1Path := filepath.Join(tempDir, ".shipyard", "consignments", "c1.md")
	_, err = os.Stat(c1Path)
	assert.True(t, os.IsNotExist(err), "c1.md should have been removed")

	// Verify c2.md still exists
	c2Path := filepath.Join(tempDir, ".shipyard", "consignments", "c2.md")
	_, err = os.Stat(c2Path)
	assert.NoError(t, err, "c2.md should still exist")
}

// TestRemoveContract_RemoveAll tests removing all pending consignments
func TestRemoveContract_RemoveAll(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)
	createTestConsignment(t, tempDir, "c1", "core", "patch", "Fix1")
	createTestConsignment(t, tempDir, "c2", "core", "minor", "Fix2")
	createTestConsignment(t, tempDir, "c3", "core", "major", "Breaking")

	// Run remove --all
	cmd := exec.Command(shipyardBin, "remove", "--all")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify: exit 0
	require.NoError(t, err, "remove --all should exit 0, got: %s", string(output))

	outputStr := string(output)
	assert.Contains(t, outputStr, "Removed", "Output should contain 'Removed'")
	assert.Contains(t, outputStr, "3", "Output should contain count '3'")

	// Verify consignments directory is empty
	consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
	entries, err := os.ReadDir(consignmentsDir)
	require.NoError(t, err, "Should be able to read consignments directory")
	assert.Equal(t, 0, len(entries), "Consignments directory should be empty")
}

// TestRemoveContract_RemoveAllEmpty tests remove --all when there are no consignments
func TestRemoveContract_RemoveAllEmpty(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	// Run remove --all with no consignments
	cmd := exec.Command(shipyardBin, "remove", "--all")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify: exit 0
	require.NoError(t, err, "remove --all with no consignments should exit 0, got: %s", string(output))

	outputStr := strings.ToLower(string(output))
	assert.True(t,
		strings.Contains(outputStr, "no pending consignments") || strings.Contains(outputStr, "0"),
		"Output should contain 'No pending consignments' or '0', got: %s", string(output))
}

// TestRemoveContract_NoFlagsError tests that remove without flags returns an error
func TestRemoveContract_NoFlagsError(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	// Run remove with no flags
	cmd := exec.Command(shipyardBin, "remove")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify: non-zero exit
	assert.Error(t, err, "remove without flags should exit with non-zero code")

	outputStr := strings.ToLower(string(output))
	assert.True(t,
		strings.Contains(outputStr, "specify --id or --all") ||
			strings.Contains(outputStr, "--id") ||
			strings.Contains(outputStr, "--all"),
		"Output should mention --id or --all, got: %s", string(output))
}

// TestRemoveContract_NotFoundError tests removing a nonexistent consignment ID
func TestRemoveContract_NotFoundError(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	// Run remove --id with nonexistent ID
	cmd := exec.Command(shipyardBin, "remove", "--id", "nonexistent")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify: non-zero exit
	assert.Error(t, err, "remove --id nonexistent should exit with non-zero code")

	outputStr := strings.ToLower(string(output))
	assert.Contains(t, outputStr, "not found", "Output should contain 'not found'")
}

// TestRemoveContract_JSONOutput tests that global --json flag produces JSON output
func TestRemoveContract_JSONOutput(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)
	createTestConsignment(t, tempDir, "c1", "core", "patch", "Fix1")
	createTestConsignment(t, tempDir, "c2", "core", "minor", "Fix2")

	// Run with global --json flag before subcommand
	cmd := exec.Command(shipyardBin, "--json", "remove", "--all")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify: exit 0
	require.NoError(t, err, "--json remove --all should exit 0, got: %s", string(output))

	outputStr := string(output)
	assert.Contains(t, outputStr, `"removed"`, "JSON output should contain 'removed' key")
	assert.Contains(t, outputStr, `"count"`, "JSON output should contain 'count' key")
}

// TestRemoveContract_HelpFlag tests the remove --help output
func TestRemoveContract_HelpFlag(t *testing.T) {
	shipyardBin := buildShipyard(t)

	// Run remove --help (no repo setup needed)
	cmd := exec.Command(shipyardBin, "remove", "--help")
	output, err := cmd.CombinedOutput()

	// Verify: exit 0
	require.NoError(t, err, "remove --help should exit 0")

	outputStr := string(output)
	assert.Contains(t, outputStr, "Remove one or more pending consignments", "Help output should contain command description")
}
