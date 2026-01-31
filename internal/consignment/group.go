package consignment

import (
	"sort"
	"time"

	"github.com/NatoNathan/shipyard/pkg/types"
)

// GroupConsignmentsByChangeType groups consignments by their change type
func GroupConsignmentsByChangeType(consignments []*Consignment) map[types.ChangeType][]*Consignment {
	groups := make(map[types.ChangeType][]*Consignment)

	for _, c := range consignments {
		groups[c.ChangeType] = append(groups[c.ChangeType], c)
	}

	return groups
}

// GetHighestChangeType returns the highest priority change type from a list of consignments
// Priority: major > minor > patch
func GetHighestChangeType(consignments []*Consignment) types.ChangeType {
	if len(consignments) == 0 {
		return types.ChangeTypePatch // Default to safest option
	}

	highest := types.ChangeTypePatch

	for _, c := range consignments {
		switch c.ChangeType {
		case types.ChangeTypeMajor:
			return types.ChangeTypeMajor // Can't get higher, return immediately
		case types.ChangeTypeMinor:
			if highest == types.ChangeTypePatch {
				highest = types.ChangeTypeMinor
			}
		case types.ChangeTypePatch:
			// Already at lowest priority, no change needed
		}
	}

	return highest
}

// CalculateDirectBumps calculates the version bump for each package based on direct consignments
// Returns a map of package name to highest change type affecting that package
func CalculateDirectBumps(consignments []*Consignment) map[string]types.ChangeType {
	// First group by package
	packageGroups := GroupConsignmentsByPackage(consignments)

	// Then find highest change type for each package
	bumps := make(map[string]types.ChangeType)

	for pkg, pkgConsignments := range packageGroups {
		bumps[pkg] = GetHighestChangeType(pkgConsignments)
	}

	return bumps
}

// FilterConsignmentsByPackage returns consignments that affect the specified package
func FilterConsignmentsByPackage(consignments []*Consignment, packageName string) []*Consignment {
	var filtered []*Consignment

	for _, c := range consignments {
		for _, pkg := range c.Packages {
			if pkg == packageName {
				filtered = append(filtered, c)
				break // Only add once even if package appears multiple times
			}
		}
	}

	return filtered
}

// SortConsignmentsByTimestamp returns a new slice sorted by timestamp (oldest first)
// Does not modify the input slice
func SortConsignmentsByTimestamp(consignments []*Consignment) []*Consignment {
	// Create a copy to avoid modifying input
	sorted := make([]*Consignment, len(consignments))
	copy(sorted, consignments)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	return sorted
}

// GroupConsignmentsByMetadataField groups consignments by a specific metadata field value
// Returns a map where keys are the field values (as strings) and values are consignment slices
func GroupConsignmentsByMetadataField(consignments []*Consignment, fieldName string) map[string][]*Consignment {
	groups := make(map[string][]*Consignment)

	for _, c := range consignments {
		if c.Metadata == nil {
			continue
		}

		fieldValue, ok := c.Metadata[fieldName]
		if !ok {
			continue
		}

		// Convert field value to string for use as map key
		var key string
		switch v := fieldValue.(type) {
		case string:
			key = v
		default:
			// For non-string values, convert to string representation
			key = ""
		}

		if key != "" {
			groups[key] = append(groups[key], c)
		}
	}

	return groups
}

// GetUniquePackages returns a sorted list of unique package names from consignments
func GetUniquePackages(consignments []*Consignment) []string {
	packageSet := make(map[string]bool)

	for _, c := range consignments {
		for _, pkg := range c.Packages {
			packageSet[pkg] = true
		}
	}

	// Convert to sorted slice
	packages := make([]string, 0, len(packageSet))
	for pkg := range packageSet {
		packages = append(packages, pkg)
	}

	sort.Strings(packages)

	return packages
}

// GroupConsignmentsByPackage groups consignments by package name
// A consignment affecting multiple packages will appear in multiple groups
func GroupConsignmentsByPackage(consignments []*Consignment) map[string][]*Consignment {
	groups := make(map[string][]*Consignment)

	for _, c := range consignments {
		for _, pkg := range c.Packages {
			groups[pkg] = append(groups[pkg], c)
		}
	}

	return groups
}

// AggregateMetadata collects all metadata fields from consignments
// If a field appears multiple times with different values, the last value wins
func AggregateMetadata(consignments []*Consignment) map[string]interface{} {
	aggregated := make(map[string]interface{})

	for _, c := range consignments {
		if c.Metadata == nil {
			continue
		}

		for key, value := range c.Metadata {
			aggregated[key] = value
		}
	}

	return aggregated
}

// GetConsignmentsByDateRange filters consignments within a date range (inclusive)
// Both start and end times are optional (nil means no bound)
func GetConsignmentsByDateRange(consignments []*Consignment, start, end *time.Time) []*Consignment {
	var filtered []*Consignment

	for _, c := range consignments {
		if start != nil && c.Timestamp.Before(*start) {
			continue
		}

		if end != nil && c.Timestamp.After(*end) {
			continue
		}

		filtered = append(filtered, c)
	}

	return filtered
}
