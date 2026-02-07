package version

import (
	"fmt"

	"github.com/NatoNathan/shipyard/internal/consignment"
	"github.com/NatoNathan/shipyard/internal/graph"
	"github.com/NatoNathan/shipyard/pkg/semver"
)

// Propagator handles version bump propagation through a dependency graph
type Propagator struct {
	graph *graph.DependencyGraph
}

// VersionBump represents a version change for a package
type VersionBump struct {
	Package    string         // Package name
	OldVersion semver.Version // Current version
	NewVersion semver.Version // New version after bump
	ChangeType string         // Type of change: "patch", "minor", or "major"
	Source     string         // Source of bump: "direct", "propagated", "cycle"
}

// NewPropagator creates a new version propagator for the given dependency graph
func NewPropagator(g *graph.DependencyGraph) (*Propagator, error) {
	if g == nil {
		return nil, fmt.Errorf("dependency graph cannot be nil")
	}

	return &Propagator{
		graph: g,
	}, nil
}

// Propagate calculates version bumps for all packages based on consignments
// and dependency relationships. Returns a map of package name to VersionBump.
//
// Algorithm:
//  1. Calculate direct bumps from consignments
//  2. Run SCC detection to identify cycles
//  3. Resolve cycle bumps (unify bump types within each SCC)
//  4. Propagate bumps through dependencies (respecting strategies)
//  5. Resolve conflicts (multiple sources requesting different bumps)
func (p *Propagator) Propagate(
	currentVersions map[string]semver.Version,
	consignments []*consignment.Consignment,
) (map[string]VersionBump, error) {
	// Calculate direct bumps from consignments
	directBumps := CalculateDirectBumps(consignments)

	if len(directBumps) == 0 {
		return make(map[string]VersionBump), nil
	}

	// Run SCC detection to identify cycles in the graph
	graph.FindStronglyConnectedComponents(p.graph)

	// Resolve cycle bumps: unify bump types within each SCC
	cycleResolved, err := ResolveCycleBumps(p.graph, currentVersions, directBumps)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve cycle bumps: %w", err)
	}

	// Build the effective direct bumps map from cycle resolution
	// This includes cycle-unified bumps and unchanged direct bumps
	effectiveBumps := make(map[string]string)
	for pkg, bump := range cycleResolved {
		effectiveBumps[pkg] = bump.ChangeType
	}
	// Also include any direct bumps not processed by cycle resolution
	// (packages not in any SCC with bumps)
	for pkg, changeType := range directBumps {
		if _, exists := effectiveBumps[pkg]; !exists {
			effectiveBumps[pkg] = changeType
		}
	}

	// Propagate through linked dependencies (includes direct bumps)
	result, err := PropagateLinked(p.graph, currentVersions, effectiveBumps)
	if err != nil {
		return nil, err
	}

	// Preserve cycle source markers from ResolveCycleBumps
	for pkg, bump := range cycleResolved {
		if bump.Source == "cycle" {
			if r, ok := result[pkg]; ok {
				r.Source = "cycle"
				result[pkg] = r
			}
		}
	}

	// Resolve any remaining conflicts
	result = ResolveConflicts(result)

	return result, nil
}

