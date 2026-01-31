package graph

import (
	"fmt"
	"sort"
	"strings"
)

// CompressedNode represents a node in the compressed graph
// It may represent a single package or a strongly connected component (cycle)
type CompressedNode struct {
	Name    string   // Meta-node name (for cycles: "scc_N", for singles: original name)
	Members []string // Package names in this node (1 for singles, multiple for cycles)
	SCC     int      // Original SCC ID
}

// CompressedEdge represents an edge in the compressed graph
type CompressedEdge struct {
	From     string
	To       string
	Strategy string            // Inherited from original edge
	BumpMap  map[string]string // Inherited from original edge
}

// CompressedGraph represents a dependency graph where SCCs are compressed into meta-nodes
type CompressedGraph struct {
	nodes map[string]*CompressedNode
	edges map[string][]CompressedEdge
}

// NewCompressedGraph creates a new empty compressed graph
func NewCompressedGraph() *CompressedGraph {
	return &CompressedGraph{
		nodes: make(map[string]*CompressedNode),
		edges: make(map[string][]CompressedEdge),
	}
}

// CompressGraph converts a dependency graph into a compressed graph
// where each SCC becomes a single meta-node.
// Prerequisites: FindStronglyConnectedComponents must be called first to set SCC IDs
func CompressGraph(g *DependencyGraph) *CompressedGraph {
	compressed := NewCompressedGraph()

	if g == nil || g.GetNodeCount() == 0 {
		return compressed
	}

	// Group nodes by SCC ID
	sccGroups := make(map[int][]string)
	for _, node := range g.GetAllNodes() {
		sccGroups[node.SCC] = append(sccGroups[node.SCC], node.Package.Name)
	}

	// Create mapping from original name to compressed node name
	nameMapping := make(map[string]string)

	// Create compressed nodes
	for sccID, members := range sccGroups {
		// Sort members for deterministic output
		sort.Strings(members)

		var nodeName string
		if len(members) == 1 {
			// Single node: use original name
			nodeName = members[0]
		} else {
			// Meta-node for cycle: use SCC-based name
			nodeName = fmt.Sprintf("scc_%d", sccID)
		}

		compressed.nodes[nodeName] = &CompressedNode{
			Name:    nodeName,
			Members: members,
			SCC:     sccID,
		}

		// Map each member to this compressed node
		for _, member := range members {
			nameMapping[member] = nodeName
		}

		// Initialize empty edge list
		compressed.edges[nodeName] = []CompressedEdge{}
	}

	// Create compressed edges
	// Only include edges between different SCCs (inter-SCC edges)
	addedEdges := make(map[string]bool) // Track edges to avoid duplicates

	for _, node := range g.GetAllNodes() {
		fromCompressed := nameMapping[node.Package.Name]
		edges := g.GetEdgesFrom(node.Package.Name)

		for _, edge := range edges {
			toCompressed := nameMapping[edge.To]

			// Skip edges within same SCC (intra-SCC edges)
			if fromCompressed == toCompressed {
				continue
			}

			// Create edge key for deduplication
			edgeKey := fromCompressed + "->" + toCompressed

			// Only add edge if not already added
			if !addedEdges[edgeKey] {
				compressed.edges[fromCompressed] = append(
					compressed.edges[fromCompressed],
					CompressedEdge{
						From:     fromCompressed,
						To:       toCompressed,
						Strategy: edge.Strategy,
						BumpMap:  edge.BumpMap,
					},
				)
				addedEdges[edgeKey] = true
			}
		}
	}

	return compressed
}

// GetNode returns the compressed node with the given name
func (cg *CompressedGraph) GetNode(name string) (*CompressedNode, bool) {
	node, exists := cg.nodes[name]
	return node, exists
}

// GetEdgesFrom returns all edges originating from the given node
func (cg *CompressedGraph) GetEdgesFrom(from string) []CompressedEdge {
	edges, exists := cg.edges[from]
	if !exists {
		return []CompressedEdge{}
	}
	return edges
}

// GetAllNodes returns all nodes in the compressed graph
func (cg *CompressedGraph) GetAllNodes() []*CompressedNode {
	nodes := make([]*CompressedNode, 0, len(cg.nodes))
	for _, node := range cg.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetNodeCount returns the number of nodes in the compressed graph
func (cg *CompressedGraph) GetNodeCount() int {
	return len(cg.nodes)
}

// GetEdgeCount returns the total number of edges in the compressed graph
func (cg *CompressedGraph) GetEdgeCount() int {
	count := 0
	for _, edges := range cg.edges {
		count += len(edges)
	}
	return count
}

// IsCycle returns true if the node represents a cycle (has multiple members)
func (cn *CompressedNode) IsCycle() bool {
	return len(cn.Members) > 1
}

// String returns a human-readable representation of the compressed node
func (cn *CompressedNode) String() string {
	if cn.IsCycle() {
		return fmt.Sprintf("%s[%s]", cn.Name, strings.Join(cn.Members, ", "))
	}
	return cn.Name
}
