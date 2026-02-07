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

// TestInitContract_SuccessfulInit tests the contract for successful initialization
func TestInitContract_SuccessfulInit(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Run shipyard init with --yes flag (non-interactive mode for testing)
	cmd := exec.Command(shipyardBin, "init", "--yes")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify exit code
	require.NoError(t, err, "Command should exit with code 0")

	// Verify output contains success message
	outputStr := string(output)
	assert.Contains(t, outputStr, "Shipyard initialized successfully", "Output should contain success message")
	assert.Contains(t, outputStr, "Configuration:", "Output should mention configuration file")
	assert.Contains(t, outputStr, "Consignments directory:", "Output should mention consignments directory")
	assert.Contains(t, outputStr, "History file:", "Output should mention history file")

	// Verify files were created
	configPath := filepath.Join(tempDir, ".shipyard", "shipyard.yaml")
	assert.FileExists(t, configPath, "Configuration file should be created")

	consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
	assert.DirExists(t, consignmentsDir, "Consignments directory should be created")

	historyPath := filepath.Join(tempDir, ".shipyard", "history.json")
	assert.FileExists(t, historyPath, "History file should be created")
}

// TestInitContract_AlreadyInitialized tests the contract when already initialized
func TestInitContract_AlreadyInitialized(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Initialize once
	cmd := exec.Command(shipyardBin, "init", "--yes")
	cmd.Dir = tempDir
	_, err := cmd.CombinedOutput()
	require.NoError(t, err, "First init should succeed")

	// Try to initialize again
	cmd = exec.Command(shipyardBin, "init", "--yes")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify exit code (should fail)
	assert.Error(t, err, "Command should exit with non-zero code")

	// Verify output contains error message
	outputStr := string(output)
	assert.Contains(t, outputStr, "already initialized", "Output should indicate already initialized")
}

// TestInitContract_ForceReinitialize tests the contract for --force flag
func TestInitContract_ForceReinitialize(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Initialize once
	cmd := exec.Command(shipyardBin, "init", "--yes")
	cmd.Dir = tempDir
	_, err := cmd.CombinedOutput()
	require.NoError(t, err, "First init should succeed")

	// Force reinitialize
	cmd = exec.Command(shipyardBin, "init", "--force", "--yes")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify exit code (should succeed)
	require.NoError(t, err, "Force init should succeed")

	// Verify output contains success message
	outputStr := string(output)
	assert.Contains(t, outputStr, "Shipyard initialized successfully", "Output should contain success message")
}

// TestInitContract_NotGitRepository tests the contract when not in git repository
func TestInitContract_NotGitRepository(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()

	// Don't initialize git

	// Run shipyard init
	cmd := exec.Command(shipyardBin, "init", "--yes")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify exit code (should fail)
	assert.Error(t, err, "Command should exit with non-zero code")

	// Verify output contains error message
	outputStr := string(output)
	assert.Contains(t, outputStr, "git repository", "Output should mention git repository requirement")
}

// TestInitContract_HelpFlag tests the contract for --help flag
func TestInitContract_HelpFlag(t *testing.T) {
	shipyardBin := buildShipyard(t)

	// Run shipyard init --help
	cmd := exec.Command(shipyardBin, "init", "--help")
	output, err := cmd.CombinedOutput()

	// Verify exit code (should succeed)
	require.NoError(t, err, "Help command should succeed")

	// Verify output contains help information
	outputStr := string(output)
	assert.Contains(t, outputStr, "Prepare your repository", "Output should contain command description")
	assert.Contains(t, outputStr, "USAGE", "Output should contain usage section")
	assert.Contains(t, outputStr, "--force", "Output should mention --force flag")
	assert.Contains(t, outputStr, "--remote", "Output should mention --remote flag")
}

// TestInitContract_RemoteFlag tests the contract for --remote flag
func TestInitContract_RemoteFlag(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Create a mock remote config
	remoteConfigDir := t.TempDir()
	remoteConfigPath := filepath.Join(remoteConfigDir, "remote.yaml")
	remoteConfig := []byte(`templates:
  changelog:
    source: "builtin:default"
`)
	require.NoError(t, os.WriteFile(remoteConfigPath, remoteConfig, 0644))

	// Run shipyard init with --remote
	cmd := exec.Command(shipyardBin, "init", "--remote", "file://"+remoteConfigPath, "--yes")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify exit code (should succeed)
	require.NoError(t, err, "Init with remote should succeed")

	// Verify output contains success message
	outputStr := string(output)
	assert.Contains(t, outputStr, "Shipyard initialized successfully", "Output should contain success message")

	// Verify config contains extends section
	configPath := filepath.Join(tempDir, ".shipyard", "shipyard.yaml")
	configContent, err := os.ReadFile(configPath)
	require.NoError(t, err, "Should be able to read config")
	assert.Contains(t, string(configContent), "extends:", "Config should contain extends section")
}

// TestInitContract_QuietFlag tests the contract for --quiet flag
func TestInitContract_QuietFlag(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Run shipyard init with --quiet
	cmd := exec.Command(shipyardBin, "init", "--quiet", "--yes")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify exit code (should succeed)
	require.NoError(t, err, "Init with --quiet should succeed")

	// Verify output is minimal (no INFO logs)
	outputStr := string(output)
	assert.NotContains(t, outputStr, "[INFO]", "Quiet mode should suppress INFO logs")

	// But files should still be created
	configPath := filepath.Join(tempDir, ".shipyard", "shipyard.yaml")
	assert.FileExists(t, configPath, "Configuration file should be created")
}

// TestInitContract_VerboseFlag tests the contract for --verbose flag
func TestInitContract_VerboseFlag(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Run shipyard init with --verbose
	cmd := exec.Command(shipyardBin, "init", "--verbose", "--yes")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()

	// Verify exit code (should succeed)
	require.NoError(t, err, "Init with --verbose should succeed")

	// Verify output contains detailed information
	outputStr := string(output)
	assert.Contains(t, outputStr, "Shipyard initialized successfully", "Output should contain success message")

	// Verbose mode should show more details
	// Note: This depends on how verbose logging is implemented
	// For now, we just verify it doesn't break anything
}

// TestInitContract_OutputFormat tests the contract for consistent output format
func TestInitContract_OutputFormat(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Run shipyard init
	cmd := exec.Command(shipyardBin, "init", "--yes")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Init should succeed")

	outputStr := string(output)
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")

	// Verify output has structured format
	// Should contain log level indicators
	hasInfoLogs := false
	for _, line := range lines {
		if strings.Contains(line, "[INFO]") {
			hasInfoLogs = true
			break
		}
	}
	assert.True(t, hasInfoLogs, "Output should contain INFO log messages")

	// Verify success indicator
	assert.Contains(t, outputStr, "✓", "Output should contain success indicator")
}

// TestInitContract_ConfigFileFormat tests the contract for generated config file
func TestInitContract_ConfigFileFormat(t *testing.T) {
	shipyardBin := buildShipyard(t)
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Create a Go module for detection
	goModPath := filepath.Join(tempDir, "go.mod")
	require.NoError(t, os.WriteFile(goModPath, []byte("module github.com/test/project\n\ngo 1.21\n"), 0644))

	// Run shipyard init
	cmd := exec.Command(shipyardBin, "init", "--yes")
	cmd.Dir = tempDir
	_, err := cmd.CombinedOutput()
	require.NoError(t, err, "Init should succeed")

	// Read and verify config file structure
	configPath := filepath.Join(tempDir, ".shipyard", "shipyard.yaml")
	configContent, err := os.ReadFile(configPath)
	require.NoError(t, err, "Should be able to read config")

	configStr := string(configContent)

	// Verify required sections
	assert.Contains(t, configStr, "packages:", "Config should contain packages section")
	assert.Contains(t, configStr, "name:", "Config should contain package name")
	assert.Contains(t, configStr, "path:", "Config should contain package path")
	assert.Contains(t, configStr, "ecosystem:", "Config should contain package ecosystem")
	assert.Contains(t, configStr, "templates:", "Config should contain templates section")
	assert.Contains(t, configStr, "changelog:", "Config should contain changelog template")
	assert.Contains(t, configStr, "consignments:", "Config should contain consignments section")
	assert.Contains(t, configStr, "history:", "Config should contain history section")

	// Verify YAML format (basic check)
	assert.True(t, strings.HasPrefix(configStr, "packages:") || strings.HasPrefix(configStr, "extends:"),
		"Config should start with a top-level key")
}
