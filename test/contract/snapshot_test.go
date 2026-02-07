package contract

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSnapshotContract_CreateSnapshot tests that snapshot creates a timestamped pre-release version
func TestSnapshotContract_CreateSnapshot(t *testing.T) {
	shipyardBin := buildShipyard(t)

	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)
	createInitialCommit(t, tempDir)
	createTestConsignment(t, tempDir, "c1", "core", "patch", "Fix")

	// Run version snapshot without git operations
	cmd := exec.Command(shipyardBin, "version", "snapshot", "--no-commit", "--no-tag")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify: exit 0
	require.NoError(t, err, "snapshot should exit 0, output: %s", string(output))

	// Verify: version file contains snapshot with expected format
	versionContent := readFileContent(t, filepath.Join(tempDir, "core", "version.go"))
	assert.Contains(t, versionContent, "snapshot", "version file should contain snapshot pre-release identifier")

	// Verify: snapshot version matches expected timestamp format (snapshot.YYYYMMDD-HHMMSS)
	snapshotPattern := regexp.MustCompile(`snapshot\.\d{8}-\d{6}`)
	assert.True(t, snapshotPattern.MatchString(versionContent), "version should match snapshot.YYYYMMDD-HHMMSS format, got: %s", versionContent)
}

// TestSnapshotContract_PreviewMode tests that preview mode shows changes without applying them
func TestSnapshotContract_PreviewMode(t *testing.T) {
	shipyardBin := buildShipyard(t)

	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)
	createInitialCommit(t, tempDir)
	createTestConsignment(t, tempDir, "c1", "core", "patch", "Fix")

	// Run version snapshot in preview mode
	cmd := exec.Command(shipyardBin, "version", "snapshot", "--preview")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify: exit 0
	require.NoError(t, err, "snapshot --preview should exit 0, output: %s", string(output))

	// Verify: output contains "preview" (case-insensitive)
	outputStr := string(output)
	assert.True(t, strings.Contains(strings.ToLower(outputStr), "preview"), "output should contain 'preview', got: %s", outputStr)

	// Verify: version file is unchanged (still 1.0.0)
	versionContent := readFileContent(t, filepath.Join(tempDir, "core", "version.go"))
	assert.Contains(t, versionContent, "1.0.0", "version file should still contain 1.0.0 in preview mode")
}

// TestSnapshotContract_NoConsignments tests that snapshot fails when there are no pending consignments
func TestSnapshotContract_NoConsignments(t *testing.T) {
	shipyardBin := buildShipyard(t)

	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)
	createInitialCommit(t, tempDir)

	// Run version snapshot with no consignments
	cmd := exec.Command(shipyardBin, "version", "snapshot")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify: non-zero exit
	assert.Error(t, err, "snapshot should fail with no consignments")
	exitCode := getExitCode(err)
	assert.NotEqual(t, 0, exitCode, "exit code should be non-zero")

	// Verify: output mentions no pending consignments
	outputStr := strings.ToLower(string(output))
	assert.Contains(t, outputStr, "no pending consignments", "output should mention 'no pending consignments', got: %s", string(output))
}

// TestSnapshotContract_JSONOutput tests that snapshot produces valid JSON with --json flag
func TestSnapshotContract_JSONOutput(t *testing.T) {
	shipyardBin := buildShipyard(t)

	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)
	createInitialCommit(t, tempDir)
	createTestConsignment(t, tempDir, "c1", "core", "patch", "Fix")

	// Run with global --json flag
	cmd := exec.Command(shipyardBin, "--json", "version", "snapshot", "--no-commit", "--no-tag")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify: exit 0
	require.NoError(t, err, "snapshot --json should exit 0, output: %s", string(output))

	// Verify: output contains JSON fields
	outputStr := string(output)
	assert.True(t,
		strings.Contains(outputStr, `"timestamp"`) || strings.Contains(outputStr, `"snapshot"`),
		"JSON output should contain 'timestamp' or 'snapshot', got: %s", outputStr)
}

// TestSnapshotContract_HelpFlag tests that --help displays usage information
func TestSnapshotContract_HelpFlag(t *testing.T) {
	shipyardBin := buildShipyard(t)

	// Run version snapshot --help (no repo setup needed)
	cmd := exec.Command(shipyardBin, "version", "snapshot", "--help")
	output, err := cmd.CombinedOutput()

	// Verify: exit 0
	require.NoError(t, err, "snapshot --help should exit 0, output: %s", string(output))

	// Verify: output contains the command description
	outputStr := string(output)
	assert.Contains(t, outputStr, "Create a timestamped snapshot pre-release version", "help output should contain command short description")
}
