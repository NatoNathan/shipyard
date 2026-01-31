package graph

import (
	"fmt"

	"github.com/NatoNathan/shipyard/internal/config"
)

// GraphNode represents a node in the dependency graph
type GraphNode struct {
	Package config.Package
	SCC     int // Strongly Connected Component ID (0 if not in cycle)
}

// GraphEdge represents a directed edge in the dependency graph
type GraphEdge struct {
	From     string
	To       string
	Strategy string                // "fixed" or "linked"
	BumpMap  map[string]string     // changeType -> changeType mapping
}

// DependencyGraph represents a directed graph of package dependencies
type DependencyGraph struct {
	nodes map[string]*GraphNode
	edges map[string][]GraphEdge
}

// NewGraph creates a new empty dependency graph
func NewGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes: make(map[string]*GraphNode),
		edges: make(map[string][]GraphEdge),
	}
}

// AddNode adds a package node to the graph
// Returns an error if a node with the same name already exists
func (g *DependencyGraph) AddNode(pkg config.Package) error {
	if _, exists := g.nodes[pkg.Name]; exists {
		return fmt.Errorf("node already exists: %s", pkg.Name)
	}

	g.nodes[pkg.Name] = &GraphNode{
		Package: pkg,
		SCC:     0, // Not in a cycle by default
	}

	// Initialize empty edge list
	if g.edges[pkg.Name] == nil {
		g.edges[pkg.Name] = []GraphEdge{}
	}

	return nil
}

// AddEdge adds a directed edge from one package to another
// Returns an error if either node doesn't exist
func (g *DependencyGraph) AddEdge(from, to string, strategy string, bumpMap map[string]string) error {
	// Verify both nodes exist
	if _, exists := g.nodes[from]; !exists {
		return fmt.Errorf("source node not found: %s", from)
	}
	if _, exists := g.nodes[to]; !exists {
		return fmt.Errorf("target node not found: %s", to)
	}

	edge := GraphEdge{
		From:     from,
		To:       to,
		Strategy: strategy,
		BumpMap:  bumpMap,
	}

	g.edges[from] = append(g.edges[from], edge)
	return nil
}

// GetNode returns the node with the given name, or nil if not found
func (g *DependencyGraph) GetNode(name string) (*GraphNode, bool) {
	node, exists := g.nodes[name]
	return node, exists
}

// GetEdgesFrom returns all edges originating from the given node
func (g *DependencyGraph) GetEdgesFrom(from string) []GraphEdge {
	edges, exists := g.edges[from]
	if !exists {
		return []GraphEdge{}
	}
	return edges
}

// GetAllNodes returns all nodes in the graph
func (g *DependencyGraph) GetAllNodes() []*GraphNode {
	nodes := make([]*GraphNode, 0, len(g.nodes))
	for _, node := range g.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// SetSCC sets the Strongly Connected Component ID for a node
func (g *DependencyGraph) SetSCC(name string, sccID int) error {
	node, exists := g.nodes[name]
	if !exists {
		return fmt.Errorf("node not found: %s", name)
	}
	node.SCC = sccID
	return nil
}

// GetNodeCount returns the number of nodes in the graph
func (g *DependencyGraph) GetNodeCount() int {
	return len(g.nodes)
}

// GetEdgeCount returns the total number of edges in the graph
func (g *DependencyGraph) GetEdgeCount() int {
	count := 0
	for _, edges := range g.edges {
		count += len(edges)
	}
	return count
}
