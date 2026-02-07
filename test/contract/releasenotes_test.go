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

// TestReleaseNotesContract_BasicOutput tests that release-notes outputs version and summary
func TestReleaseNotesContract_BasicOutput(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	writeHistoryJSON(t, tempDir, `[
  {
    "version": "1.0.1",
    "package": "core",
    "tag": "core/v1.0.1",
    "timestamp": "2026-01-30T10:00:00Z",
    "consignments": [
      {
        "id": "c1",
        "summary": "Fixed critical bug",
        "changeType": "patch"
      }
    ]
  }
]`)

	cmd := exec.Command(shipyardBin, "release-notes", "--package", "core")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "release-notes should exit 0: %s", string(output))
	assert.Contains(t, string(output), "1.0.1", "Output should contain version")
	assert.Contains(t, string(output), "Fixed critical bug", "Output should contain summary")
}

// TestReleaseNotesContract_NoHistory tests behavior when no release history exists
func TestReleaseNotesContract_NoHistory(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	writeHistoryJSON(t, tempDir, `[]`)

	cmd := exec.Command(shipyardBin, "release-notes", "--package", "core")
	cmd.Dir = tempDir
	output, _ := cmd.CombinedOutput()

	outputStr := strings.ToLower(string(output))
	assert.True(t,
		strings.Contains(outputStr, "no releases found") ||
			strings.Contains(outputStr, "no releases"),
		"Output should indicate no releases found, got: %s", string(output))
}

// TestReleaseNotesContract_JSONOutput tests that global --json flag produces JSON output
func TestReleaseNotesContract_JSONOutput(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	writeHistoryJSON(t, tempDir, `[
  {
    "version": "1.0.1",
    "package": "core",
    "tag": "core/v1.0.1",
    "timestamp": "2026-01-30T10:00:00Z",
    "consignments": [
      {
        "id": "c1",
        "summary": "Fixed critical bug",
        "changeType": "patch"
      }
    ]
  }
]`)

	cmd := exec.Command(shipyardBin, "--json", "release-notes", "--package", "core")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "release-notes --json should exit 0: %s", string(output))
	outputStr := string(output)
	assert.True(t,
		strings.Contains(outputStr, "{") || strings.Contains(outputStr, "["),
		"JSON output should contain { or [")
	assert.True(t,
		strings.Contains(outputStr, `"version"`) || strings.Contains(outputStr, `"1.0.1"`),
		"JSON output should contain version info")
}

// TestReleaseNotesContract_OutputToFile tests writing release notes to a file
func TestReleaseNotesContract_OutputToFile(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	writeHistoryJSON(t, tempDir, `[
  {
    "version": "1.0.1",
    "package": "core",
    "tag": "core/v1.0.1",
    "timestamp": "2026-01-30T10:00:00Z",
    "consignments": [
      {
        "id": "c1",
        "summary": "Fixed critical bug",
        "changeType": "patch"
      }
    ]
  }
]`)

	outputFile := filepath.Join(t.TempDir(), "notes.md")

	cmd := exec.Command(shipyardBin, "release-notes", "--package", "core", "--output", outputFile)
	cmd.Dir = tempDir
	cmdOutput, err := cmd.CombinedOutput()

	require.NoError(t, err, "release-notes --output should exit 0: %s", string(cmdOutput))

	// Verify file exists and contains expected content
	_, statErr := os.Stat(outputFile)
	require.NoError(t, statErr, "Output file should exist")

	content := readFileContent(t, outputFile)
	assert.Contains(t, content, "1.0.1", "Output file should contain version")
}

// TestReleaseNotesContract_HelpFlag tests that --help shows expected flags
func TestReleaseNotesContract_HelpFlag(t *testing.T) {
	shipyardBin := buildShipyard(t)

	cmd := exec.Command(shipyardBin, "release-notes", "--help")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "release-notes --help should exit 0")
	outputStr := string(output)
	assert.Contains(t, outputStr, "Recount the journey from the captain", "Help should contain command description")
	assert.Contains(t, outputStr, "--package", "Help should mention --package flag")
	assert.Contains(t, outputStr, "--output", "Help should mention --output flag")
}

// TestReleaseNotesContract_MultiPackageRequiresFlag tests that multi-package repos require --package
func TestReleaseNotesContract_MultiPackageRequiresFlag(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepoMultiPackage(t, shipyardBin, tempDir)

	writeHistoryJSON(t, tempDir, `[
  {"version":"1.0.1","package":"core","tag":"core/v1.0.1","timestamp":"2026-01-30T10:00:00Z","consignments":[{"id":"c1","summary":"Core fix","changeType":"patch"}]},
  {"version":"1.1.0","package":"api","tag":"api/v1.1.0","timestamp":"2026-01-30T10:00:00Z","consignments":[{"id":"c2","summary":"API feature","changeType":"minor"}]}
]`)

	cmd := exec.Command(shipyardBin, "release-notes")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	assert.Error(t, err, "release-notes without --package should fail for multi-package repos")
	outputStr := strings.ToLower(string(output))
	assert.True(t,
		strings.Contains(outputStr, "--package") ||
			strings.Contains(outputStr, "package is required"),
		"Output should mention --package or that package is required, got: %s", string(output))
}
