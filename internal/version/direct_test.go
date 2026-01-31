package version

import (
	"testing"
	"time"

	"github.com/NatoNathan/shipyard/internal/consignment"
	"github.com/NatoNathan/shipyard/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestCalculateDirectBumps(t *testing.T) {
	t.Run("single consignment single package", func(t *testing.T) {
		consignments := []*consignment.Consignment{
			{
				ID:         "test-1",
				Timestamp:  time.Now(),
				Packages:   []string{"core"},
				ChangeType: types.ChangeTypePatch,
				Summary:    "Fix bug",
			},
		}

		bumps := CalculateDirectBumps(consignments)
		assert.Len(t, bumps, 1)
		assert.Equal(t, "patch", bumps["core"])
	})

	t.Run("single consignment multiple packages", func(t *testing.T) {
		consignments := []*consignment.Consignment{
			{
				ID:         "test-1",
				Timestamp:  time.Now(),
				Packages:   []string{"core", "api", "web"},
				ChangeType: types.ChangeTypeMinor,
				Summary:    "Add feature to multiple packages",
			},
		}

		bumps := CalculateDirectBumps(consignments)
		assert.Len(t, bumps, 3)
		assert.Equal(t, "minor", bumps["core"])
		assert.Equal(t, "minor", bumps["api"])
		assert.Equal(t, "minor", bumps["web"])
	})

	t.Run("multiple consignments same package - highest priority wins", func(t *testing.T) {
		consignments := []*consignment.Consignment{
			{
				ID:         "test-1",
				Timestamp:  time.Now(),
				Packages:   []string{"core"},
				ChangeType: types.ChangeTypePatch,
				Summary:    "Fix bug",
			},
			{
				ID:         "test-2",
				Timestamp:  time.Now(),
				Packages:   []string{"core"},
				ChangeType: types.ChangeTypeMinor,
				Summary:    "Add feature",
			},
		}

		bumps := CalculateDirectBumps(consignments)
		assert.Len(t, bumps, 1)
		// minor has higher priority than patch
		assert.Equal(t, "minor", bumps["core"])
	})

	t.Run("major change has highest priority", func(t *testing.T) {
		consignments := []*consignment.Consignment{
			{
				ID:         "test-1",
				Timestamp:  time.Now(),
				Packages:   []string{"core"},
				ChangeType: types.ChangeTypePatch,
				Summary:    "Fix bug",
			},
			{
				ID:         "test-2",
				Timestamp:  time.Now(),
				Packages:   []string{"core"},
				ChangeType: types.ChangeTypeMajor,
				Summary:    "Breaking change",
			},
			{
				ID:         "test-3",
				Timestamp:  time.Now(),
				Packages:   []string{"core"},
				ChangeType: types.ChangeTypeMinor,
				Summary:    "Add feature",
			},
		}

		bumps := CalculateDirectBumps(consignments)
		assert.Len(t, bumps, 1)
		// major has highest priority
		assert.Equal(t, "major", bumps["core"])
	})

	t.Run("multiple packages different changes", func(t *testing.T) {
		consignments := []*consignment.Consignment{
			{
				ID:         "test-1",
				Timestamp:  time.Now(),
				Packages:   []string{"core"},
				ChangeType: types.ChangeTypeMajor,
				Summary:    "Breaking change in core",
			},
			{
				ID:         "test-2",
				Timestamp:  time.Now(),
				Packages:   []string{"api"},
				ChangeType: types.ChangeTypeMinor,
				Summary:    "Feature in api",
			},
			{
				ID:         "test-3",
				Timestamp:  time.Now(),
				Packages:   []string{"utils"},
				ChangeType: types.ChangeTypePatch,
				Summary:    "Fix in utils",
			},
		}

		bumps := CalculateDirectBumps(consignments)
		assert.Len(t, bumps, 3)
		assert.Equal(t, "major", bumps["core"])
		assert.Equal(t, "minor", bumps["api"])
		assert.Equal(t, "patch", bumps["utils"])
	})

	t.Run("overlapping packages get highest priority", func(t *testing.T) {
		consignments := []*consignment.Consignment{
			{
				ID:         "test-1",
				Timestamp:  time.Now(),
				Packages:   []string{"core", "api"},
				ChangeType: types.ChangeTypePatch,
				Summary:    "Bug fix",
			},
			{
				ID:         "test-2",
				Timestamp:  time.Now(),
				Packages:   []string{"api", "web"},
				ChangeType: types.ChangeTypeMinor,
				Summary:    "Feature",
			},
			{
				ID:         "test-3",
				Timestamp:  time.Now(),
				Packages:   []string{"web"},
				ChangeType: types.ChangeTypeMajor,
				Summary:    "Breaking",
			},
		}

		bumps := CalculateDirectBumps(consignments)
		assert.Len(t, bumps, 3)
		assert.Equal(t, "patch", bumps["core"])   // only patch
		assert.Equal(t, "minor", bumps["api"])    // patch + minor = minor
		assert.Equal(t, "major", bumps["web"])    // minor + major = major
	})

	t.Run("empty consignments returns empty map", func(t *testing.T) {
		bumps := CalculateDirectBumps([]*consignment.Consignment{})
		assert.Empty(t, bumps)
	})

	t.Run("nil consignments returns empty map", func(t *testing.T) {
		bumps := CalculateDirectBumps(nil)
		assert.Empty(t, bumps)
	})
}

func TestGetChangePriority(t *testing.T) {
	tests := []struct {
		name       string
		changeType string
		expected   int
	}{
		{"patch has priority 1", "patch", 1},
		{"minor has priority 2", "minor", 2},
		{"major has priority 3", "major", 3},
		{"unknown defaults to 0", "unknown", 0},
		{"empty string defaults to 0", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priority := GetChangePriority(tt.changeType)
			assert.Equal(t, tt.expected, priority)
		})
	}
}

func TestIsHigherPriority(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{"major > minor", "major", "minor", true},
		{"major > patch", "major", "patch", true},
		{"minor > patch", "minor", "patch", true},
		{"minor < major", "minor", "major", false},
		{"patch < minor", "patch", "minor", false},
		{"patch < major", "patch", "major", false},
		{"major == major", "major", "major", false},
		{"minor == minor", "minor", "minor", false},
		{"patch == patch", "patch", "patch", false},
		{"unknown < patch", "unknown", "patch", false},
		{"patch > unknown", "patch", "unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsHigherPriority(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}
