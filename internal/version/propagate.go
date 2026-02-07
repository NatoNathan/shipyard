package version

import (
	"fmt"
	"sort"

	"github.com/NatoNathan/shipyard/internal/graph"
	"github.com/NatoNathan/shipyard/pkg/semver"
)

// PropagateLinked propagates version bumps through linked dependencies in the graph.
// Returns a map of all version bumps (both direct and propagated).
//
// Algorithm:
//  1. Start with direct bumps and apply them to result
//  2. Process changed packages in sorted order for deterministic results
//  3. For each changed package, find packages that depend on it
//  4. If dependency has "linked" strategy, propagate the bump (respecting bump mapping)
//  5. Skip packages that already have direct bumps (direct takes precedence)
//  6. When a package receives bumps from multiple paths, keep the higher-priority bump
//  7. Continue until no more changes propagate
func PropagateLinked(
	g *graph.DependencyGraph,
	currentVersions map[string]semver.Version,
	directBumps map[string]string,
) (map[string]VersionBump, error) {
	if len(directBumps) == 0 {
		return make(map[string]VersionBump), nil
	}

	result := make(map[string]VersionBump)

	// Apply direct bumps first
	for pkgName, changeType := range directBumps {
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

	// Collect initial changed packages into a sorted slice for deterministic order
	changed := make([]string, 0, len(directBumps))
	for pkg := range directBumps {
		changed = append(changed, pkg)
	}
	sort.Strings(changed)

	for len(changed) > 0 {
		// Process all packages in the current round
		var nextChanged []string
		nextChangedSet := make(map[string]bool)

		for _, changedPkg := range changed {
			// Find packages that depend on this one
			for _, node := range g.GetAllNodes() {
				edges := g.GetEdgesFrom(node.Package.Name)
				for _, edge := range edges {
					// If this edge points to our changed package
					if edge.To == changedPkg {
						dependent := node.Package.Name

						// Only propagate for linked strategy
						if edge.Strategy != "linked" {
							continue
						}

						// Skip if dependent already has a direct bump (direct takes precedence)
						if _, hasDirectBump := directBumps[dependent]; hasDirectBump {
							continue
						}

						// Get the change type from the dependency
						changeType := result[changedPkg].ChangeType

						// Apply bump mapping if present
						if edge.BumpMap != nil {
							if mapped, ok := edge.BumpMap[changeType]; ok {
								changeType = mapped
							}
						}

						// If already processed, keep the higher-priority bump (diamond dependency handling)
						if existing, alreadyProcessed := result[dependent]; alreadyProcessed {
							if existing.Source == "propagated" && IsHigherPriority(changeType, existing.ChangeType) {
								// Upgrade to higher-priority bump
								currentVer := currentVersions[dependent]
								newVer, err := currentVer.Bump(changeType)
								if err != nil {
									return nil, fmt.Errorf("failed to bump version for %s: %w", dependent, err)
								}
								result[dependent] = VersionBump{
									Package:    dependent,
									OldVersion: currentVer,
									NewVersion: newVer,
									ChangeType: changeType,
									Source:     "propagated",
								}
								// Re-propagate since bump increased
								if !nextChangedSet[dependent] {
									nextChanged = append(nextChanged, dependent)
									nextChangedSet[dependent] = true
								}
							}
							continue
						}

						// Apply the bump
						currentVer, ok := currentVersions[dependent]
						if !ok {
							return nil, fmt.Errorf("missing current version for package: %s", dependent)
						}

						newVer, err := currentVer.Bump(changeType)
						if err != nil {
							return nil, fmt.Errorf("failed to bump version for %s: %w", dependent, err)
						}

						result[dependent] = VersionBump{
							Package:    dependent,
							OldVersion: currentVer,
							NewVersion: newVer,
							ChangeType: changeType,
							Source:     "propagated",
						}

						// Mark as changed for further propagation
						if !nextChangedSet[dependent] {
							nextChanged = append(nextChanged, dependent)
							nextChangedSet[dependent] = true
						}
					}
				}
			}
		}

		// Sort next round for deterministic processing
		sort.Strings(nextChanged)
		changed = nextChanged
	}

	return result, nil
}
