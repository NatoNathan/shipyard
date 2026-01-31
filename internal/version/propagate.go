package version

import (
	"fmt"

	"github.com/NatoNathan/shipyard/internal/graph"
	"github.com/NatoNathan/shipyard/pkg/semver"
)

// PropagateLinked propagates version bumps through linked dependencies in the graph.
// Returns a map of all version bumps (both direct and propagated).
//
// Algorithm:
//  1. Start with direct bumps and apply them to result
//  2. For each changed package, find packages that depend on it
//  3. If dependency has "linked" strategy, propagate the bump (respecting bump mapping)
//  4. Skip packages that already have direct bumps (direct takes precedence)
//  5. Continue until no more changes propagate
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

	// Propagate changes through linked dependencies
	changed := make(map[string]bool)
	for pkg := range directBumps {
		changed[pkg] = true
	}

	for len(changed) > 0 {
		// Get one changed package
		var changedPkg string
		for pkg := range changed {
			changedPkg = pkg
			break
		}
		delete(changed, changedPkg)

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

					// Skip if already processed in this propagation
					if _, alreadyProcessed := result[dependent]; alreadyProcessed {
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
					changed[dependent] = true
				}
			}
		}
	}

	return result, nil
}
