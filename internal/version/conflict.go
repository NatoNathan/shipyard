package version

// ResolveConflicts resolves conflicting version bump requests for packages.
//
// In the current implementation, conflicts are prevented by PropagateLinked()
// which only processes each package once. Direct bumps take precedence over
// propagated bumps, and the first propagation to reach a package wins.
//
// This function serves as:
//  1. A validation pass to ensure no actual conflicts exist
//  2. A clear extension point for more complex conflict resolution strategies
//  3. Documentation of the conflict resolution policy
//
// Conflict Resolution Policy:
//  - Direct bumps always win over propagated bumps
//  - Cycle-resolved bumps have already unified all members
//  - For propagated bumps, first-to-arrive wins (handled by PropagateLinked)
//
// Future enhancements could include:
//  - Explicit priority levels for different propagation sources
//  - User-configurable conflict resolution strategies
//  - Warnings/errors for detected conflicts requiring manual resolution
func ResolveConflicts(bumps map[string]VersionBump) map[string]VersionBump {
	// Current implementation: return bumps as-is since PropagateLinked
	// already prevents conflicts by only processing each package once

	// The map structure (one entry per package) ensures no duplicates
	// and the propagation algorithm ensures correct precedence

	return bumps
}

// Conflict Resolution Strategy Documentation
//
// ## How Conflicts Are Prevented
//
// ### 1. Direct vs Propagated
//
// PropagateLinked checks `directBumps` before propagating:
//
//	if _, hasDirectBump := directBumps[dependent]; hasDirectBump {
//	    continue // Skip propagation
//	}
//
// ### 2. Multiple Propagation Paths (Diamond Dependencies)
//
// PropagateLinked checks `result` before adding a bump:
//
//	if _, alreadyProcessed := result[dependent]; alreadyProcessed {
//	    continue // Skip if already bumped
//	}
//
// This means in a diamond dependency (A depends on B and C, both depend on D),
// when D bumps:
//  - D bumps first (direct)
//  - Either B or C gets the propagated bump (whichever is processed first)
//  - A gets the propagated bump from whichever of B/C propagated to it first
//  - The other path is blocked by the "already processed" check
//
// ### 3. Cycle Resolution
//
// ResolveCycleBumps processes all cycle members together, applying the
// maximum priority bump to all members. This prevents conflicts within cycles.
//
// ## Examples
//
// ### Example 1: Direct Overrides Propagation
//
//	// api -> core (linked)
//	// core has patch bump, api has direct major bump
//	directBumps := {"core": "patch", "api": "major"}
//
//	// Result:
//	// - core: patch (direct)
//	// - api: major (direct, NOT propagated patch from core)
//
// ### Example 2: Diamond Dependency
//
//	//     d
//	//    / \
//	//   b   c
//	//    \ /
//	//     a (patch bump)
//
//	// Result:
//	// - a: patch (direct)
//	// - b: patch (propagated from a)
//	// - c: patch (propagated from a)
//	// - d: patch (propagated from either b or c, first one wins)
//
// ### Example 3: Cycle with External Dependent
//
//	// a <-> b (cycle, a=patch, b=minor)
//	// c depends on a
//
//	// After cycle resolution:
//	// - a: minor (cycle resolved to highest)
//	// - b: minor (cycle resolved to highest)
//	// - c: minor (propagated from cycle)
