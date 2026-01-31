package graph

import (
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressGraph(t *testing.T) {
	t.Run("no cycles - graph unchanged", func(t *testing.T) {
		// Linear dependency chain: api -> core -> utils
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "utils", Path: "./utils", Ecosystem: config.EcosystemGo},
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "utils", Strategy: "linked"},
					},
				},
				{Name: "api", Path: "./api", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "core", Strategy: "linked"},
					},
				},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)

		// Run SCC detection first
		FindStronglyConnectedComponents(g)

		compressed := CompressGraph(g)

		// Should have same number of nodes (no compression needed)
		assert.Equal(t, g.GetNodeCount(), compressed.GetNodeCount())
		assert.Equal(t, 3, compressed.GetNodeCount())

		// All original nodes should exist
		for _, name := range []string{"utils", "core", "api"} {
			node, exists := compressed.GetNode(name)
			assert.True(t, exists, "Node %s should exist", name)
			assert.Len(t, node.Members, 1)
			assert.Equal(t, name, node.Members[0])
		}

		// Edges should be preserved
		edges := compressed.GetEdgesFrom("api")
		assert.Len(t, edges, 1)
		assert.Equal(t, "core", edges[0].To)

		edges = compressed.GetEdgesFrom("core")
		assert.Len(t, edges, 1)
		assert.Equal(t, "utils", edges[0].To)
	})

	t.Run("simple cycle - compressed to meta-node", func(t *testing.T) {
		// a <-> b
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "a", Path: "./a", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "b", Strategy: "linked"},
					},
				},
				{Name: "b", Path: "./b", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "a", Strategy: "linked"},
					},
				},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)

		FindStronglyConnectedComponents(g)
		compressed := CompressGraph(g)

		// Should have 1 meta-node
		assert.Equal(t, 1, compressed.GetNodeCount())

		// Find the meta-node (should be named with SCC ID)
		nodes := compressed.GetAllNodes()
		require.Len(t, nodes, 1)
		metaNode := nodes[0]

		// Meta-node should contain both members
		assert.Len(t, metaNode.Members, 2)
		memberSet := make(map[string]bool)
		for _, member := range metaNode.Members {
			memberSet[member] = true
		}
		assert.True(t, memberSet["a"])
		assert.True(t, memberSet["b"])

		// Should have no edges (cycle is internal)
		assert.Empty(t, compressed.GetEdgesFrom(metaNode.Name))
	})

	t.Run("cycle with external dependencies", func(t *testing.T) {
		// a <-> b, both depend on c
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "a", Path: "./a", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "b", Strategy: "linked"},
						{Package: "c", Strategy: "linked"},
					},
				},
				{Name: "b", Path: "./b", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "a", Strategy: "linked"},
						{Package: "c", Strategy: "linked"},
					},
				},
				{Name: "c", Path: "./c", Ecosystem: config.EcosystemGo},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)

		FindStronglyConnectedComponents(g)
		compressed := CompressGraph(g)

		// Should have 2 nodes: meta-node for {a,b} and c
		assert.Equal(t, 2, compressed.GetNodeCount())

		// Find meta-node and c
		var metaNode, cNode *CompressedNode
		for _, node := range compressed.GetAllNodes() {
			if len(node.Members) > 1 {
				metaNode = node
			} else if len(node.Members) == 1 && node.Members[0] == "c" {
				cNode = node
			}
		}

		require.NotNil(t, metaNode)
		require.NotNil(t, cNode)
		assert.Len(t, metaNode.Members, 2)

		// Meta-node should have edge to c
		edges := compressed.GetEdgesFrom(metaNode.Name)
		assert.Len(t, edges, 1)
		assert.Equal(t, cNode.Name, edges[0].To)
	})

	t.Run("external node depends on cycle", func(t *testing.T) {
		// d -> a <-> b
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "a", Path: "./a", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "b", Strategy: "linked"},
					},
				},
				{Name: "b", Path: "./b", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "a", Strategy: "linked"},
					},
				},
				{Name: "d", Path: "./d", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "a", Strategy: "linked"},
					},
				},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)

		FindStronglyConnectedComponents(g)
		compressed := CompressGraph(g)

		// Should have 2 nodes: meta-node for {a,b} and d
		assert.Equal(t, 2, compressed.GetNodeCount())

		// d should have edge to meta-node
		dNode, exists := compressed.GetNode("d")
		require.True(t, exists)
		assert.Len(t, dNode.Members, 1)

		edges := compressed.GetEdgesFrom("d")
		assert.Len(t, edges, 1)
		// Edge should point to meta-node
		toNode, _ := compressed.GetNode(edges[0].To)
		assert.Len(t, toNode.Members, 2)
	})

	t.Run("multiple cycles - multiple meta-nodes", func(t *testing.T) {
		// Cycle 1: a <-> b
		// Cycle 2: c <-> d
		// e (independent)
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "a", Path: "./a", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "b", Strategy: "linked"},
					},
				},
				{Name: "b", Path: "./b", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "a", Strategy: "linked"},
					},
				},
				{Name: "c", Path: "./c", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "d", Strategy: "linked"},
					},
				},
				{Name: "d", Path: "./d", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "c", Strategy: "linked"},
					},
				},
				{Name: "e", Path: "./e", Ecosystem: config.EcosystemGo},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)

		FindStronglyConnectedComponents(g)
		compressed := CompressGraph(g)

		// Should have 3 nodes: 2 meta-nodes and e
		assert.Equal(t, 3, compressed.GetNodeCount())

		// Count nodes by member size
		var metaNodeCount, singleNodeCount int
		for _, node := range compressed.GetAllNodes() {
			if len(node.Members) == 2 {
				metaNodeCount++
			} else if len(node.Members) == 1 {
				singleNodeCount++
			}
		}

		assert.Equal(t, 2, metaNodeCount, "Should have 2 meta-nodes")
		assert.Equal(t, 1, singleNodeCount, "Should have 1 single node")
	})

	t.Run("complex cycle - three nodes compressed", func(t *testing.T) {
		// a -> b -> c -> a
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "a", Path: "./a", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "b", Strategy: "linked"},
					},
				},
				{Name: "b", Path: "./b", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "c", Strategy: "linked"},
					},
				},
				{Name: "c", Path: "./c", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "a", Strategy: "linked"},
					},
				},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)

		FindStronglyConnectedComponents(g)
		compressed := CompressGraph(g)

		// Should have 1 meta-node with 3 members
		assert.Equal(t, 1, compressed.GetNodeCount())

		node := compressed.GetAllNodes()[0]
		assert.Len(t, node.Members, 3)

		memberSet := make(map[string]bool)
		for _, member := range node.Members {
			memberSet[member] = true
		}
		assert.True(t, memberSet["a"])
		assert.True(t, memberSet["b"])
		assert.True(t, memberSet["c"])
	})

	t.Run("empty graph", func(t *testing.T) {
		g := NewGraph()
		FindStronglyConnectedComponents(g)
		compressed := CompressGraph(g)

		assert.Equal(t, 0, compressed.GetNodeCount())
	})

	t.Run("self-cycle compressed", func(t *testing.T) {
		// a -> a
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "a", Path: "./a", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "a", Strategy: "linked"},
					},
				},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)

		FindStronglyConnectedComponents(g)
		compressed := CompressGraph(g)

		// Should have 1 meta-node
		assert.Equal(t, 1, compressed.GetNodeCount())

		node := compressed.GetAllNodes()[0]
		assert.Len(t, node.Members, 1)
		assert.Equal(t, "a", node.Members[0])

		// Should have no external edges (self-loop removed)
		assert.Empty(t, compressed.GetEdgesFrom(node.Name))
	})
}
