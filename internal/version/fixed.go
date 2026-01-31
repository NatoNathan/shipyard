package version

// Fixed Dependency Handling
//
// This file documents the behavior of "fixed" dependency strategy in version propagation.
//
// ## Strategy Overview
//
// The "fixed" strategy indicates that a package's version is independent of its dependencies'
// versions. When a dependency is marked as "fixed", version bumps in that dependency do NOT
// propagate to the dependent package.
//
// This is useful for:
// - Stable APIs where consumers control their upgrade timeline
// - Internal tooling that should version independently
// - Breaking the propagation chain in specific places
//
// ## Implementation
//
// Fixed dependencies are handled in PropagateLinked() by checking the edge strategy:
//
//	if edge.Strategy != "linked" {
//	    continue // Skip propagation for non-linked dependencies
//	}
//
// Only "linked" dependencies participate in version propagation. All other strategies
// (including "fixed" and any unrecognized strategies) block propagation.
//
// ## Default Strategy
//
// When no strategy is specified in the configuration, BuildGraph() defaults to "linked":
//
//	strategy := dep.Strategy
//	if strategy == "" {
//	    strategy = "linked" // Default strategy
//	}
//
// This ensures backward compatibility and sensible default behavior.
//
// ## Examples
//
// ### Fixed Dependency
//
//	// web depends on api with fixed strategy
//	web -> api (fixed)
//
//	// If api bumps from 1.0.0 to 2.0.0 (major breaking change):
//	// - api: 1.0.0 -> 2.0.0 ✓
//	// - web: stays at current version (no propagation)
//
// ### Mixed Strategies
//
//	// Multiple packages depend on core with different strategies
//	api  -> core (linked)  // Will receive propagated bumps
//	web  -> core (fixed)   // Will NOT receive propagated bumps
//
//	// If core bumps:
//	// - core: bumped ✓
//	// - api: bumped ✓ (linked)
//	// - web: unchanged (fixed)
//
// ### Transitive Blocking
//
//	// Chain with fixed in the middle
//	app -> service (fixed) -> lib (linked)
//
//	// If lib bumps:
//	// - lib: bumped ✓
//	// - service: bumped ✓ (linked to lib)
//	// - app: unchanged (fixed blocks transitive propagation)
//
// ## See Also
//
// - PropagateLinked() in propagate.go for the actual implementation
// - BuildGraph() in graph/build.go for default strategy handling
// - TestFixedDependencyHandling in fixed_test.go for comprehensive test coverage
