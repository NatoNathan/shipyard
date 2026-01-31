package consignment

import (
	"testing"
	"time"

	"github.com/NatoNathan/shipyard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupConsignmentsByPackage(t *testing.T) {
	// Create test consignments
	now := time.Now()

	consignments := []*Consignment{
		{
			ID:         "c1",
			Timestamp:  now,
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "Fix bug in core",
		},
		{
			ID:         "c2",
			Timestamp:  now.Add(time.Minute),
			Packages:   []string{"api"},
			ChangeType: types.ChangeTypeMinor,
			Summary:    "Add feature to api",
		},
		{
			ID:         "c3",
			Timestamp:  now.Add(2 * time.Minute),
			Packages:   []string{"core", "api"},
			ChangeType: types.ChangeTypeMajor,
			Summary:    "Breaking change in both",
		},
		{
			ID:         "c4",
			Timestamp:  now.Add(3 * time.Minute),
			Packages:   []string{"web"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "Fix bug in web",
		},
	}

	groups := GroupConsignmentsByPackage(consignments)

	// Check core package group
	assert.Len(t, groups["core"], 2, "core should have 2 consignments")
	assert.Equal(t, "c1", groups["core"][0].ID)
	assert.Equal(t, "c3", groups["core"][1].ID)

	// Check api package group
	assert.Len(t, groups["api"], 2, "api should have 2 consignments")
	assert.Equal(t, "c2", groups["api"][0].ID)
	assert.Equal(t, "c3", groups["api"][1].ID)

	// Check web package group
	assert.Len(t, groups["web"], 1, "web should have 1 consignment")
	assert.Equal(t, "c4", groups["web"][0].ID)
}

func TestGroupConsignmentsByPackage_EmptyInput(t *testing.T) {
	groups := GroupConsignmentsByPackage([]*Consignment{})
	assert.Empty(t, groups, "empty input should return empty map")
}

func TestGroupConsignmentsByPackage_SinglePackage(t *testing.T) {
	now := time.Now()

	consignments := []*Consignment{
		{
			ID:         "c1",
			Timestamp:  now,
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "Change 1",
		},
		{
			ID:         "c2",
			Timestamp:  now.Add(time.Minute),
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypeMinor,
			Summary:    "Change 2",
		},
		{
			ID:         "c3",
			Timestamp:  now.Add(2 * time.Minute),
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "Change 3",
		},
	}

	groups := GroupConsignmentsByPackage(consignments)

	assert.Len(t, groups, 1, "should have only one package group")
	assert.Len(t, groups["core"], 3, "core should have all 3 consignments")
}

func TestGroupConsignmentsByChangeType(t *testing.T) {
	now := time.Now()

	consignments := []*Consignment{
		{
			ID:         "c1",
			Timestamp:  now,
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "Patch 1",
		},
		{
			ID:         "c2",
			Timestamp:  now.Add(time.Minute),
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypeMinor,
			Summary:    "Minor 1",
		},
		{
			ID:         "c3",
			Timestamp:  now.Add(2 * time.Minute),
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "Patch 2",
		},
		{
			ID:         "c4",
			Timestamp:  now.Add(3 * time.Minute),
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypeMajor,
			Summary:    "Major 1",
		},
	}

	groups := GroupConsignmentsByChangeType(consignments)

	assert.Len(t, groups[types.ChangeTypePatch], 2, "should have 2 patch changes")
	assert.Len(t, groups[types.ChangeTypeMinor], 1, "should have 1 minor change")
	assert.Len(t, groups[types.ChangeTypeMajor], 1, "should have 1 major change")

	// Verify specific consignments
	assert.Equal(t, "c1", groups[types.ChangeTypePatch][0].ID)
	assert.Equal(t, "c3", groups[types.ChangeTypePatch][1].ID)
	assert.Equal(t, "c2", groups[types.ChangeTypeMinor][0].ID)
	assert.Equal(t, "c4", groups[types.ChangeTypeMajor][0].ID)
}

func TestGetHighestChangeType(t *testing.T) {
	tests := []struct {
		name          string
		consignments  []*Consignment
		expectedType  types.ChangeType
		description   string
	}{
		{
			name: "all patch",
			consignments: []*Consignment{
				{ChangeType: types.ChangeTypePatch},
				{ChangeType: types.ChangeTypePatch},
			},
			expectedType: types.ChangeTypePatch,
			description:  "multiple patches should return patch",
		},
		{
			name: "patch and minor",
			consignments: []*Consignment{
				{ChangeType: types.ChangeTypePatch},
				{ChangeType: types.ChangeTypeMinor},
				{ChangeType: types.ChangeTypePatch},
			},
			expectedType: types.ChangeTypeMinor,
			description:  "minor takes precedence over patch",
		},
		{
			name: "all types mixed",
			consignments: []*Consignment{
				{ChangeType: types.ChangeTypePatch},
				{ChangeType: types.ChangeTypeMinor},
				{ChangeType: types.ChangeTypeMajor},
				{ChangeType: types.ChangeTypePatch},
			},
			expectedType: types.ChangeTypeMajor,
			description:  "major takes precedence over all",
		},
		{
			name: "only major",
			consignments: []*Consignment{
				{ChangeType: types.ChangeTypeMajor},
			},
			expectedType: types.ChangeTypeMajor,
			description:  "single major should return major",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetHighestChangeType(tt.consignments)
			assert.Equal(t, tt.expectedType, result, tt.description)
		})
	}
}

func TestGetHighestChangeType_Empty(t *testing.T) {
	// Empty slice should return patch as default (safest option)
	result := GetHighestChangeType([]*Consignment{})
	assert.Equal(t, types.ChangeTypePatch, result)
}

func TestCalculateDirectBumps(t *testing.T) {
	now := time.Now()

	consignments := []*Consignment{
		{
			ID:         "c1",
			Timestamp:  now,
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "Fix in core",
		},
		{
			ID:         "c2",
			Timestamp:  now.Add(time.Minute),
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypeMinor,
			Summary:    "Feature in core",
		},
		{
			ID:         "c3",
			Timestamp:  now.Add(2 * time.Minute),
			Packages:   []string{"api"},
			ChangeType: types.ChangeTypeMajor,
			Summary:    "Breaking change in api",
		},
		{
			ID:         "c4",
			Timestamp:  now.Add(3 * time.Minute),
			Packages:   []string{"core", "api"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "Small fix in both",
		},
	}

	bumps := CalculateDirectBumps(consignments)

	// Core has patch and minor, should be minor
	assert.Equal(t, types.ChangeTypeMinor, bumps["core"])

	// API has major and patch, should be major
	assert.Equal(t, types.ChangeTypeMajor, bumps["api"])
}

func TestFilterConsignmentsByPackage(t *testing.T) {
	now := time.Now()

	consignments := []*Consignment{
		{
			ID:         "c1",
			Timestamp:  now,
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "Core change",
		},
		{
			ID:         "c2",
			Timestamp:  now.Add(time.Minute),
			Packages:   []string{"api"},
			ChangeType: types.ChangeTypeMinor,
			Summary:    "API change",
		},
		{
			ID:         "c3",
			Timestamp:  now.Add(2 * time.Minute),
			Packages:   []string{"core", "api"},
			ChangeType: types.ChangeTypeMajor,
			Summary:    "Both change",
		},
	}

	// Filter for core
	coreConsignments := FilterConsignmentsByPackage(consignments, "core")
	assert.Len(t, coreConsignments, 2, "should have 2 consignments affecting core")
	assert.Equal(t, "c1", coreConsignments[0].ID)
	assert.Equal(t, "c3", coreConsignments[1].ID)

	// Filter for api
	apiConsignments := FilterConsignmentsByPackage(consignments, "api")
	assert.Len(t, apiConsignments, 2, "should have 2 consignments affecting api")
	assert.Equal(t, "c2", apiConsignments[0].ID)
	assert.Equal(t, "c3", apiConsignments[1].ID)

	// Filter for non-existent package
	webConsignments := FilterConsignmentsByPackage(consignments, "web")
	assert.Empty(t, webConsignments, "should have no consignments for web")
}

func TestSortConsignmentsByTimestamp(t *testing.T) {
	now := time.Now()

	consignments := []*Consignment{
		{
			ID:        "c3",
			Timestamp: now.Add(2 * time.Minute),
		},
		{
			ID:        "c1",
			Timestamp: now,
		},
		{
			ID:        "c2",
			Timestamp: now.Add(time.Minute),
		},
	}

	sorted := SortConsignmentsByTimestamp(consignments)

	// Should be sorted oldest to newest
	assert.Equal(t, "c1", sorted[0].ID)
	assert.Equal(t, "c2", sorted[1].ID)
	assert.Equal(t, "c3", sorted[2].ID)
}

func TestGroupConsignmentsByMetadata(t *testing.T) {
	now := time.Now()

	consignments := []*Consignment{
		{
			ID:        "c1",
			Timestamp: now,
			Packages:  []string{"core"},
			Metadata: map[string]interface{}{
				"author": "alice@example.com",
			},
		},
		{
			ID:        "c2",
			Timestamp: now.Add(time.Minute),
			Packages:  []string{"api"},
			Metadata: map[string]interface{}{
				"author": "bob@example.com",
			},
		},
		{
			ID:        "c3",
			Timestamp: now.Add(2 * time.Minute),
			Packages:  []string{"web"},
			Metadata: map[string]interface{}{
				"author": "alice@example.com",
			},
		},
	}

	groups := GroupConsignmentsByMetadataField(consignments, "author")

	// Check alice's consignments
	aliceGroup, ok := groups["alice@example.com"]
	require.True(t, ok)
	assert.Len(t, aliceGroup, 2)

	// Check bob's consignments
	bobGroup, ok := groups["bob@example.com"]
	require.True(t, ok)
	assert.Len(t, bobGroup, 1)
}
