package contract

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigShowContract_DefaultYAMLOutput tests that config show outputs YAML by default
func TestConfigShowContract_DefaultYAMLOutput(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	cmd := exec.Command(shipyardBin, "config", "show")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "config show should exit 0: %s", string(output))
	outputStr := string(output)
	assert.Contains(t, outputStr, "packages:")
	assert.Contains(t, outputStr, "ecosystem:")
}

// TestConfigShowContract_JSONOutput tests that --json flag produces JSON output
func TestConfigShowContract_JSONOutput(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepo(t, shipyardBin, tempDir)

	cmd := exec.Command(shipyardBin, "--json", "config", "show")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "config show --json should exit 0: %s", string(output))
	outputStr := string(output)
	assert.Contains(t, outputStr, "{")
	assert.Contains(t, outputStr, "\"Packages\"")
}

// TestConfigShowContract_NotInitialized tests error when shipyard is not initialized
func TestConfigShowContract_NotInitialized(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()

	cmd := exec.Command(shipyardBin, "config", "show")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	assert.Error(t, err, "config show should fail when not initialized")
	outputStr := strings.ToLower(string(output))
	assert.True(t,
		strings.Contains(outputStr, "config") ||
			strings.Contains(outputStr, "load") ||
			strings.Contains(outputStr, "not initialized"),
		"Output should mention config, load, or not initialized; got: %s", string(output))
}

// TestConfigShowContract_HelpFlag tests that --help flag works
func TestConfigShowContract_HelpFlag(t *testing.T) {
	shipyardBin := buildShipyard(t)

	cmd := exec.Command(shipyardBin, "config", "show", "--help")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "config show --help should exit 0: %s", string(output))
	outputStr := strings.ToLower(string(output))
	assert.True(t,
		strings.Contains(outputStr, "display") ||
			strings.Contains(outputStr, "configuration"),
		"Help output should mention display or configuration; got: %s", string(output))
}

// TestConfigShowContract_MultiPackage tests output includes all packages in a multi-package repo
func TestConfigShowContract_MultiPackage(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initializeTestRepoMultiPackage(t, shipyardBin, tempDir)

	cmd := exec.Command(shipyardBin, "config", "show")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "config show should exit 0: %s", string(output))
	outputStr := string(output)
	assert.Contains(t, outputStr, "core")
	assert.Contains(t, outputStr, "api")
	assert.Contains(t, outputStr, "web")
}
