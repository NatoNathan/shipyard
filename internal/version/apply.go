package version

import (
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/NatoNathan/shipyard/pkg/types"
)

// ApplyBump applies a change type bump to a version and returns the new version
func ApplyBump(current semver.Version, changeType types.ChangeType) semver.Version {
	newVersion, err := current.Bump(string(changeType))
	if err != nil {
		// If bump fails, return current version unchanged
		return current
	}
	return newVersion
}

// ApplyBumpMap applies version bumps to a map of packages
// Returns a new map with updated versions
// Packages not in the bumps map remain at their current version
func ApplyBumpMap(currentVersions map[string]semver.Version, bumps map[string]types.ChangeType) map[string]semver.Version {
	result := make(map[string]semver.Version)

	// Copy all current versions first
	for pkg, ver := range currentVersions {
		result[pkg] = ver
	}

	// Apply bumps where they exist
	for pkg, bump := range bumps {
		if current, ok := currentVersions[pkg]; ok {
			result[pkg] = ApplyBump(current, bump)
		}
	}

	return result
}

// CalculateNewVersions combines direct and propagated bumps, taking the higher of the two
// Returns a map of package name to new version
func CalculateNewVersions(
	currentVersions map[string]semver.Version,
	directBumps map[string]types.ChangeType,
	propagatedBumps map[string]types.ChangeType,
) map[string]semver.Version {
	// Merge bump maps, taking the maximum bump type for each package
	mergedBumps := MergeBumpMaps(directBumps, propagatedBumps)

	// Apply merged bumps
	return ApplyBumpMap(currentVersions, mergedBumps)
}

// MaxChangeType returns the higher priority change type
// Priority order: major > minor > patch
func MaxChangeType(a, b types.ChangeType) types.ChangeType {
	priorityA := GetBumpPriority(a)
	priorityB := GetBumpPriority(b)

	if priorityA > priorityB {
		return a
	}
	return b
}

// GetBumpPriority returns a numeric priority for a change type
// Higher number = higher priority
func GetBumpPriority(changeType types.ChangeType) int {
	switch changeType {
	case types.ChangeTypeMajor:
		return 3
	case types.ChangeTypeMinor:
		return 2
	case types.ChangeTypePatch:
		return 1
	default:
		return 0
	}
}

// MergeBumpMaps merges two bump maps, taking the maximum change type for each package
// If a package exists in both maps, the higher priority change type wins
func MergeBumpMaps(map1, map2 map[string]types.ChangeType) map[string]types.ChangeType {
	result := make(map[string]types.ChangeType)

	// Copy all from map1
	for pkg, bump := range map1 {
		result[pkg] = bump
	}

	// Merge from map2, taking max
	for pkg, bump := range map2 {
		if existing, ok := result[pkg]; ok {
			result[pkg] = MaxChangeType(existing, bump)
		} else {
			result[pkg] = bump
		}
	}

	return result
}

// FilterBumpsByPackages filters a bump map to only include specified packages
func FilterBumpsByPackages(bumps map[string]types.ChangeType, packages []string) map[string]types.ChangeType {
	if len(packages) == 0 {
		return bumps // No filter, return all
	}

	packageSet := make(map[string]bool)
	for _, pkg := range packages {
		packageSet[pkg] = true
	}

	filtered := make(map[string]types.ChangeType)
	for pkg, bump := range bumps {
		if packageSet[pkg] {
			filtered[pkg] = bump
		}
	}

	return filtered
}

// GetPackagesWithBumps returns a sorted list of package names that have version bumps
func GetPackagesWithBumps(bumps map[string]types.ChangeType) []string {
	packages := make([]string, 0, len(bumps))
	for pkg := range bumps {
		packages = append(packages, pkg)
	}

	// Sort for consistent output
	// Note: Using a simple sort here, could be optimized if needed
	for i := 0; i < len(packages); i++ {
		for j := i + 1; j < len(packages); j++ {
			if packages[i] > packages[j] {
				packages[i], packages[j] = packages[j], packages[i]
			}
		}
	}

	return packages
}

// VersionDiff represents a version change
type VersionDiff struct {
	Package    string
	OldVersion semver.Version
	NewVersion semver.Version
	ChangeType types.ChangeType
}

// CalculateVersionDiffs computes the differences between current and new versions
func CalculateVersionDiffs(
	currentVersions map[string]semver.Version,
	newVersions map[string]semver.Version,
	bumps map[string]types.ChangeType,
) []VersionDiff {
	var diffs []VersionDiff

	for pkg, newVer := range newVersions {
		if currentVer, ok := currentVersions[pkg]; ok {
			// Only include if version actually changed
			if currentVer.String() != newVer.String() {
				diff := VersionDiff{
					Package:    pkg,
					OldVersion: currentVer,
					NewVersion: newVer,
					ChangeType: bumps[pkg],
				}
				diffs = append(diffs, diff)
			}
		}
	}

	return diffs
}
