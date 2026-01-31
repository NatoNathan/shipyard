package commands

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAddCommand_NewConsignment tests creating a new consignment interactively
func TestAddCommand_NewConsignment(t *testing.T) {
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)
	initShipyardConfig(t, tempDir)

	// Run add command with options
	err := runAdd(tempDir, AddOptions{
		Packages:  []string{"core"},
		Type:      "minor",
		Summary:   "Added new feature",
		Metadata:  map[string]string{"author": "test@example.com"},
		Timestamp: time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC),
	})
	require.NoError(t, err, "Add command should succeed")

	// Verify consignment file was created
	consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
	entries, err := os.ReadDir(consignmentsDir)
	require.NoError(t, err, "Should be able to read consignments directory")
	assert.Equal(t, 1, len(entries), "Should have created one consignment file")
	assert.True(t, filepath.Ext(entries[0].Name()) == ".md", "Consignment should be markdown file")
}

// TestAddCommand_NonInteractive tests creating a consignment with flags
func TestAddCommand_NonInteractive(t *testing.T) {
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)
	initShipyardConfig(t, tempDir)

	// Run add command with all flags
	err := runAdd(tempDir, AddOptions{
		Packages:  []string{"core", "api"},
		Type:      "major",
		Summary:   "Breaking change",
		Metadata:  map[string]string{"issue": "JIRA-123"},
		Timestamp: time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC),
	})
	require.NoError(t, err, "Add command should succeed in non-interactive mode")

	// Verify consignment file was created
	consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
	entries, err := os.ReadDir(consignmentsDir)
	require.NoError(t, err, "Should be able to read consignments directory")
	assert.Equal(t, 1, len(entries), "Should have created one consignment file")
}

// TestAddCommand_InvalidPackage tests handling of invalid package names
func TestAddCommand_InvalidPackage(t *testing.T) {
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)
	initShipyardConfig(t, tempDir)

	// Run add command with invalid package
	err := runAdd(tempDir, AddOptions{
		Packages:  []string{"nonexistent"},
		Type:      "patch",
		Summary:   "Test",
		Timestamp: time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC),
	})
	assert.Error(t, err, "Add command should fail with invalid package")
	assert.Contains(t, err.Error(), "invalid package", "Error should mention invalid package")
}

// TestAddCommand_InvalidChangeType tests handling of invalid change types
func TestAddCommand_InvalidChangeType(t *testing.T) {
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)
	initShipyardConfig(t, tempDir)

	// Run add command with invalid change type
	err := runAdd(tempDir, AddOptions{
		Packages:  []string{"core"},
		Type:      "breaking",
		Summary:   "Test",
		Timestamp: time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC),
	})
	assert.Error(t, err, "Add command should fail with invalid change type")
	assert.Contains(t, err.Error(), "change type", "Error should mention change type")
}

// TestAddCommand_EmptySummary tests handling of empty summary
func TestAddCommand_EmptySummary(t *testing.T) {
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)
	initShipyardConfig(t, tempDir)

	// Run add command with empty summary
	err := runAdd(tempDir, AddOptions{
		Packages:  []string{"core"},
		Type:      "patch",
		Summary:   "",
		Timestamp: time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC),
	})
	assert.Error(t, err, "Add command should fail with empty summary")
	assert.Contains(t, err.Error(), "summary", "Error should mention summary")
}

// TestAddCommand_MetadataValidation tests metadata validation against config
func TestAddCommand_MetadataValidation(t *testing.T) {
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)
	initShipyardConfig(t, tempDir)

	// TODO (Future Enhancement): Implement metadata validation against config schema
	// Current behavior: accepts any metadata key-value pairs without validation
	// Requires: MetadataConfig.Fields schema definition and validation framework
	// See spec.md for planned metadata validation requirements
	err := runAdd(tempDir, AddOptions{
		Packages: []string{"core"},
		Type:     "patch",
		Summary:  "Test",
		Metadata: map[string]string{
			"author": "test@example.com",
			"issue":  "JIRA-123",
		},
		Timestamp: time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC),
	})
	require.NoError(t, err, "Add command should succeed with valid metadata")
}

// TestAddCommand_MultiplePackages tests creating a consignment affecting multiple packages
func TestAddCommand_MultiplePackages(t *testing.T) {
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)
	initShipyardConfig(t, tempDir)

	// Run add command with multiple packages
	err := runAdd(tempDir, AddOptions{
		Packages:  []string{"core", "api"},
		Type:      "minor",
		Summary:   "Cross-cutting change",
		Timestamp: time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC),
	})
	require.NoError(t, err, "Add command should succeed with multiple packages")

	// Verify consignment file was created
	consignmentsDir := filepath.Join(tempDir, ".shipyard", "consignments")
	entries, err := os.ReadDir(consignmentsDir)
	require.NoError(t, err, "Should be able to read consignments directory")
	assert.Equal(t, 1, len(entries), "Should have created one consignment file")
}

// TestAddCommand_NotGitRepository tests behavior when not in a git repository
func TestAddCommand_NotGitRepository(t *testing.T) {
	tempDir := t.TempDir()

	// Don't initialize git
	initShipyardConfig(t, tempDir)

	// Run add command
	err := runAdd(tempDir, AddOptions{
		Packages:  []string{"core"},
		Type:      "patch",
		Summary:   "Test",
		Timestamp: time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC),
	})
	assert.Error(t, err, "Add command should fail when not in git repository")
}

// TestAddCommand_NotInitialized tests behavior when Shipyard is not initialized
func TestAddCommand_NotInitialized(t *testing.T) {
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Don't initialize shipyard

	// Run add command
	err := runAdd(tempDir, AddOptions{
		Packages:  []string{"core"},
		Type:      "patch",
		Summary:   "Test",
		Timestamp: time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC),
	})
	assert.Error(t, err, "Add command should fail when Shipyard is not initialized")
}

// initShipyardConfig initializes a basic shipyard configuration for testing
func initShipyardConfig(t *testing.T, dir string) {
	t.Helper()

	// Create .shipyard directory
	shipyardDir := filepath.Join(dir, ".shipyard")
	require.NoError(t, os.MkdirAll(shipyardDir, 0755))

	// Create consignments directory
	consignmentsDir := filepath.Join(shipyardDir, "consignments")
	require.NoError(t, os.MkdirAll(consignmentsDir, 0755))

	// Create history file
	historyPath := filepath.Join(shipyardDir, "history.json")
	require.NoError(t, os.WriteFile(historyPath, []byte("[]"), 0644))

	// Create basic config
	cfg := &config.Config{
		Packages: []config.Package{
			{
				Name:      "core",
				Path:      "./",
				Ecosystem: config.EcosystemGo,
			},
			{
				Name:      "api",
				Path:      "./api",
				Ecosystem: config.EcosystemGo,
			},
		},
		Consignments: config.ConsignmentConfig{
			Path: ".shipyard/consignments",
		},
		History: config.HistoryConfig{
			Path: ".shipyard/history.json",
		},
	}

	configPath := filepath.Join(shipyardDir, "shipyard.yaml")
	require.NoError(t, config.WriteConfig(cfg, configPath))
}
