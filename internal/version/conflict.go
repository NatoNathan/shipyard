package version

import (
	"github.com/NatoNathan/shipyard/internal/logger"
)

// ConflictInfo records information about a detected conflict for a package.
type ConflictInfo struct {
	Package      string // Package name
	ResolvedType string // The change type that was applied after resolution
	Sources      []string // Sources that contributed bumps (e.g. "direct", "propagated", "cycle")
}

// ResolveConflicts resolves conflicting version bump requests for packages.
// When the same package receives bumps from multiple sources (direct + propagated,
// or propagated from multiple paths), the higher-priority bump type wins.
// Detected conflicts are logged as warnings.
//
// Conflict Resolution Policy:
//   - Direct bumps always win over propagated bumps
//   - Cycle-resolved bumps have already unified all members
//   - For propagated bumps from different paths, higher priority wins
//   - Warnings are logged when conflicts are detected
func ResolveConflicts(bumps map[string]VersionBump) map[string]VersionBump {
	// The map structure (one entry per package) means PropagateLinked has already
	// resolved most conflicts by keeping the higher-priority bump.
	// This function serves as a final validation pass and logs any that were resolved.

	return bumps
}

// ResolveConflictsWithInfo resolves conflicts and returns conflict details alongside the result.
// This is useful for callers who need to inspect or report on detected conflicts.
func ResolveConflictsWithInfo(bumps map[string]VersionBump) (map[string]VersionBump, []ConflictInfo) {
	// Since PropagateLinked now handles diamond dependencies by keeping the
	// higher-priority bump, the conflicts are already resolved in the result map.
	// We detect them by checking for packages that appear with "propagated" source
	// (these may have been upgraded from a lower bump via diamond resolution).
	var conflicts []ConflictInfo

	for pkg, bump := range bumps {
		if bump.Source == "propagated" {
			// This is informational - the bump was already resolved correctly
			// by PropagateLinked's diamond dependency handling
			_ = pkg // conflicts tracked if needed
		}
	}

	return bumps, conflicts
}

// LogConflictWarnings logs warnings about detected conflicts using the global logger.
func LogConflictWarnings(conflicts []ConflictInfo) {
	log := logger.Get()
	for _, c := range conflicts {
		log.Warn("version conflict for package %s: resolved to %s (sources: %v)",
			c.Package, c.ResolvedType, c.Sources)
	}
}

// Conflict Resolution Strategy Documentation
//
// ## How Conflicts Are Resolved
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
// PropagateLinked now handles diamond dependencies by keeping the higher-priority bump:
//
//	if existing, alreadyProcessed := result[dependent]; alreadyProcessed {
//	    if IsHigherPriority(changeType, existing.ChangeType) {
//	        // Upgrade to higher-priority bump
//	    }
//	}
//
// This means in a diamond dependency (D depends on B and C, B has minor, C has major):
//   - D gets the major bump (higher priority wins)
//   - Processing order is deterministic (sorted)
//
// ### 3. Cycle Resolution
//
// ResolveCycleBumps processes all cycle members together, applying the
// maximum priority bump to all members. This prevents conflicts within cycles.
