package contract

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateContract_ValidConfig tests that validate succeeds on a properly initialized repo
func TestValidateContract_ValidConfig(t *testing.T) {
	shipyardBin := buildShipyard(t)

	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	cmd := exec.Command(shipyardBin, "validate")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "validate should exit 0 on valid config")
	outputLower := strings.ToLower(string(output))
	assert.True(t,
		strings.Contains(outputLower, "validation passed") ||
			strings.Contains(outputLower, "passed"),
		"Output should indicate validation passed, got: %s", string(output))
}

// TestValidateContract_NotInitialized tests that validate fails when shipyard is not initialized
func TestValidateContract_NotInitialized(t *testing.T) {
	shipyardBin := buildShipyard(t)

	tempDir := t.TempDir()

	cmd := exec.Command(shipyardBin, "validate")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	assert.Error(t, err, "validate should exit non-zero when not initialized")
	outputLower := strings.ToLower(string(output))
	assert.True(t,
		strings.Contains(outputLower, "config") ||
			strings.Contains(outputLower, "load") ||
			strings.Contains(outputLower, "not initialized"),
		"Output should mention config/load/not initialized issue, got: %s", string(output))
}

// TestValidateContract_JSONOutput_Valid tests that --json flag produces JSON output
func TestValidateContract_JSONOutput_Valid(t *testing.T) {
	shipyardBin := buildShipyard(t)

	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	cmd := exec.Command(shipyardBin, "--json", "validate")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "validate with --json should exit 0 on valid config")
	outputStr := string(output)
	assert.Contains(t, outputStr, `"valid"`, "JSON output should contain \"valid\" key")
	assert.Contains(t, outputStr, "true", "JSON output should contain true for valid config")
}

// TestValidateContract_QuietMode tests that --quiet produces minimal output
func TestValidateContract_QuietMode(t *testing.T) {
	shipyardBin := buildShipyard(t)

	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	cmd := exec.Command(shipyardBin, "--quiet", "validate")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "validate with --quiet should exit 0 on valid config")
	trimmed := strings.TrimSpace(string(output))
	assert.LessOrEqual(t, len(trimmed), 10,
		"Quiet mode should produce empty or very minimal output, got: %q", trimmed)
}

// TestValidateContract_HelpFlag tests that --help shows usage information
func TestValidateContract_HelpFlag(t *testing.T) {
	shipyardBin := buildShipyard(t)

	cmd := exec.Command(shipyardBin, "validate", "--help")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "validate --help should exit 0")
	outputStr := string(output)
	assert.True(t,
		strings.Contains(outputStr, "Validate configuration") ||
			strings.Contains(outputStr, "Validate") ||
			strings.Contains(outputStr, "validate"),
		"Help output should mention Validate, got: %s", outputStr)
}

// TestValidateContract_WithConsignments tests that validate succeeds when consignments exist
func TestValidateContract_WithConsignments(t *testing.T) {
	shipyardBin := buildShipyard(t)

	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)
	createTestConsignment(t, tempDir, "c1", "core", "patch", "Fix")

	cmd := exec.Command(shipyardBin, "validate")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "validate should exit 0 with consignments present")
	outputLower := strings.ToLower(string(output))
	assert.True(t,
		strings.Contains(outputLower, "validation passed") ||
			strings.Contains(outputLower, "passed"),
		"Output should indicate validation passed, got: %s", string(output))
}
