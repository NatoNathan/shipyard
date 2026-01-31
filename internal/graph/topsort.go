package graph

import (
	"fmt"
)

// TopologicalSort performs a topological sort on a compressed graph using Kahn's algorithm.
// Returns nodes in dependency order (dependencies before dependents).
// The compressed graph must be a DAG (cycles should be compressed first).
// Returns an error if a cycle is detected (should not happen with properly compressed graph).
func TopologicalSort(cg *CompressedGraph) ([]*CompressedNode, error) {
	if cg == nil || cg.GetNodeCount() == 0 {
		return []*CompressedNode{}, nil
	}

	// Calculate in-degree for each node
	inDegree := make(map[string]int)
	for _, node := range cg.GetAllNodes() {
		inDegree[node.Name] = 0
	}

	for _, node := range cg.GetAllNodes() {
		edges := cg.GetEdgesFrom(node.Name)
		for _, edge := range edges {
			inDegree[edge.To]++
		}
	}

	// Initialize queue with nodes that have no incoming edges
	queue := []string{}
	for nodeName, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeName)
		}
	}

	// Process nodes in topological order
	sorted := []*CompressedNode{}
	for len(queue) > 0 {
		// Pop from queue
		current := queue[0]
		queue = queue[1:]

		// Add to result
		node, _ := cg.GetNode(current)
		sorted = append(sorted, node)

		// Reduce in-degree of neighbors
		edges := cg.GetEdgesFrom(current)
		for _, edge := range edges {
			inDegree[edge.To]--
			if inDegree[edge.To] == 0 {
				queue = append(queue, edge.To)
			}
		}
	}

	// Check if all nodes were processed (if not, there's a cycle)
	if len(sorted) != cg.GetNodeCount() {
		return nil, fmt.Errorf("cycle detected in compressed graph: sorted %d nodes but graph has %d nodes",
			len(sorted), cg.GetNodeCount())
	}

	// Reverse the result to get dependencies before dependents
	// (Our edges go FROM dependent TO dependency, so Kahn's gives us dependents first)
	reversed := make([]*CompressedNode, len(sorted))
	for i, node := range sorted {
		reversed[len(sorted)-1-i] = node
	}

	return reversed, nil
}
