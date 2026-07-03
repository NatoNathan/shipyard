package version

// Conflict Resolution Strategy Documentation
//
// This file documents how conflicting version bump requests are resolved.
// Conflicts are resolved inline during propagation and cycle resolution;
// there is no separate resolution pass.
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
// PropagateLinked handles diamond dependencies by keeping the higher-priority bump:
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
