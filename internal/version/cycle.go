package version

import (
	"fmt"

	"github.com/NatoNathan/shipyard/internal/graph"
	"github.com/NatoNathan/shipyard/pkg/semver"
)

// ResolveCycleBumps resolves version bumps for packages in cycles (SCCs).
// When packages form a cycle, all members must receive the same version bump.
// The highest priority bump among cycle members is selected and applied to all.
//
// Prerequisites: FindStronglyConnectedComponents must be called on the graph first
// to populate SCC IDs.
//
// Algorithm:
//  1. Group packages by their SCC ID
//  2. For each SCC with direct bumps:
//     - Find the highest priority bump among members
//     - Apply that bump to ALL members of the SCC
//  3. Mark cycle-resolved bumps with source="cycle"
//  4. Non-cycle packages keep source="direct"
func ResolveCycleBumps(
	g *graph.DependencyGraph,
	currentVersions map[string]semver.Version,
	directBumps map[string]string,
) (map[string]VersionBump, error) {
	if len(directBumps) == 0 {
		return make(map[string]VersionBump), nil
	}

	result := make(map[string]VersionBump)

	// Group packages by SCC ID
	sccGroups := make(map[int][]string)
	for _, node := range g.GetAllNodes() {
		sccGroups[node.SCC] = append(sccGroups[node.SCC], node.Package.Name)
	}

	// Process each SCC
	for _, members := range sccGroups {
		// Find all direct bumps in this SCC
		bumpsInSCC := make(map[string]string)
		for _, member := range members {
			if bump, hasBump := directBumps[member]; hasBump {
				bumpsInSCC[member] = bump
			}
		}

		// If no bumps in this SCC, skip it
		if len(bumpsInSCC) == 0 {
			continue
		}

		// Determine if this is a real cycle (multiple members OR single member that's a cycle)
		isCycle := len(members) > 1 || (len(members) == 1 && isSelfCycle(g, members[0]))

		if isCycle {
			// Find highest priority bump among cycle members
			highestBump := findHighestPriorityBump(bumpsInSCC)

			// Apply highest bump to ALL members of the cycle
			for _, member := range members {
				currentVer, ok := currentVersions[member]
				if !ok {
					return nil, fmt.Errorf("missing current version for package: %s", member)
				}

				newVer, err := currentVer.Bump(highestBump)
				if err != nil {
					return nil, fmt.Errorf("failed to bump version for %s: %w", member, err)
				}

				result[member] = VersionBump{
					Package:    member,
					OldVersion: currentVer,
					NewVersion: newVer,
					ChangeType: highestBump,
					Source:     "cycle",
				}
			}
		} else {
			// Not a cycle - apply direct bumps with "direct" source
			for pkgName, changeType := range bumpsInSCC {
				currentVer, ok := currentVersions[pkgName]
				if !ok {
					return nil, fmt.Errorf("missing current version for package: %s", pkgName)
				}

				newVer, err := currentVer.Bump(changeType)
				if err != nil {
					return nil, fmt.Errorf("failed to bump version for %s: %w", pkgName, err)
				}

				result[pkgName] = VersionBump{
					Package:    pkgName,
					OldVersion: currentVer,
					NewVersion: newVer,
					ChangeType: changeType,
					Source:     "direct",
				}
			}
		}
	}

	return result, nil
}

// findHighestPriorityBump returns the change type with the highest priority
// from a map of package names to change types.
func findHighestPriorityBump(bumps map[string]string) string {
	highestPriority := 0
	highestBump := "patch"

	for _, changeType := range bumps {
		priority := GetChangePriority(changeType)
		if priority > highestPriority {
			highestPriority = priority
			highestBump = changeType
		}
	}

	return highestBump
}

// isSelfCycle checks if a package has a dependency on itself
func isSelfCycle(g *graph.DependencyGraph, pkgName string) bool {
	edges := g.GetEdgesFrom(pkgName)
	for _, edge := range edges {
		if edge.To == pkgName {
			return true
		}
	}
	return false
}
