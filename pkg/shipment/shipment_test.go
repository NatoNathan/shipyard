package shipment

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NatoNathan/shipyard/pkg/config"
	"github.com/NatoNathan/shipyard/pkg/semver"
)

func TestShipmentHistory(t *testing.T) {
	tempDir := t.TempDir()
	historyFile := filepath.Join(tempDir, "shipment-history.json")

	projectConfig := &config.ProjectConfig{
		Type: config.RepositoryTypeSingleRepo,
		Package: config.Package{
			Name: "test-package",
			Path: ".",
		},
		Changelog: config.ChangelogConfig{
			Template: "keepachangelog",
		},
	}

	history := NewShipmentHistoryWithFile(projectConfig, historyFile)

	// Test recording a shipment
	consignments := []*Consignment{
		{
			ID: "test1",
			Packages: map[string]string{
				"test-package": "minor",
			},
			Summary: "Added new feature",
			Created: time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	version, _ := semver.Parse("1.1.0")
	versions := map[string]*semver.Version{
		"test-package": version,
	}

	shipment, err := history.RecordShipment(consignments, versions, "keepachangelog")
	if err != nil {
		t.Fatalf("Failed to record shipment: %v", err)
	}

	if shipment.ID == "" {
		t.Error("Expected shipment ID to be generated")
	}

	if len(shipment.Consignments) != 1 {
		t.Errorf("Expected 1 consignment, got %d", len(shipment.Consignments))
	}

	if shipment.Template != "keepachangelog" {
		t.Errorf("Expected template 'keepachangelog', got %s", shipment.Template)
	}

	// Test loading history
	loadedHistory, err := history.LoadHistory()
	if err != nil {
		t.Fatalf("Failed to load history: %v", err)
	}

	if len(loadedHistory) != 1 {
		t.Errorf("Expected 1 shipment in history, got %d", len(loadedHistory))
	}

	// Test getting latest shipment
	latest, err := history.GetLatestShipment()
	if err != nil {
		t.Fatalf("Failed to get latest shipment: %v", err)
	}

	if latest.ID != shipment.ID {
		t.Errorf("Expected latest shipment ID %s, got %s", shipment.ID, latest.ID)
	}

	// Test getting shipments for package
	packageShipments, err := history.GetShipmentsForPackage("test-package")
	if err != nil {
		t.Fatalf("Failed to get shipments for package: %v", err)
	}

	if len(packageShipments) != 1 {
		t.Errorf("Expected 1 shipment for package, got %d", len(packageShipments))
	}
}

func TestShipmentHistoryMultipleShipments(t *testing.T) {
	tempDir := t.TempDir()
	historyFile := filepath.Join(tempDir, "shipment-history.json")

	projectConfig := &config.ProjectConfig{
		Type: config.RepositoryTypeSingleRepo,
		Package: config.Package{
			Name: "test-package",
			Path: ".",
		},
		Changelog: config.ChangelogConfig{
			Template: "keepachangelog",
		},
	}

	history := NewShipmentHistoryWithFile(projectConfig, historyFile)

	// Record multiple shipments
	shipments := []struct {
		version string
		summary string
	}{
		{"1.0.0", "Initial release"},
		{"1.1.0", "Added new feature"},
		{"1.1.1", "Bug fix"},
	}

	var recordedShipments []*Shipment
	for i, s := range shipments {
		// Add a small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)

		consignments := []*Consignment{
			{
				ID: "test-" + s.version,
				Packages: map[string]string{
					"test-package": "patch",
				},
				Summary: s.summary,
				Created: time.Now(),
			},
		}

		version, _ := semver.Parse(s.version)
		versions := map[string]*semver.Version{
			"test-package": version,
		}

		shipment, err := history.RecordShipment(consignments, versions, "keepachangelog")
		if err != nil {
			t.Fatalf("Failed to record shipment %s: %v", s.version, err)
		}
		recordedShipments = append(recordedShipments, shipment)

		// For testing, we expect the last recorded shipment to be the latest
		if i == len(shipments)-1 {
			// Verify this is the latest
			latest, err := history.GetLatestShipment()
			if err != nil {
				t.Fatalf("Failed to get latest shipment: %v", err)
			}
			if latest.ID != shipment.ID {
				t.Errorf("Expected latest shipment ID %s, got %s", shipment.ID, latest.ID)
			}
		}
	}

	// Test that history is sorted by date (newest first)
	loadedHistory, err := history.LoadHistory()
	if err != nil {
		t.Fatalf("Failed to load history: %v", err)
	}

	if len(loadedHistory) != 3 {
		t.Errorf("Expected 3 shipments in history, got %d", len(loadedHistory))
	}

	// Check that the latest shipment is first
	latest, err := history.GetLatestShipment()
	if err != nil {
		t.Fatalf("Failed to get latest shipment: %v", err)
	}

	if latest.ID != loadedHistory[0].ID {
		t.Errorf("Latest shipment should be first in history. Latest ID: %s, First in history ID: %s", latest.ID, loadedHistory[0].ID)
	}

	// Test version history
	versionHistory, err := history.GetVersionHistory()
	if err != nil {
		t.Fatalf("Failed to get version history: %v", err)
	}

	packageVersions, exists := versionHistory["test-package"]
	if !exists {
		t.Error("Expected version history for test-package")
	}

	if len(packageVersions) != 3 {
		t.Errorf("Expected 3 versions in history, got %d", len(packageVersions))
	}
}

func TestShipmentHistoryEmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	historyFile := filepath.Join(tempDir, "nonexistent-history.json")

	projectConfig := &config.ProjectConfig{
		Type: config.RepositoryTypeSingleRepo,
		Package: config.Package{
			Name: "test-package",
			Path: ".",
		},
	}

	history := NewShipmentHistoryWithFile(projectConfig, historyFile)

	// Test loading non-existent history
	loadedHistory, err := history.LoadHistory()
	if err != nil {
		t.Fatalf("Failed to load empty history: %v", err)
	}

	if len(loadedHistory) != 0 {
		t.Errorf("Expected empty history, got %d shipments", len(loadedHistory))
	}

	// Test getting latest shipment from empty history
	_, err = history.GetLatestShipment()
	if err == nil {
		t.Error("Expected error when getting latest shipment from empty history")
	}
}

func TestShipmentHistoryClearHistory(t *testing.T) {
	tempDir := t.TempDir()
	historyFile := filepath.Join(tempDir, "shipment-history.json")

	projectConfig := &config.ProjectConfig{
		Type: config.RepositoryTypeSingleRepo,
		Package: config.Package{
			Name: "test-package",
			Path: ".",
		},
	}

	history := NewShipmentHistoryWithFile(projectConfig, historyFile)

	// Record a shipment
	consignments := []*Consignment{
		{
			ID: "test1",
			Packages: map[string]string{
				"test-package": "minor",
			},
			Summary: "Test shipment",
			Created: time.Now(),
		},
	}

	version, _ := semver.Parse("1.0.0")
	versions := map[string]*semver.Version{
		"test-package": version,
	}

	_, err := history.RecordShipment(consignments, versions, "keepachangelog")
	if err != nil {
		t.Fatalf("Failed to record shipment: %v", err)
	}

	// Verify history exists
	if _, err := os.Stat(historyFile); os.IsNotExist(err) {
		t.Error("History file should exist")
	}

	// Clear history
	err = history.ClearHistory()
	if err != nil {
		t.Fatalf("Failed to clear history: %v", err)
	}

	// Verify history file is removed
	if _, err := os.Stat(historyFile); !os.IsNotExist(err) {
		t.Error("History file should be removed after clearing")
	}
}
