package version

import (
	"github.com/NatoNathan/shipyard/internal/consignment"
)

// CalculateDirectBumps calculates the highest priority change type for each package
// based on consignments. Returns a map of package name to change type string.
//
// When multiple consignments affect the same package, the highest priority change
// type is selected according to: major > minor > patch
func CalculateDirectBumps(consignments []*consignment.Consignment) map[string]string {
	if consignments == nil {
		return make(map[string]string)
	}

	bumps := make(map[string]string)

	for _, c := range consignments {
		changeType := string(c.ChangeType)

		for _, pkg := range c.Packages {
			// If package not seen yet, or this change is higher priority
			if existing, ok := bumps[pkg]; !ok || IsHigherPriority(changeType, existing) {
				bumps[pkg] = changeType
			}
		}
	}

	return bumps
}

// GetChangePriority returns the numeric priority of a change type.
// Higher numbers indicate higher priority.
// Priority order: major (3) > minor (2) > patch (1) > unknown (0)
func GetChangePriority(changeType string) int {
	priorities := map[string]int{
		"patch": 1,
		"minor": 2,
		"major": 3,
	}

	if priority, ok := priorities[changeType]; ok {
		return priority
	}

	return 0 // Unknown change type
}

// IsHigherPriority returns true if change type a has higher priority than b.
// Priority order: major > minor > patch
func IsHigherPriority(a, b string) bool {
	return GetChangePriority(a) > GetChangePriority(b)
}
