package graph

// FindStronglyConnectedComponents uses Tarjan's algorithm to identify
// strongly connected components (SCCs) in the dependency graph.
// Returns a slice of SCCs, where each SCC is a slice of package names.
// Also sets the SCC field on each node to its component ID.
func FindStronglyConnectedComponents(g *DependencyGraph) [][]string {
	if g == nil || len(g.nodes) == 0 {
		return [][]string{}
	}

	state := &tarjanState{
		graph:    g,
		index:    0,
		indices:  make(map[string]int),
		lowlinks: make(map[string]int),
		onStack:  make(map[string]bool),
		stack:    []string{},
		sccs:     [][]string{},
		sccID:    1, // Start SCC IDs at 1 (0 means not in cycle)
	}

	// Run algorithm for each unvisited node
	for name := range g.nodes {
		if _, visited := state.indices[name]; !visited {
			state.strongConnect(name)
		}
	}

	return state.sccs
}

// tarjanState holds the state for Tarjan's algorithm
type tarjanState struct {
	graph    *DependencyGraph
	index    int
	indices  map[string]int
	lowlinks map[string]int
	onStack  map[string]bool
	stack    []string
	sccs     [][]string
	sccID    int
}

// strongConnect is the recursive heart of Tarjan's algorithm
func (s *tarjanState) strongConnect(name string) {
	// Set the depth index for this node
	s.indices[name] = s.index
	s.lowlinks[name] = s.index
	s.index++
	s.stack = append(s.stack, name)
	s.onStack[name] = true

	// Consider successors (dependencies) of this node
	edges := s.graph.GetEdgesFrom(name)
	for _, edge := range edges {
		successor := edge.To

		if _, visited := s.indices[successor]; !visited {
			// Successor has not yet been visited; recurse on it
			s.strongConnect(successor)
			s.lowlinks[name] = min(s.lowlinks[name], s.lowlinks[successor])
		} else if s.onStack[successor] {
			// Successor is on stack and hence in the current SCC
			// If successor is not on stack, then (name, successor) is a cross-edge
			// in the DFS tree and must be ignored
			s.lowlinks[name] = min(s.lowlinks[name], s.indices[successor])
		}
	}

	// If name is a root node, pop the stack and generate an SCC
	if s.lowlinks[name] == s.indices[name] {
		scc := []string{}
		for {
			w := s.stack[len(s.stack)-1]
			s.stack = s.stack[:len(s.stack)-1]
			s.onStack[w] = false
			scc = append(scc, w)

			// Set SCC ID on the node
			_ = s.graph.SetSCC(w, s.sccID)

			if w == name {
				break
			}
		}

		s.sccs = append(s.sccs, scc)
		s.sccID++
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
