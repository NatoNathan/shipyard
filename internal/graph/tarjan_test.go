package graph

import (
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindStronglyConnectedComponents(t *testing.T) {
	t.Run("no cycles - each node in its own SCC", func(t *testing.T) {
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

		sccs := FindStronglyConnectedComponents(g)

		// Each node should be in its own SCC (3 SCCs total)
		assert.Len(t, sccs, 3)

		// Each SCC should have exactly 1 node
		for _, scc := range sccs {
			assert.Len(t, scc, 1)
		}

		// Verify SCC IDs are set in graph nodes
		for _, pkg := range []string{"utils", "core", "api"} {
			node, exists := g.GetNode(pkg)
			assert.True(t, exists)
			assert.NotEqual(t, 0, node.SCC, "SCC should be set for %s", pkg)
		}
	})

	t.Run("simple cycle - two nodes", func(t *testing.T) {
		// a -> b -> a
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

		sccs := FindStronglyConnectedComponents(g)

		// Should have 1 SCC containing both nodes
		assert.Len(t, sccs, 1)
		assert.Len(t, sccs[0], 2)

		// Both nodes should have the same SCC ID
		nodeA, _ := g.GetNode("a")
		nodeB, _ := g.GetNode("b")
		assert.Equal(t, nodeA.SCC, nodeB.SCC)
		assert.NotEqual(t, 0, nodeA.SCC)
	})

	t.Run("self-cycle - single node", func(t *testing.T) {
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

		sccs := FindStronglyConnectedComponents(g)

		// Should have 1 SCC containing the single node
		assert.Len(t, sccs, 1)
		assert.Len(t, sccs[0], 1)
		assert.Equal(t, "a", sccs[0][0])

		// Node should have SCC ID set
		nodeA, _ := g.GetNode("a")
		assert.NotEqual(t, 0, nodeA.SCC)
	})

	t.Run("complex cycle - three nodes", func(t *testing.T) {
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

		sccs := FindStronglyConnectedComponents(g)

		// Should have 1 SCC containing all three nodes
		assert.Len(t, sccs, 1)
		assert.Len(t, sccs[0], 3)

		// All nodes should have the same SCC ID
		nodeA, _ := g.GetNode("a")
		nodeB, _ := g.GetNode("b")
		nodeC, _ := g.GetNode("c")
		assert.Equal(t, nodeA.SCC, nodeB.SCC)
		assert.Equal(t, nodeB.SCC, nodeC.SCC)
		assert.NotEqual(t, 0, nodeA.SCC)
	})

	t.Run("multiple cycles - separate SCCs", func(t *testing.T) {
		// Cycle 1: a <-> b
		// Cycle 2: c <-> d
		// e (no cycle)
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

		sccs := FindStronglyConnectedComponents(g)

		// Should have 3 SCCs: {a,b}, {c,d}, {e}
		assert.Len(t, sccs, 3)

		// Find SCCs by size
		var scc2Count, scc1Count int
		for _, scc := range sccs {
			if len(scc) == 2 {
				scc2Count++
			} else if len(scc) == 1 {
				scc1Count++
			}
		}
		assert.Equal(t, 2, scc2Count, "Should have 2 SCCs with 2 nodes")
		assert.Equal(t, 1, scc1Count, "Should have 1 SCC with 1 node")

		// a and b should share SCC ID
		nodeA, _ := g.GetNode("a")
		nodeB, _ := g.GetNode("b")
		assert.Equal(t, nodeA.SCC, nodeB.SCC)

		// c and d should share SCC ID (different from a,b)
		nodeC, _ := g.GetNode("c")
		nodeD, _ := g.GetNode("d")
		assert.Equal(t, nodeC.SCC, nodeD.SCC)
		assert.NotEqual(t, nodeA.SCC, nodeC.SCC)

		// e should have unique SCC ID
		nodeE, _ := g.GetNode("e")
		assert.NotEqual(t, nodeE.SCC, nodeA.SCC)
		assert.NotEqual(t, nodeE.SCC, nodeC.SCC)
	})

	t.Run("complex graph with mixed cycles and acyclic paths", func(t *testing.T) {
		// Graph structure:
		//   a -> b -> c (acyclic chain)
		//   d <-> e   (cycle)
		//   f -> d    (points to cycle)
		//   b -> e    (acyclic node points to cycle)
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
						{Package: "e", Strategy: "linked"},
					},
				},
				{Name: "c", Path: "./c", Ecosystem: config.EcosystemGo},
				{Name: "d", Path: "./d", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "e", Strategy: "linked"},
					},
				},
				{Name: "e", Path: "./e", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "d", Strategy: "linked"},
					},
				},
				{Name: "f", Path: "./f", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "d", Strategy: "linked"},
					},
				},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)

		sccs := FindStronglyConnectedComponents(g)

		// Should have 5 SCCs: {d,e} (cycle), and {a}, {b}, {c}, {f} (individuals)
		assert.Len(t, sccs, 5)

		// d and e should be in same SCC
		nodeD, _ := g.GetNode("d")
		nodeE, _ := g.GetNode("e")
		assert.Equal(t, nodeD.SCC, nodeE.SCC)

		// All other nodes should have unique SCCs
		nodeA, _ := g.GetNode("a")
		nodeB, _ := g.GetNode("b")
		nodeC, _ := g.GetNode("c")
		nodeF, _ := g.GetNode("f")

		uniqueSCCs := make(map[int]bool)
		uniqueSCCs[nodeA.SCC] = true
		uniqueSCCs[nodeB.SCC] = true
		uniqueSCCs[nodeC.SCC] = true
		uniqueSCCs[nodeD.SCC] = true
		uniqueSCCs[nodeF.SCC] = true

		assert.Equal(t, 5, len(uniqueSCCs), "Should have 5 distinct SCC IDs")
	})

	t.Run("empty graph", func(t *testing.T) {
		g := NewGraph()
		sccs := FindStronglyConnectedComponents(g)
		assert.Empty(t, sccs)
	})

	t.Run("single node no edges", func(t *testing.T) {
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "solo", Path: "./solo", Ecosystem: config.EcosystemGo},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)

		sccs := FindStronglyConnectedComponents(g)

		assert.Len(t, sccs, 1)
		assert.Len(t, sccs[0], 1)
		assert.Equal(t, "solo", sccs[0][0])

		node, _ := g.GetNode("solo")
		assert.NotEqual(t, 0, node.SCC)
	})
}
