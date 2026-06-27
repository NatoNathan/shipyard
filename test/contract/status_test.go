package contract

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStatusCommand_ExitCodes tests that status command returns correct exit codes
func TestStatusCommand_ExitCodes(t *testing.T) {
	// Build the shipyard binary for testing
	shipyardBin := buildShipyard(t)

	t.Run("exit 0 with no consignments", func(t *testing.T) {
		// Setup: Create initialized repo
		tempDir := t.TempDir()
		initializeTestRepo(t, shipyardBin, tempDir)

		// Test: Run status command
		cmd := exec.Command(shipyardBin, "status")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()

		// Verify: Exit code 0
		require.NoError(t, err, "status should exit 0 with no consignments")
		assert.Contains(t, string(output), "No pending consignments")
	})

	t.Run("exit 0 with pending consignments", func(t *testing.T) {
		// Setup: Create initialized repo with consignment
		tempDir := t.TempDir()
		initializeTestRepo(t, shipyardBin, tempDir)
		createTestConsignment(t, tempDir, "test-001", "core", "patch", "Fix bug")

		// Test: Run status command
		cmd := exec.Command(shipyardBin, "status")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()

		// Verify: Exit code 0
		require.NoError(t, err, "status should exit 0 with consignments")
		assert.Contains(t, string(output), "core")
		assert.Contains(t, string(output), "1.0.0")
		assert.Contains(t, string(output), "1.0.1")
	})

	t.Run("clean clone with missing consignments returns empty JSON", func(t *testing.T) {
		tempDir := t.TempDir()
		initializeTestRepo(t, shipyardBin, tempDir)
		require.NoError(t, os.RemoveAll(filepath.Join(tempDir, ".shipyard", "consignments")))

		cmd := exec.Command(shipyardBin, "--json", "status")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()

		require.NoError(t, err, "status should exit 0 with a missing consignments directory: %s", output)
		assert.True(t, json.Valid(output), "status should return valid JSON: %s", output)
		assert.JSONEq(t, `{}`, string(output))
	})

	t.Run("uses configured consignments path", func(t *testing.T) {
		tempDir := t.TempDir()
		initializeTestRepo(t, shipyardBin, tempDir)
		createTestConsignment(t, tempDir, "custom-1", "core", "patch", "Custom path fix")

		customDir := filepath.Join(tempDir, "changes", "pending")
		require.NoError(t, os.MkdirAll(customDir, 0755))
		require.NoError(t, os.Rename(
			filepath.Join(tempDir, ".shipyard", "consignments", "custom-1.md"),
			filepath.Join(customDir, "custom-1.md"),
		))
		require.NoError(t, os.RemoveAll(filepath.Join(tempDir, ".shipyard", "consignments")))
		writeConfig(t, tempDir, `packages:
  - name: core
    path: ./core
    ecosystem: go
consignments:
  path: changes/pending
history:
  path: .shipyard/history.json
`)

		cmd := exec.Command(shipyardBin, "status", "--verbose")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()

		require.NoError(t, err, "status should honor configured consignments path: %s", output)
		assert.Contains(t, string(output), "Custom path fix")
	})

	t.Run("invalid consignments path exits 1", func(t *testing.T) {
		tempDir := t.TempDir()
		initializeTestRepo(t, shipyardBin, tempDir)
		invalidPath := filepath.Join(tempDir, "consignments-file")
		require.NoError(t, os.WriteFile(invalidPath, []byte("not a directory"), 0644))
		writeConfig(t, tempDir, `packages:
  - name: core
    path: ./core
    ecosystem: go
consignments:
  path: consignments-file
`)

		cmd := exec.Command(shipyardBin, "status")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()

		require.Error(t, err, "invalid consignments path should fail: %s", output)
		assert.Equal(t, 1, getExitCode(err))
	})

	t.Run("exit 1 when not initialized", func(t *testing.T) {
		// Setup: Non-initialized directory
		tempDir := t.TempDir()

		// Test: Run status command
		cmd := exec.Command(shipyardBin, "status")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()

		// Verify: Exit code 1
		require.Error(t, err, "status should exit 1 when not initialized")
		assert.Contains(t, string(output), "not initialized")
	})
}

// TestStatusCommand_OutputFormat tests output format consistency
func TestStatusCommand_OutputFormat(t *testing.T) {
	shipyardBin := buildShipyard(t)

	t.Run("table format by default", func(t *testing.T) {
		// Setup
		tempDir := t.TempDir()
		initializeTestRepo(t, shipyardBin, tempDir)
		createTestConsignment(t, tempDir, "c1", "core", "minor", "Add feature")

		// Test: Run without --output flag
		cmd := exec.Command(shipyardBin, "status")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()

		// Verify: Table format with column headers and data
		require.NoError(t, err, "command failed with output: %s", string(output))
		assert.Contains(t, string(output), "Pending consignments")
		assert.Contains(t, string(output), "core")
		assert.Contains(t, string(output), "Package")
		assert.Contains(t, string(output), "Current")
	})

	t.Run("json format with --output json", func(t *testing.T) {
		// Setup
		tempDir := t.TempDir()
		initializeTestRepo(t, shipyardBin, tempDir)
		createTestConsignment(t, tempDir, "c1", "core", "minor", "Add feature")

		// Test: Run with --output json
		cmd := exec.Command(shipyardBin, "status", "--output", "json")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()

		// Verify: Valid JSON
		require.NoError(t, err)
		assert.Contains(t, string(output), "{")
		assert.Contains(t, string(output), "\"core\"")
		assert.Contains(t, string(output), "\"bump\"")
		assert.Contains(t, string(output), "\"minor\"")
	})

	t.Run("quiet mode minimal output", func(t *testing.T) {
		// Setup
		tempDir := t.TempDir()
		initializeTestRepo(t, shipyardBin, tempDir)
		createTestConsignment(t, tempDir, "c1", "core", "patch", "Fix")

		// Test: Run with --quiet
		cmd := exec.Command(shipyardBin, "status", "--quiet")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()

		// Verify: Minimal output (just package: bump)
		require.NoError(t, err)
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		// Should be minimal - just package and bump info
		assert.LessOrEqual(t, len(lines), 3, "quiet mode should have minimal output")
		assert.Contains(t, string(output), "core")
		assert.Contains(t, string(output), "patch")
	})

	t.Run("verbose mode detailed output", func(t *testing.T) {
		// Setup
		tempDir := t.TempDir()
		initializeTestRepo(t, shipyardBin, tempDir)
		createTestConsignment(t, tempDir, "c1", "core", "minor", "Add feature")

		// Test: Run with --verbose
		cmd := exec.Command(shipyardBin, "status", "--verbose")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()

		// Verify: Detailed output
		require.NoError(t, err)
		assert.Contains(t, string(output), "core")
		assert.Contains(t, string(output), "minor")
		assert.Contains(t, string(output), "Add feature")
		assert.Contains(t, string(output), "c1") // Should show consignment ID in verbose
	})
}

// TestStatusCommand_PackageFiltering tests package filtering behavior
// TestStatusCommand_GlobalJSONFlag tests that global --json flag is respected
func TestStatusCommand_GlobalJSONFlag(t *testing.T) {
	shipyardBin := buildShipyard(t)

	t.Run("global --json flag produces JSON output", func(t *testing.T) {
		// Setup
		tempDir := t.TempDir()
		initializeTestRepo(t, shipyardBin, tempDir)
		createTestConsignment(t, tempDir, "c1", "core", "minor", "Add feature")

		// Test: Use global --json flag (before the subcommand)
		cmd := exec.Command(shipyardBin, "--json", "status")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()

		// Verify: JSON output
		require.NoError(t, err, "command failed with output: %s", string(output))
		assert.Contains(t, string(output), "{", "Should output JSON with global flag")
		assert.Contains(t, string(output), "\"core\"", "Should contain package in JSON")
		assert.Contains(t, string(output), "\"minor\"", "Should contain change type in JSON")
	})

	t.Run("local --output flag overrides global --json", func(t *testing.T) {
		// Setup
		tempDir := t.TempDir()
		initializeTestRepo(t, shipyardBin, tempDir)
		createTestConsignment(t, tempDir, "c1", "core", "patch", "Fix bug")

		// Test: Global --json with local --output table
		cmd := exec.Command(shipyardBin, "--json", "status", "--output", "table")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()

		// Verify: Table output (local flag wins)
		require.NoError(t, err)
		assert.Contains(t, string(output), "📦 Pending consignments", "Should output table format when local flag specified")
		assert.NotContains(t, string(output), "{", "Should not be JSON when local flag overrides")
	})
}

func TestStatusCommand_PackageFiltering(t *testing.T) {
	shipyardBin := buildShipyard(t)

	t.Run("filter single package", func(t *testing.T) {
		// Setup: Multiple packages
		tempDir := t.TempDir()
		initializeTestRepo(t, shipyardBin, tempDir)
		createTestConsignment(t, tempDir, "c1", "core", "patch", "Core fix")
		createTestConsignment(t, tempDir, "c2", "api", "minor", "API feature")

		// Test: Filter to core only
		cmd := exec.Command(shipyardBin, "status", "--package", "core")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()

		// Verify: Only core shown
		require.NoError(t, err)
		assert.Contains(t, string(output), "core")
		assert.NotContains(t, string(output), "api")
	})

	t.Run("filter multiple packages", func(t *testing.T) {
		// Setup: Three packages
		tempDir := t.TempDir()
		initializeTestRepoMultiPackage(t, shipyardBin, tempDir)
		createTestConsignment(t, tempDir, "c1", "core", "patch", "Core")
		createTestConsignment(t, tempDir, "c2", "api", "minor", "API")
		createTestConsignment(t, tempDir, "c3", "web", "major", "Web")

		// Test: Filter to core and api
		cmd := exec.Command(shipyardBin, "status", "--package", "core", "--package", "api")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()

		// Verify: Core and api shown, not web
		require.NoError(t, err)
		assert.Contains(t, string(output), "core")
		assert.Contains(t, string(output), "api")
		assert.NotContains(t, string(output), "web")
	})
}

// Helper functions are in helpers_test.go
