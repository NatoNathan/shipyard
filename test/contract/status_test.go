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

// TestStatusCommand_ExitCodes tests that status command returns correct exit codes
func TestStatusCommand_ExitCodes(t *testing.T) {
	// Build the shipyard binary for testing
	shipyardBin := buildShipyard(t)
	defer os.Remove(shipyardBin)

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
		assert.Contains(t, string(output), "patch")
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
	defer os.Remove(shipyardBin)

	t.Run("table format by default", func(t *testing.T) {
		// Setup
		tempDir := t.TempDir()
		initializeTestRepo(t, shipyardBin, tempDir)
		createTestConsignment(t, tempDir, "c1", "core", "minor", "Add feature")

		// Test: Run without --output flag
		cmd := exec.Command(shipyardBin, "status")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()

		// Verify: Table format
		require.NoError(t, err, "command failed with output: %s", string(output))
		assert.Contains(t, string(output), "Package:")
		assert.Contains(t, string(output), "Bump:")
		assert.Contains(t, string(output), "Consignments:")
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
	defer os.Remove(shipyardBin)

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
		assert.Contains(t, string(output), "Package:", "Should output table format when local flag specified")
		assert.NotContains(t, string(output), "{", "Should not be JSON when local flag overrides")
	})
}

func TestStatusCommand_PackageFiltering(t *testing.T) {
	shipyardBin := buildShipyard(t)
	defer os.Remove(shipyardBin)

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

// Helper functions

func initializeTestRepo(t *testing.T, shipyardBin, dir string) {
	t.Helper()

	// Initialize git repo using existing helper
	initGitRepo(t, dir)

	// Create a "core" package directory
	coreDir := filepath.Join(dir, "core")
	require.NoError(t, os.MkdirAll(coreDir, 0755))

	// Create go.mod for package detection
	goModContent := `module example.com/core
go 1.21
`
	require.NoError(t, os.WriteFile(filepath.Join(coreDir, "go.mod"), []byte(goModContent), 0644))

	// Create version file
	versionContent := `package core

const Version = "1.0.0"
`
	require.NoError(t, os.WriteFile(filepath.Join(coreDir, "version.go"), []byte(versionContent), 0644))

	// Run shipyard init
	cmd := exec.Command(shipyardBin, "init", "--yes")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "failed to init shipyard: %s", string(output))
}

func initializeTestRepoMultiPackage(t *testing.T, shipyardBin, dir string) {
	t.Helper()

	// Initialize git repo using existing helper
	initGitRepo(t, dir)

	// Create packages
	packages := []struct {
		name string
		path string
	}{
		{"core", "core"},
		{"api", "api"},
		{"web", "web"},
	}

	for _, pkg := range packages {
		pkgDir := filepath.Join(dir, pkg.path)
		require.NoError(t, os.MkdirAll(pkgDir, 0755))

		versionContent := `package ` + pkg.name + `

const Version = "1.0.0"
`
		require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "version.go"), []byte(versionContent), 0644))

		goModContent := `module example.com/` + pkg.name + `
go 1.21
`
		require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte(goModContent), 0644))
	}

	// Run shipyard init
	cmd := exec.Command(shipyardBin, "init", "--yes")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "failed to init shipyard: %s", string(output))
}

func createTestConsignment(t *testing.T, dir, id, packageName, changeType, summary string) {
	t.Helper()

	content := `---
id: ` + id + `
packages:
  - ` + packageName + `
changeType: ` + changeType + `
summary: ` + summary + `
timestamp: 2026-01-30T00:00:00Z
---
# Change

` + summary + `
`

	consignmentPath := filepath.Join(dir, ".shipyard", "consignments", id+".md")
	require.NoError(t, os.WriteFile(consignmentPath, []byte(content), 0644))
}
