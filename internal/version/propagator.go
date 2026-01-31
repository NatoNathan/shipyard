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
//  2. Run SCC detection and compress graph
//  3. Topologically sort compressed graph
//  4. Propagate bumps through dependencies (respecting strategies)
//  5. Resolve conflicts (multiple sources requesting different bumps)
func (p *Propagator) Propagate(
	currentVersions map[string]semver.Version,
	consignments []*consignment.Consignment,
) (map[string]VersionBump, error) {
	// Calculate direct bumps from consignments
	directBumps := CalculateDirectBumps(consignments)

	// Propagate through linked dependencies (includes direct bumps)
	result, err := PropagateLinked(p.graph, currentVersions, directBumps)
	if err != nil {
		return nil, err
	}

	// Note: Advanced dependency resolution available but not currently integrated:
	// - ResolveCycleBumps() in cycle.go handles circular dependencies
	// - ResolveConflicts() in conflict.go handles diamond dependency conflicts
	// Current implementation works correctly for acyclic graphs and simple propagation.
	// The graph's FindStronglyConnectedComponents() would need to be called before
	// integrating cycle resolution. See tasks.md T099-T102 for full design.

	return result, nil
}

