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

// TestAddContract_SuccessfulAdd tests the contract for successful consignment creation
func TestAddContract_SuccessfulAdd(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Initialize shipyard
	initCmd := exec.Command(shipyardBin, "init", "--yes")
	initCmd.Dir = tempDir
	_, err := initCmd.CombinedOutput()
	require.NoError(t, err, "Init should succeed")

	// Run shipyard add
	cmd := exec.Command(shipyardBin, "add",
		"--package", "default",
		"--type", "patch",
		"--summary", "Test consignment")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify exit code
	require.NoError(t, err, "Command should exit with code 0")

	// Verify output contains success message
	outputStr := string(output)
	assert.Contains(t, outputStr, "Created consignment", "Output should contain success message")
	assert.Contains(t, outputStr, "Packages: default", "Output should mention package")
	assert.Contains(t, outputStr, "Type: patch", "Output should mention type")

	// Verify consignment file was created
	consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
	entries, err := os.ReadDir(consignmentsDir)
	require.NoError(t, err, "Should be able to read consignments directory")
	assert.Equal(t, 1, len(entries), "Should have created one consignment file")
	assert.True(t, strings.HasSuffix(entries[0].Name(), ".md"), "Consignment should be markdown file")
}

// TestAddContract_MultiplePackages tests adding consignment for multiple packages
func TestAddContract_MultiplePackages(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Initialize shipyard
	initCmd := exec.Command(shipyardBin, "init", "--yes")
	initCmd.Dir = tempDir
	_, err := initCmd.CombinedOutput()
	require.NoError(t, err, "Init should succeed")

	// Run shipyard add with multiple packages
	cmd := exec.Command(shipyardBin, "add",
		"--package", "default",
		"--package", "default", // Note: in a real scenario, these would be different packages
		"--type", "minor",
		"--summary", "Cross-cutting change")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify exit code
	require.NoError(t, err, "Command should exit with code 0")

	// Verify output
	outputStr := string(output)
	assert.Contains(t, outputStr, "Created consignment", "Output should contain success message")
}

// TestAddContract_WithMetadata tests adding consignment with metadata
func TestAddContract_WithMetadata(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Initialize shipyard
	initCmd := exec.Command(shipyardBin, "init", "--yes")
	initCmd.Dir = tempDir
	_, err := initCmd.CombinedOutput()
	require.NoError(t, err, "Init should succeed")

	// Run shipyard add with metadata
	cmd := exec.Command(shipyardBin, "add",
		"--package", "default",
		"--type", "patch",
		"--summary", "Fixed bug",
		"--metadata", "author=test@example.com",
		"--metadata", "issue=JIRA-123")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify exit code
	require.NoError(t, err, "Command should exit with code 0")

	// Verify output
	outputStr := string(output)
	assert.Contains(t, outputStr, "Created consignment", "Output should contain success message")

	// Verify metadata in file
	consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
	entries, err := os.ReadDir(consignmentsDir)
	require.NoError(t, err, "Should be able to read consignments directory")
	require.Equal(t, 1, len(entries), "Should have one consignment")

	// Read file and check metadata
	consignmentPath := filepath.Join(consignmentsDir, entries[0].Name())
	content, err := os.ReadFile(consignmentPath)
	require.NoError(t, err, "Should be able to read consignment")
	assert.Contains(t, string(content), "author: test@example.com", "Should contain author metadata")
	assert.Contains(t, string(content), "issue: JIRA-123", "Should contain issue metadata")
}

// TestAddContract_InvalidPackage tests error handling for invalid package
func TestAddContract_InvalidPackage(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Initialize shipyard
	initCmd := exec.Command(shipyardBin, "init", "--yes")
	initCmd.Dir = tempDir
	_, err := initCmd.CombinedOutput()
	require.NoError(t, err, "Init should succeed")

	// Run shipyard add with invalid package
	cmd := exec.Command(shipyardBin, "add",
		"--package", "nonexistent",
		"--type", "patch",
		"--summary", "Test")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify exit code (should fail)
	assert.Error(t, err, "Command should exit with non-zero code")

	// Verify output contains error message
	outputStr := string(output)
	assert.Contains(t, outputStr, "invalid package", "Output should mention invalid package")
}

// TestAddContract_InvalidChangeType tests error handling for invalid change type
func TestAddContract_InvalidChangeType(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Initialize shipyard
	initCmd := exec.Command(shipyardBin, "init", "--yes")
	initCmd.Dir = tempDir
	_, err := initCmd.CombinedOutput()
	require.NoError(t, err, "Init should succeed")

	// Run shipyard add with invalid change type
	cmd := exec.Command(shipyardBin, "add",
		"--package", "default",
		"--type", "breaking",
		"--summary", "Test")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify exit code (should fail)
	assert.Error(t, err, "Command should exit with non-zero code")

	// Verify output contains error message
	outputStr := string(output)
	assert.Contains(t, outputStr, "change type", "Output should mention change type")
}

// TestAddContract_HelpFlag tests the contract for --help flag
func TestAddContract_HelpFlag(t *testing.T) {
	shipyardBin := buildShipyard(t)

	// Run shipyard add --help
	cmd := exec.Command(shipyardBin, "add", "--help")
	output, err := cmd.CombinedOutput()

	// Verify exit code (should succeed)
	require.NoError(t, err, "Help command should succeed")

	// Verify output contains help information
	outputStr := string(output)
	assert.Contains(t, outputStr, "Record new cargo", "Output should contain command description")
	assert.Contains(t, outputStr, "Usage:", "Output should contain usage section")
	assert.Contains(t, outputStr, "--package", "Output should mention --package flag")
	assert.Contains(t, outputStr, "--type", "Output should mention --type flag")
	assert.Contains(t, outputStr, "--summary", "Output should mention --summary flag")
	assert.Contains(t, outputStr, "--metadata", "Output should mention --metadata flag")
}

// TestAddContract_NotInitialized tests behavior when Shipyard is not initialized
func TestAddContract_NotInitialized(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Don't initialize shipyard

	// Run shipyard add
	cmd := exec.Command(shipyardBin, "add",
		"--package", "core",
		"--type", "patch",
		"--summary", "Test")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify exit code (should fail)
	assert.Error(t, err, "Command should exit with non-zero code")

	// Verify output contains error message
	outputStr := string(output)
	// The error could be about config not found or not being initialized
	assert.True(t,
		strings.Contains(outputStr, "Config File") ||
		strings.Contains(outputStr, "config") ||
		strings.Contains(outputStr, "Not Found"),
		"Output should mention configuration issue")
}

// TestAddContract_NotGitRepository tests behavior when not in git repository
func TestAddContract_NotGitRepository(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()

	// Don't initialize git

	// Run shipyard add
	cmd := exec.Command(shipyardBin, "add",
		"--package", "core",
		"--type", "patch",
		"--summary", "Test")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify exit code (should fail)
	assert.Error(t, err, "Command should exit with non-zero code")

	// Verify output contains error message
	outputStr := string(output)
	assert.Contains(t, outputStr, "git repository", "Output should mention git repository requirement")
}

// TestAddContract_QuietFlag tests the contract for --quiet flag
func TestAddContract_QuietFlag(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Initialize shipyard
	initCmd := exec.Command(shipyardBin, "init", "--yes")
	initCmd.Dir = tempDir
	_, err := initCmd.CombinedOutput()
	require.NoError(t, err, "Init should succeed")

	// Run shipyard add with --quiet
	cmd := exec.Command(shipyardBin, "add", "--quiet",
		"--package", "default",
		"--type", "patch",
		"--summary", "Test")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify exit code (should succeed)
	require.NoError(t, err, "Add with --quiet should succeed")

	// Verify output is minimal (no INFO logs)
	outputStr := string(output)
	assert.NotContains(t, outputStr, "[INFO]", "Quiet mode should suppress INFO logs")

	// But file should still be created
	consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
	entries, err := os.ReadDir(consignmentsDir)
	require.NoError(t, err, "Should be able to read consignments directory")
	assert.Equal(t, 1, len(entries), "Should have created consignment file")
}
