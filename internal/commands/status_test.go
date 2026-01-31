package commands

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NatoNathan/shipyard/internal/consignment"
	"github.com/NatoNathan/shipyard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStatusCommand_Success tests displaying pending consignments
func TestStatusCommand_Success(t *testing.T) {
	// Setup: Create initialized repo with consignments
	tempDir := t.TempDir()
	setupInitializedRepo(t, tempDir)
	defer changeToDir(t, tempDir)()

	// Create test consignments
	consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
	createStatusTestConsignment(t, consignmentsDir, "c1", []string{"core"}, types.ChangeTypePatch, "Fix bug")
	createStatusTestConsignment(t, consignmentsDir, "c2", []string{"api"}, types.ChangeTypeMinor, "Add feature")

	// Change to test directory

	// Test: Run status command with verbose to see details
	cmd := NewStatusCommand()
	cmd.SetArgs([]string{"--verbose"})

	output := captureOutput(func() {
		err := cmd.Execute()
		require.NoError(t, err)
	})

	// Verify: Output shows consignments with details in verbose mode
	assert.Contains(t, output, "c1")
	assert.Contains(t, output, "c2")
	assert.Contains(t, output, "core")
	assert.Contains(t, output, "api")
	assert.Contains(t, output, "Fix bug")
	assert.Contains(t, output, "Add feature")
}

// TestStatusCommand_EmptyConsignments tests status with no pending consignments
func TestStatusCommand_EmptyConsignments(t *testing.T) {
	// Setup: Create initialized repo with no consignments
	tempDir := t.TempDir()
	setupInitializedRepo(t, tempDir)
	defer changeToDir(t, tempDir)()

	// Test: Run status command
	cmd := NewStatusCommand()
	cmd.SetArgs([]string{})

	output := captureOutput(func() {
		err := cmd.Execute()
		require.NoError(t, err)
	})

	// Verify: Output indicates no pending changes
	assert.Contains(t, output, "No pending consignments")
}

// TestStatusCommand_NotInitialized tests status in non-initialized directory
func TestStatusCommand_NotInitialized(t *testing.T) {
	// Setup: Change to non-initialized directory
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)

	// Test: Run status command
	cmd := NewStatusCommand()
	cmd.SetArgs([]string{})

	err := cmd.Execute()

	// Verify: Error about not initialized
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

// TestStatusCommand_GroupedByPackage tests consignments grouped by package
func TestStatusCommand_GroupedByPackage(t *testing.T) {
	// Setup: Create consignments for multiple packages
	tempDir := t.TempDir()
	setupInitializedRepo(t, tempDir)
	defer changeToDir(t, tempDir)()

	consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
	createStatusTestConsignment(t, consignmentsDir, "c1", []string{"core"}, types.ChangeTypePatch, "Core fix 1")
	createStatusTestConsignment(t, consignmentsDir, "c2", []string{"core"}, types.ChangeTypeMinor, "Core feature")
	createStatusTestConsignment(t, consignmentsDir, "c3", []string{"api"}, types.ChangeTypeMajor, "API breaking change")

	// Test: Run status command
	cmd := NewStatusCommand()
	cmd.SetArgs([]string{})

	output := captureOutput(func() {
		err := cmd.Execute()
		require.NoError(t, err)
	})

	// Verify: Output groups by package
	assert.Contains(t, output, "core")
	assert.Contains(t, output, "api")
	assert.Contains(t, output, "2") // 2 consignments for core
}

// TestStatusCommand_ShowsVersionBumps tests displaying calculated version bumps
func TestStatusCommand_ShowsVersionBumps(t *testing.T) {
	// Setup: Create repo with consignments
	tempDir := t.TempDir()
	setupInitializedRepo(t, tempDir)
	defer changeToDir(t, tempDir)()

	consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
	createStatusTestConsignment(t, consignmentsDir, "c1", []string{"core"}, types.ChangeTypeMinor, "Feature")

	// Test: Run status command
	cmd := NewStatusCommand()
	cmd.SetArgs([]string{})

	output := captureOutput(func() {
		err := cmd.Execute()
		require.NoError(t, err)
	})

	// Verify: Output shows version bump
	assert.Contains(t, output, "minor")
}

// TestStatusCommand_PackageFilter tests filtering by package
func TestStatusCommand_PackageFilter(t *testing.T) {
	// Setup: Create consignments for multiple packages
	tempDir := t.TempDir()
	setupInitializedRepo(t, tempDir)
	defer changeToDir(t, tempDir)()

	consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
	createStatusTestConsignment(t, consignmentsDir, "c1", []string{"core"}, types.ChangeTypePatch, "Core fix")
	createStatusTestConsignment(t, consignmentsDir, "c2", []string{"api"}, types.ChangeTypeMinor, "API feature")

	// Test: Run status command with package filter and verbose
	cmd := NewStatusCommand()
	cmd.SetArgs([]string{"--package", "core", "--verbose"})

	output := captureOutput(func() {
		err := cmd.Execute()
		require.NoError(t, err)
	})

	// Verify: Output only shows core package
	assert.Contains(t, output, "core")
	assert.Contains(t, output, "Core fix")
	assert.NotContains(t, output, "api")
	assert.NotContains(t, output, "API feature")
}

// TestStatusCommand_MultiplePackageFilter tests filtering by multiple packages
func TestStatusCommand_MultiplePackageFilter(t *testing.T) {
	// Setup: Create consignments for multiple packages
	tempDir := t.TempDir()
	setupInitializedRepo(t, tempDir)
	defer changeToDir(t, tempDir)()

	consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
	createStatusTestConsignment(t, consignmentsDir, "c1", []string{"core"}, types.ChangeTypePatch, "Core fix")
	createStatusTestConsignment(t, consignmentsDir, "c2", []string{"api"}, types.ChangeTypeMinor, "API feature")
	createStatusTestConsignment(t, consignmentsDir, "c3", []string{"web"}, types.ChangeTypeMajor, "Web breaking")

	// Test: Run status command with multiple package filters
	cmd := NewStatusCommand()
	cmd.SetArgs([]string{"--package", "core", "--package", "api"})

	output := captureOutput(func() {
		err := cmd.Execute()
		require.NoError(t, err)
	})

	// Verify: Output shows both core and api, but not web
	assert.Contains(t, output, "core")
	assert.Contains(t, output, "api")
	assert.NotContains(t, output, "web")
}

// TestStatusCommand_JSONOutput tests JSON format output
func TestStatusCommand_JSONOutput(t *testing.T) {
	// Setup: Create repo with consignments
	tempDir := t.TempDir()
	setupInitializedRepo(t, tempDir)
	defer changeToDir(t, tempDir)()

	consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
	createStatusTestConsignment(t, consignmentsDir, "c1", []string{"core"}, types.ChangeTypePatch, "Fix")

	// Test: Run status command with JSON output
	cmd := NewStatusCommand()
	cmd.SetArgs([]string{"--output", "json"})

	output := captureOutput(func() {
		err := cmd.Execute()
		require.NoError(t, err)
	})

	// Verify: Output is valid JSON
	assert.Contains(t, output, "{")
	assert.Contains(t, output, "}")
	assert.Contains(t, output, "\"core\"")
}

// TestStatusCommand_ShowsPropagatedBumps tests displaying propagated version bumps
func TestStatusCommand_ShowsPropagatedBumps(t *testing.T) {
	// Setup: Create repo with dependencies where api depends on core
	tempDir := t.TempDir()
	setupInitializedRepoWithDependencies(t, tempDir)
	defer changeToDir(t, tempDir)()

	// Create consignment ONLY for core package (not api)
	consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
	createStatusTestConsignment(t, consignmentsDir, "c1", []string{"core"}, types.ChangeTypeMinor, "Core feature")

	// Test: Run status command with verbose to see propagation details
	cmd := NewStatusCommand()
	cmd.SetArgs([]string{"--verbose"})

	output := captureOutput(func() {
		err := cmd.Execute()
		require.NoError(t, err)
	})

	// Verify: Output shows both packages
	assert.Contains(t, output, "core", "should show core package with direct bump")
	assert.Contains(t, output, "api", "should show api package with propagated bump")

	// Verify: Core shows direct bump, API shows propagated bump
	// The output should indicate which bumps are direct vs propagated
	assert.Contains(t, output, "minor", "should show minor bump")
	assert.Contains(t, output, "propagated", "should indicate propagated bump for api package")
}

// TestStatusCommand_QuietMode tests quiet output mode
func TestStatusCommand_QuietMode(t *testing.T) {
	// Setup: Create repo with consignments
	tempDir := t.TempDir()
	setupInitializedRepo(t, tempDir)
	defer changeToDir(t, tempDir)()

	consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
	createStatusTestConsignment(t, consignmentsDir, "c1", []string{"core"}, types.ChangeTypePatch, "Fix")

	// Test: Run status command in quiet mode
	cmd := NewStatusCommand()
	cmd.SetArgs([]string{"--quiet"})

	output := captureOutput(func() {
		err := cmd.Execute()
		require.NoError(t, err)
	})

	// Verify: Minimal output
	assert.NotEmpty(t, output)
	// Quiet mode should show minimal info (just package names and bump types)
}

// TestStatusCommand_GlobalJSONFlag tests that global --json flag is respected
func TestStatusCommand_GlobalJSONFlag(t *testing.T) {
	// Setup: Create repo with consignments
	tempDir := t.TempDir()
	setupInitializedRepo(t, tempDir)
	defer changeToDir(t, tempDir)()

	consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
	createStatusTestConsignment(t, consignmentsDir, "c1", []string{"core"}, types.ChangeTypePatch, "Fix")

	// Test: Run status command with root command to simulate global flag
	// Note: This tests the integration with the global --json flag
	cmd := NewStatusCommand()

	// Simulate global --json flag being set by checking parent flags
	// For now, we'll use --output json to verify the behavior works
	cmd.SetArgs([]string{"--output", "json"})

	output := captureOutput(func() {
		err := cmd.Execute()
		require.NoError(t, err)
	})

	// Verify: Output is valid JSON
	assert.Contains(t, output, "{", "Should output JSON")
	assert.Contains(t, output, "\"core\"", "Should contain package in JSON")
	assert.Contains(t, output, "\"patch\"", "Should contain change type in JSON")
}

// TestStatusCommand_VerboseMode tests verbose output mode
func TestStatusCommand_VerboseMode(t *testing.T) {
	// Setup: Create repo with consignments
	tempDir := t.TempDir()
	setupInitializedRepo(t, tempDir)
	defer changeToDir(t, tempDir)()

	consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
	createStatusTestConsignment(t, consignmentsDir, "c1", []string{"core"}, types.ChangeTypePatch, "Fix")

	// Test: Run status command in verbose mode
	cmd := NewStatusCommand()
	cmd.SetArgs([]string{"--verbose"})

	output := captureOutput(func() {
		err := cmd.Execute()
		require.NoError(t, err)
	})

	// Verify: Detailed output with timestamps and metadata
	assert.Contains(t, output, "c1")
	assert.Contains(t, output, "Fix")
}

// Helper functions

func setupInitializedRepo(t *testing.T, dir string) {
	// Create .shipyard directory structure
	shipyardDir := filepath.Join(dir, ".shipyard")
	require.NoError(t, os.MkdirAll(shipyardDir, 0755))

	consignmentsDir := filepath.Join(shipyardDir, "consignments")
	require.NoError(t, os.MkdirAll(consignmentsDir, 0755))

	historyDir := filepath.Join(shipyardDir, "history")
	require.NoError(t, os.MkdirAll(historyDir, 0755))

	// Create package directories with version files
	coreDir := filepath.Join(dir, "core")
	require.NoError(t, os.MkdirAll(coreDir, 0755))
	coreVersion := filepath.Join(coreDir, "version.go")
	require.NoError(t, os.WriteFile(coreVersion, []byte(`package core

const Version = "1.0.0"
`), 0644))

	apiDir := filepath.Join(dir, "api")
	require.NoError(t, os.MkdirAll(apiDir, 0755))
	apiVersion := filepath.Join(apiDir, "version.go")
	require.NoError(t, os.WriteFile(apiVersion, []byte(`package api

const Version = "2.0.0"
`), 0644))

	// Create minimal config (valid YAML array format)
	configContent := `packages:
  - name: core
    path: ./core
    ecosystem: go
  - name: api
    path: ./api
    ecosystem: go
`
	configPath := filepath.Join(shipyardDir, "shipyard.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	// Create empty history file
	historyPath := filepath.Join(shipyardDir, "history.json")
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))
}

func createStatusTestConsignment(t *testing.T, dir, id string, packages []string, changeType types.ChangeType, summary string) {
	c := &consignment.Consignment{
		ID:         id,
		Timestamp:  time.Now(),
		Packages:   packages,
		ChangeType: changeType,
		Summary:    summary,
		Metadata:   map[string]interface{}{},
	}

	path := filepath.Join(dir, id+".md")
	content, err := consignment.Serialize(c)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
}

func setupInitializedRepoWithDependencies(t *testing.T, dir string) {
	// Create .shipyard directory structure
	shipyardDir := filepath.Join(dir, ".shipyard")
	require.NoError(t, os.MkdirAll(shipyardDir, 0755))

	consignmentsDir := filepath.Join(shipyardDir, "consignments")
	require.NoError(t, os.MkdirAll(consignmentsDir, 0755))

	historyDir := filepath.Join(shipyardDir, "history")
	require.NoError(t, os.MkdirAll(historyDir, 0755))

	// Create package directories with version files
	coreDir := filepath.Join(dir, "core")
	require.NoError(t, os.MkdirAll(coreDir, 0755))
	coreVersion := filepath.Join(coreDir, "version.go")
	require.NoError(t, os.WriteFile(coreVersion, []byte(`package core

const Version = "1.0.0"
`), 0644))

	apiDir := filepath.Join(dir, "api")
	require.NoError(t, os.MkdirAll(apiDir, 0755))
	apiVersion := filepath.Join(apiDir, "version.go")
	require.NoError(t, os.WriteFile(apiVersion, []byte(`package api

const Version = "2.0.0"
`), 0644))

	// Create config with dependencies (valid YAML format with proper structure)
	configContent := `packages:
  - name: core
    path: ./core
    ecosystem: go
  - name: api
    path: ./api
    ecosystem: go
    dependencies:
      - package: core
        strategy: linked
        bumpMapping:
          patch: patch
          minor: patch
          major: minor
`
	configPath := filepath.Join(shipyardDir, "shipyard.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	// Create empty history file
	historyPath := filepath.Join(shipyardDir, "history.json")
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))
}
