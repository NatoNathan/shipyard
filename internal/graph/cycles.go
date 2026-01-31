package graph

// DetectCycles identifies cycles in the dependency graph using Tarjan's algorithm.
// Returns:
//   - hasCycles: true if any cycles exist in the graph
//   - cycles: slice of cycles, where each cycle is a slice of package names
//
// A cycle is defined as an SCC with more than one node, or a single node with a self-loop.
func DetectCycles(g *DependencyGraph) (bool, [][]string) {
	if g == nil || g.GetNodeCount() == 0 {
		return false, [][]string{}
	}

	// Run Tarjan's algorithm to find all SCCs
	sccs := FindStronglyConnectedComponents(g)

	// Filter SCCs to find actual cycles
	cycles := [][]string{}
	for _, scc := range sccs {
		if isCycle(g, scc) {
			cycles = append(cycles, scc)
		}
	}

	return len(cycles) > 0, cycles
}

// isCycle determines if an SCC represents an actual cycle
// A cycle is either:
//   1. An SCC with more than one node (multi-node cycle)
//   2. A single node with a self-edge (self-loop)
func isCycle(g *DependencyGraph, scc []string) bool {
	// Multi-node SCC is always a cycle
	if len(scc) > 1 {
		return true
	}

	// Single node - check for self-loop
	if len(scc) == 1 {
		nodeName := scc[0]
		edges := g.GetEdgesFrom(nodeName)
		for _, edge := range edges {
			if edge.To == nodeName {
				return true // Self-loop detected
			}
		}
	}

	return false
}
