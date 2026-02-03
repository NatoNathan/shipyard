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

// TestFilterConsignmentsByMetadata tests filtering consignments within entries by metadata
func TestFilterConsignmentsByMetadata(t *testing.T) {
	entries := []Entry{
		{
			Version: "1.0.0",
			Package: "core",
			Consignments: []Consignment{
				{ID: "c1", Summary: "Prod change", ChangeType: "minor", Metadata: map[string]interface{}{"environment": "production"}},
				{ID: "c2", Summary: "Dev change", ChangeType: "patch", Metadata: map[string]interface{}{"environment": "development"}},
			},
		},
		{
			Version: "1.1.0",
			Package: "core",
			Consignments: []Consignment{
				{ID: "c3", Summary: "Another prod change", ChangeType: "minor", Metadata: map[string]interface{}{"environment": "production"}},
			},
		},
		{
			Version: "1.2.0",
			Package: "api",
			Consignments: []Consignment{
				{ID: "c4", Summary: "No metadata", ChangeType: "patch"},
			},
		},
	}

	t.Run("filters consignments by string metadata", func(t *testing.T) {
		result := FilterConsignmentsByMetadata(entries, "environment", "production")

		// All entries returned, but only matching consignments
		require.Len(t, result, 3)

		// First entry: should have 1 consignment (c1)
		assert.Len(t, result[0].Consignments, 1)
		assert.Equal(t, "c1", result[0].Consignments[0].ID)

		// Second entry: should have 1 consignment (c3)
		assert.Len(t, result[1].Consignments, 1)
		assert.Equal(t, "c3", result[1].Consignments[0].ID)

		// Third entry: should have 0 consignments (no match)
		assert.Len(t, result[2].Consignments, 0)
	})

	t.Run("preserves entries with no matching consignments", func(t *testing.T) {
		result := FilterConsignmentsByMetadata(entries, "environment", "staging")

		// All entries returned, all with empty consignments
		require.Len(t, result, 3)
		assert.Len(t, result[0].Consignments, 0)
		assert.Len(t, result[1].Consignments, 0)
		assert.Len(t, result[2].Consignments, 0)
	})

	t.Run("handles missing metadata key", func(t *testing.T) {
		result := FilterConsignmentsByMetadata(entries, "team", "backend")

		// All entries returned with empty consignments
		require.Len(t, result, 3)
		for _, entry := range result {
			assert.Len(t, entry.Consignments, 0)
		}
	})

	t.Run("handles empty entries", func(t *testing.T) {
		result := FilterConsignmentsByMetadata([]Entry{}, "environment", "production")
		assert.Len(t, result, 0)
	})
}

// TestSortByTimestamp tests sorting entries by timestamp
func TestSortByTimestamp(t *testing.T) {
	t.Run("sorts descending (newest first)", func(t *testing.T) {
		entries := []Entry{
			{Version: "1.0.0", Timestamp: mustParseTime("2026-01-01T10:00:00Z")},
			{Version: "1.2.0", Timestamp: mustParseTime("2026-01-03T10:00:00Z")},
			{Version: "1.1.0", Timestamp: mustParseTime("2026-01-02T10:00:00Z")},
		}

		result := SortByTimestamp(entries, true)
		require.Len(t, result, 3)
		assert.Equal(t, "1.2.0", result[0].Version)
		assert.Equal(t, "1.1.0", result[1].Version)
		assert.Equal(t, "1.0.0", result[2].Version)
	})

	t.Run("sorts ascending (oldest first)", func(t *testing.T) {
		entries := []Entry{
			{Version: "1.2.0", Timestamp: mustParseTime("2026-01-03T10:00:00Z")},
			{Version: "1.0.0", Timestamp: mustParseTime("2026-01-01T10:00:00Z")},
			{Version: "1.1.0", Timestamp: mustParseTime("2026-01-02T10:00:00Z")},
		}

		result := SortByTimestamp(entries, false)
		require.Len(t, result, 3)
		assert.Equal(t, "1.0.0", result[0].Version)
		assert.Equal(t, "1.1.0", result[1].Version)
		assert.Equal(t, "1.2.0", result[2].Version)
	})

	t.Run("handles empty entries", func(t *testing.T) {
		result := SortByTimestamp([]Entry{}, true)
		assert.Len(t, result, 0)
	})
}

// mustParseTime is a helper for parsing time strings in tests
func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
