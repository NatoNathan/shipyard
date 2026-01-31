package version

import (
	"testing"

	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/NatoNathan/shipyard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyBump(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		changeType     types.ChangeType
		expectedVersion string
		description    string
	}{
		{
			name:           "patch bump",
			currentVersion: "1.2.3",
			changeType:     types.ChangeTypePatch,
			expectedVersion: "1.2.4",
			description:    "patch increments patch version",
		},
		{
			name:           "minor bump",
			currentVersion: "1.2.3",
			changeType:     types.ChangeTypeMinor,
			expectedVersion: "1.3.0",
			description:    "minor increments minor and resets patch",
		},
		{
			name:           "major bump",
			currentVersion: "1.2.3",
			changeType:     types.ChangeTypeMajor,
			expectedVersion: "2.0.0",
			description:    "major increments major and resets minor and patch",
		},
		{
			name:           "patch from zero",
			currentVersion: "0.0.0",
			changeType:     types.ChangeTypePatch,
			expectedVersion: "0.0.1",
			description:    "patch works from 0.0.0",
		},
		{
			name:           "minor from zero",
			currentVersion: "0.0.0",
			changeType:     types.ChangeTypeMinor,
			expectedVersion: "0.1.0",
			description:    "minor works from 0.0.0",
		},
		{
			name:           "major from zero",
			currentVersion: "0.1.2",
			changeType:     types.ChangeTypeMajor,
			expectedVersion: "1.0.0",
			description:    "major from 0.x.x goes to 1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, err := semver.Parse(tt.currentVersion)
			require.NoError(t, err)

			result := ApplyBump(current, tt.changeType)

			assert.Equal(t, tt.expectedVersion, result.String(), tt.description)
		})
	}
}

func TestApplyBumpMap(t *testing.T) {
	tests := []struct {
		name            string
		currentVersions map[string]string
		bumps           map[string]types.ChangeType
		expected        map[string]string
		description     string
	}{
		{
			name: "single package",
			currentVersions: map[string]string{
				"core": "1.0.0",
			},
			bumps: map[string]types.ChangeType{
				"core": types.ChangeTypeMinor,
			},
			expected: map[string]string{
				"core": "1.1.0",
			},
			description: "single package bump",
		},
		{
			name: "multiple packages different bumps",
			currentVersions: map[string]string{
				"core": "1.0.0",
				"api":  "2.5.3",
				"web":  "0.1.0",
			},
			bumps: map[string]types.ChangeType{
				"core": types.ChangeTypeMajor,
				"api":  types.ChangeTypePatch,
				"web":  types.ChangeTypeMinor,
			},
			expected: map[string]string{
				"core": "2.0.0",
				"api":  "2.5.4",
				"web":  "0.2.0",
			},
			description: "multiple packages with different bump types",
		},
		{
			name: "package with no bump stays same",
			currentVersions: map[string]string{
				"core": "1.0.0",
				"api":  "2.0.0",
			},
			bumps: map[string]types.ChangeType{
				"core": types.ChangeTypeMinor,
			},
			expected: map[string]string{
				"core": "1.1.0",
				"api":  "2.0.0", // No bump, stays same
			},
			description: "packages not in bump map remain unchanged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse current versions
			currentVersionMap := make(map[string]semver.Version)
			for pkg, ver := range tt.currentVersions {
				v, err := semver.Parse(ver)
				require.NoError(t, err)
				currentVersionMap[pkg] = v
			}

			// Apply bumps
			result := ApplyBumpMap(currentVersionMap, tt.bumps)

			// Verify results
			for pkg, expectedVer := range tt.expected {
				actual, ok := result[pkg]
				require.True(t, ok, "package %s should be in result", pkg)
				assert.Equal(t, expectedVer, actual.String(), "package %s version mismatch", pkg)
			}
		})
	}
}

func TestCalculateNewVersions(t *testing.T) {
	tests := []struct {
		name            string
		currentVersions map[string]string
		directBumps     map[string]types.ChangeType
		propagatedBumps map[string]types.ChangeType
		expected        map[string]string
		description     string
	}{
		{
			name: "direct bump only",
			currentVersions: map[string]string{
				"core": "1.0.0",
			},
			directBumps: map[string]types.ChangeType{
				"core": types.ChangeTypeMinor,
			},
			propagatedBumps: map[string]types.ChangeType{},
			expected: map[string]string{
				"core": "1.1.0",
			},
			description: "direct bump without propagation",
		},
		{
			name: "propagated bump only",
			currentVersions: map[string]string{
				"api": "2.0.0",
			},
			directBumps: map[string]types.ChangeType{},
			propagatedBumps: map[string]types.ChangeType{
				"api": types.ChangeTypePatch,
			},
			expected: map[string]string{
				"api": "2.0.1",
			},
			description: "propagated bump without direct changes",
		},
		{
			name: "both direct and propagated - direct wins if higher",
			currentVersions: map[string]string{
				"api": "2.0.0",
			},
			directBumps: map[string]types.ChangeType{
				"api": types.ChangeTypeMinor,
			},
			propagatedBumps: map[string]types.ChangeType{
				"api": types.ChangeTypePatch,
			},
			expected: map[string]string{
				"api": "2.1.0", // Minor wins over patch
			},
			description: "higher bump type wins",
		},
		{
			name: "both direct and propagated - propagated wins if higher",
			currentVersions: map[string]string{
				"api": "2.0.0",
			},
			directBumps: map[string]types.ChangeType{
				"api": types.ChangeTypePatch,
			},
			propagatedBumps: map[string]types.ChangeType{
				"api": types.ChangeTypeMajor,
			},
			expected: map[string]string{
				"api": "3.0.0", // Major wins over patch
			},
			description: "propagated major overrides direct patch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse current versions
			currentVersionMap := make(map[string]semver.Version)
			for pkg, ver := range tt.currentVersions {
				v, err := semver.Parse(ver)
				require.NoError(t, err)
				currentVersionMap[pkg] = v
			}

			// Calculate new versions
			result := CalculateNewVersions(currentVersionMap, tt.directBumps, tt.propagatedBumps)

			// Verify results
			for pkg, expectedVer := range tt.expected {
				actual, ok := result[pkg]
				require.True(t, ok, "package %s should be in result", pkg)
				assert.Equal(t, expectedVer, actual.String(), tt.description)
			}
		})
	}
}

func TestMaxChangeType(t *testing.T) {
	tests := []struct {
		name        string
		a           types.ChangeType
		b           types.ChangeType
		expected    types.ChangeType
		description string
	}{
		{
			name:        "major vs minor",
			a:           types.ChangeTypeMajor,
			b:           types.ChangeTypeMinor,
			expected:    types.ChangeTypeMajor,
			description: "major is higher than minor",
		},
		{
			name:        "major vs patch",
			a:           types.ChangeTypeMajor,
			b:           types.ChangeTypePatch,
			expected:    types.ChangeTypeMajor,
			description: "major is higher than patch",
		},
		{
			name:        "minor vs patch",
			a:           types.ChangeTypeMinor,
			b:           types.ChangeTypePatch,
			expected:    types.ChangeTypeMinor,
			description: "minor is higher than patch",
		},
		{
			name:        "same type",
			a:           types.ChangeTypeMinor,
			b:           types.ChangeTypeMinor,
			expected:    types.ChangeTypeMinor,
			description: "same type returns same",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaxChangeType(tt.a, tt.b)
			assert.Equal(t, tt.expected, result, tt.description)

			// Test symmetry
			result2 := MaxChangeType(tt.b, tt.a)
			assert.Equal(t, tt.expected, result2, "should be symmetric")
		})
	}
}

func TestMergeBumpMaps(t *testing.T) {
	map1 := map[string]types.ChangeType{
		"core": types.ChangeTypeMinor,
		"api":  types.ChangeTypePatch,
	}

	map2 := map[string]types.ChangeType{
		"api": types.ChangeTypeMajor, // Should override patch
		"web": types.ChangeTypeMinor,
	}

	result := MergeBumpMaps(map1, map2)

	// Core should be from map1
	assert.Equal(t, types.ChangeTypeMinor, result["core"])

	// API should be major (higher of patch and major)
	assert.Equal(t, types.ChangeTypeMajor, result["api"])

	// Web should be from map2
	assert.Equal(t, types.ChangeTypeMinor, result["web"])
}

func TestGetBumpPriority(t *testing.T) {
	tests := []struct {
		changeType types.ChangeType
		expected   int
	}{
		{types.ChangeTypePatch, 1},
		{types.ChangeTypeMinor, 2},
		{types.ChangeTypeMajor, 3},
	}

	for _, tt := range tests {
		t.Run(string(tt.changeType), func(t *testing.T) {
			priority := GetBumpPriority(tt.changeType)
			assert.Equal(t, tt.expected, priority)
		})
	}

	// Verify ordering
	assert.Less(t, GetBumpPriority(types.ChangeTypePatch), GetBumpPriority(types.ChangeTypeMinor))
	assert.Less(t, GetBumpPriority(types.ChangeTypeMinor), GetBumpPriority(types.ChangeTypeMajor))
}
