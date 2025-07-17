// Package shipment provides functionality for tracking shipped consignments.
// A shipment represents a collection of consignments that were shipped together
// as part of a release, along with the versions they were shipped as.
package shipment

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/NatoNathan/shipyard/pkg/config"
	"github.com/NatoNathan/shipyard/pkg/semver"
)

// Consignment represents a consignment stored in shipment history
// This is a copy of the consignment structure to avoid circular imports
type Consignment struct {
	ID       string            `json:"id"`
	Packages map[string]string `json:"packages"` // package name -> change type
	Summary  string            `json:"summary"`
	Created  time.Time         `json:"created"`
}

// Shipment represents a collection of consignments that were shipped together
type Shipment struct {
	ID           string                     `json:"id"`
	Date         time.Time                  `json:"date"`
	Versions     map[string]*semver.Version `json:"versions"`     // package name -> shipped version
	Consignments []*Consignment             `json:"consignments"` // consignments included in this shipment
	Template     string                     `json:"template"`     // changelog template used
	Tags         map[string]string          `json:"tags"`         // git tags created for this shipment (package -> tag)
}

// ShipmentHistory manages the history of shipped consignments
type ShipmentHistory struct {
	projectConfig *config.ProjectConfig
	historyFile   string
}

// NewShipmentHistory creates a new shipment history manager
func NewShipmentHistory(projectConfig *config.ProjectConfig) *ShipmentHistory {
	return &ShipmentHistory{
		projectConfig: projectConfig,
		historyFile:   filepath.Join(".shipyard", "shipment-history.json"),
	}
}

// NewShipmentHistoryWithFile creates a new shipment history manager with a custom file path
func NewShipmentHistoryWithFile(projectConfig *config.ProjectConfig, filePath string) *ShipmentHistory {
	return &ShipmentHistory{
		projectConfig: projectConfig,
		historyFile:   filePath,
	}
}

// EnsureHistoryDir creates the history directory if it doesn't exist
func (h *ShipmentHistory) EnsureHistoryDir() error {
	dir := filepath.Dir(h.historyFile)
	return os.MkdirAll(dir, 0755)
}

// RecordShipment records a new shipment in the history
func (h *ShipmentHistory) RecordShipment(consignments []*Consignment, versions map[string]*semver.Version, template string) (*Shipment, error) {
	return h.RecordShipmentWithTags(consignments, versions, template, nil)
}

// RecordShipmentWithTags records a new shipment in the history with provided git tags
func (h *ShipmentHistory) RecordShipmentWithTags(consignments []*Consignment, versions map[string]*semver.Version, template string, gitTags map[string]string) (*Shipment, error) {
	if err := h.EnsureHistoryDir(); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	// Create new shipment
	shipment := &Shipment{
		ID:           generateShipmentID(),
		Date:         time.Now(),
		Versions:     make(map[string]*semver.Version),
		Consignments: make([]*Consignment, len(consignments)),
		Template:     template,
		Tags:         make(map[string]string),
	}

	// Copy versions
	for pkg, version := range versions {
		shipment.Versions[pkg] = version.Copy()
	}

	// Copy consignments
	copy(shipment.Consignments, consignments)

	// Use provided git tags or generate fallback tags
	if gitTags != nil {
		for pkg, tag := range gitTags {
			shipment.Tags[pkg] = tag
		}
	} else {
		// Generate fallback tags for backward compatibility
		for pkg, version := range versions {
			if h.projectConfig.Type == config.RepositoryTypeMonorepo {
				shipment.Tags[pkg] = fmt.Sprintf("shipyard-history/%s/%s", pkg, version.String())
			} else {
				shipment.Tags[pkg] = fmt.Sprintf("shipyard-history/%s", version.String())
			}
		}
	}

	// Load existing history
	history, err := h.LoadHistory()
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load existing history: %w", err)
	}

	// Add new shipment to history
	history = append(history, shipment)

	// Sort history by date (newest first)
	sort.Slice(history, func(i, j int) bool {
		return history[i].Date.After(history[j].Date)
	})

	// Save updated history
	if err := h.SaveHistory(history); err != nil {
		return nil, fmt.Errorf("failed to save shipment history: %w", err)
	}

	return shipment, nil
}

// LoadHistory loads the shipment history from the JSON file
func (h *ShipmentHistory) LoadHistory() ([]*Shipment, error) {
	if _, err := os.Stat(h.historyFile); os.IsNotExist(err) {
		return []*Shipment{}, nil
	}

	data, err := os.ReadFile(h.historyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	var history []*Shipment
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to parse history file: %w", err)
	}

	return history, nil
}

// SaveHistory saves the shipment history to the JSON file
func (h *ShipmentHistory) SaveHistory(history []*Shipment) error {
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	return os.WriteFile(h.historyFile, data, 0644)
}

// GetShipmentsForPackage returns all shipments that included changes for a specific package
func (h *ShipmentHistory) GetShipmentsForPackage(packageName string) ([]*Shipment, error) {
	history, err := h.LoadHistory()
	if err != nil {
		return nil, fmt.Errorf("failed to load history: %w", err)
	}

	var filtered []*Shipment
	for _, shipment := range history {
		if _, exists := shipment.Versions[packageName]; exists {
			filtered = append(filtered, shipment)
		}
	}

	return filtered, nil
}

// GetAllConsignmentsFromHistory returns all consignments from the shipment history
func (h *ShipmentHistory) GetAllConsignmentsFromHistory() ([]*Consignment, error) {
	history, err := h.LoadHistory()
	if err != nil {
		return nil, fmt.Errorf("failed to load history: %w", err)
	}

	var allConsignments []*Consignment
	for _, shipment := range history {
		allConsignments = append(allConsignments, shipment.Consignments...)
	}

	return allConsignments, nil
}

// GetVersionHistory returns a map of package names to their version history
func (h *ShipmentHistory) GetVersionHistory() (map[string][]*semver.Version, error) {
	history, err := h.LoadHistory()
	if err != nil {
		return nil, fmt.Errorf("failed to load history: %w", err)
	}

	versionHistory := make(map[string][]*semver.Version)

	for _, shipment := range history {
		for pkg, version := range shipment.Versions {
			versionHistory[pkg] = append(versionHistory[pkg], version.Copy())
		}
	}

	// Sort versions for each package (newest first)
	for pkg := range versionHistory {
		sort.Slice(versionHistory[pkg], func(i, j int) bool {
			return versionHistory[pkg][i].GreaterThan(versionHistory[pkg][j])
		})
	}

	return versionHistory, nil
}

// GetShipmentByID returns a shipment by its ID
func (h *ShipmentHistory) GetShipmentByID(id string) (*Shipment, error) {
	history, err := h.LoadHistory()
	if err != nil {
		return nil, fmt.Errorf("failed to load history: %w", err)
	}

	for _, shipment := range history {
		if shipment.ID == id {
			return shipment, nil
		}
	}

	return nil, fmt.Errorf("shipment with ID %s not found", id)
}

// GetLatestShipment returns the most recent shipment
func (h *ShipmentHistory) GetLatestShipment() (*Shipment, error) {
	history, err := h.LoadHistory()
	if err != nil {
		return nil, fmt.Errorf("failed to load history: %w", err)
	}

	if len(history) == 0 {
		return nil, fmt.Errorf("no shipments found in history")
	}

	// History is already sorted by date (newest first)
	return history[0], nil
}

// generateShipmentID generates a unique ID for the shipment
func generateShipmentID() string {
	// Use timestamp-based ID for shipments to ensure ordering
	return fmt.Sprintf("shipment-%d", time.Now().UnixNano())
}

// GetHistoryFile returns the path to the history file
func (h *ShipmentHistory) GetHistoryFile() string {
	return h.historyFile
}

// ClearHistory removes all shipment history (use with caution)
func (h *ShipmentHistory) ClearHistory() error {
	if _, err := os.Stat(h.historyFile); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to clear
	}

	return os.Remove(h.historyFile)
}
