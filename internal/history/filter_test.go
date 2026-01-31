package history

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFilterByPackage tests filtering history entries by package name
func TestFilterByPackage(t *testing.T) {
	timestamp := time.Now()

	entries := []Entry{
		{
			Version:   "1.1.0",
			Package:   "core",
			Timestamp: timestamp,
			Consignments: []Consignment{
				{ID: "c1", Summary: "Add feature", ChangeType: "minor"},
			},
		},
		{
			Version:   "2.0.0",
			Package:   "api",
			Timestamp: timestamp,
			Consignments: []Consignment{
				{ID: "c2", Summary: "Breaking change", ChangeType: "major"},
			},
		},
		{
			Version:   "1.0.1",
			Package:   "core",
			Timestamp: timestamp,
			Consignments: []Consignment{
				{ID: "c3", Summary: "Fix bug", ChangeType: "patch"},
			},
		},
	}

	t.Run("filters by single package", func(t *testing.T) {
		// Test: Filter by "core" package
		filtered := FilterByPackage(entries, "core")

		// Verify: Only core entries returned
		require.Len(t, filtered, 2)
		assert.Equal(t, "core", filtered[0].Package)
		assert.Equal(t, "core", filtered[1].Package)
		assert.Equal(t, "1.1.0", filtered[0].Version)
		assert.Equal(t, "1.0.1", filtered[1].Version)
	})

	t.Run("returns empty slice for non-existent package", func(t *testing.T) {
		// Test: Filter by package that doesn't exist
		filtered := FilterByPackage(entries, "nonexistent")

		// Verify: Empty slice returned
		assert.Len(t, filtered, 0)
	})

	t.Run("returns all entries for empty package name", func(t *testing.T) {
		// Test: Filter with empty package name
		filtered := FilterByPackage(entries, "")

		// Verify: All entries returned
		assert.Len(t, filtered, 3)
	})
}

// TestFilterByVersion tests filtering history entries by version
func TestFilterByVersion(t *testing.T) {
	timestamp := time.Now()

	entries := []Entry{
		{
			Version:   "1.1.0",
			Package:   "core",
			Timestamp: timestamp,
			Consignments: []Consignment{
				{ID: "c1", Summary: "Add feature", ChangeType: "minor"},
			},
		},
		{
			Version:   "2.0.0",
			Package:   "api",
			Timestamp: timestamp,
			Consignments: []Consignment{
				{ID: "c2", Summary: "Breaking change", ChangeType: "major"},
			},
		},
	}

	t.Run("filters by specific version", func(t *testing.T) {
		// Test: Filter by version "1.1.0"
		filtered := FilterByVersion(entries, "1.1.0")

		// Verify: Only matching version returned
		require.Len(t, filtered, 1)
		assert.Equal(t, "1.1.0", filtered[0].Version)
		assert.Equal(t, "core", filtered[0].Package)
	})

	t.Run("returns empty slice for non-existent version", func(t *testing.T) {
		// Test: Filter by version that doesn't exist
		filtered := FilterByVersion(entries, "99.99.99")

		// Verify: Empty slice returned
		assert.Len(t, filtered, 0)
	})

	t.Run("returns all entries for empty version", func(t *testing.T) {
		// Test: Filter with empty version
		filtered := FilterByVersion(entries, "")

		// Verify: All entries returned
		assert.Len(t, filtered, 2)
	})
}

// TestCombinedFilters tests applying both package and version filters
func TestCombinedFilters(t *testing.T) {
	timestamp := time.Now()

	entries := []Entry{
		{
			Version:   "1.1.0",
			Package:   "core",
			Timestamp: timestamp,
			Consignments: []Consignment{
				{ID: "c1", Summary: "Add feature", ChangeType: "minor"},
			},
		},
		{
			Version:   "1.1.0",
			Package:   "api",
			Timestamp: timestamp,
			Consignments: []Consignment{
				{ID: "c2", Summary: "Add API", ChangeType: "minor"},
			},
		},
		{
			Version:   "1.0.1",
			Package:   "core",
			Timestamp: timestamp,
			Consignments: []Consignment{
				{ID: "c3", Summary: "Fix bug", ChangeType: "patch"},
			},
		},
	}

	t.Run("applies both filters", func(t *testing.T) {
		// Test: Filter by package "core" then version "1.1.0"
		filtered := FilterByPackage(entries, "core")
		filtered = FilterByVersion(filtered, "1.1.0")

		// Verify: Only matching entry returned
		require.Len(t, filtered, 1)
		assert.Equal(t, "core", filtered[0].Package)
		assert.Equal(t, "1.1.0", filtered[0].Version)
	})
}
