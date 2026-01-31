package graph

import (
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTopologicalSort(t *testing.T) {
	t.Run("linear dependency chain", func(t *testing.T) {
		// api -> core -> utils
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

		FindStronglyConnectedComponents(g)
		compressed := CompressGraph(g)

		sorted, err := TopologicalSort(compressed)
		require.NoError(t, err)
		assert.Len(t, sorted, 3)

		// Verify order: dependencies before dependents
		// utils should come before core, core before api
		positions := make(map[string]int)
		for i, node := range sorted {
			positions[node.Name] = i
		}

		assert.Less(t, positions["utils"], positions["core"])
		assert.Less(t, positions["core"], positions["api"])
	})

	t.Run("diamond dependency", func(t *testing.T) {
		// Diamond: a -> b -> d
		//          a -> c -> d
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "d", Path: "./d", Ecosystem: config.EcosystemGo},
				{Name: "b", Path: "./b", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "d", Strategy: "linked"},
					},
				},
				{Name: "c", Path: "./c", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "d", Strategy: "linked"},
					},
				},
				{Name: "a", Path: "./a", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "b", Strategy: "linked"},
						{Package: "c", Strategy: "linked"},
					},
				},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)

		FindStronglyConnectedComponents(g)
		compressed := CompressGraph(g)

		sorted, err := TopologicalSort(compressed)
		require.NoError(t, err)
		assert.Len(t, sorted, 4)

		positions := make(map[string]int)
		for i, node := range sorted {
			positions[node.Name] = i
		}

		// d must come before b and c
		assert.Less(t, positions["d"], positions["b"])
		assert.Less(t, positions["d"], positions["c"])

		// b and c must come before a
		assert.Less(t, positions["b"], positions["a"])
		assert.Less(t, positions["c"], positions["a"])
	})

	t.Run("compressed cycle with dependencies", func(t *testing.T) {
		// a <-> b (cycle), both depend on c, d depends on the cycle
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "c", Path: "./c", Ecosystem: config.EcosystemGo},
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

		sorted, err := TopologicalSort(compressed)
		require.NoError(t, err)

		// Should have 3 nodes: c, meta-node for {a,b}, and d
		assert.Len(t, sorted, 3)

		positions := make(map[string]int)
		for i, node := range sorted {
			positions[node.Name] = i
		}

		// Find meta-node name
		var metaNodeName string
		for _, node := range sorted {
			if len(node.Members) > 1 {
				metaNodeName = node.Name
				break
			}
		}
		require.NotEmpty(t, metaNodeName)

		// c must come before meta-node
		assert.Less(t, positions["c"], positions[metaNodeName])

		// meta-node must come before d
		assert.Less(t, positions[metaNodeName], positions["d"])
	})

	t.Run("multiple independent nodes", func(t *testing.T) {
		// a, b, c (no dependencies)
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "a", Path: "./a", Ecosystem: config.EcosystemGo},
				{Name: "b", Path: "./b", Ecosystem: config.EcosystemGo},
				{Name: "c", Path: "./c", Ecosystem: config.EcosystemGo},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)

		FindStronglyConnectedComponents(g)
		compressed := CompressGraph(g)

		sorted, err := TopologicalSort(compressed)
		require.NoError(t, err)
		assert.Len(t, sorted, 3)

		// Any order is valid since nodes are independent
		names := make(map[string]bool)
		for _, node := range sorted {
			names[node.Name] = true
		}
		assert.True(t, names["a"])
		assert.True(t, names["b"])
		assert.True(t, names["c"])
	})

	t.Run("complex monorepo structure", func(t *testing.T) {
		// utils, logging (independent base)
		// core depends on utils, logging
		// api depends on core
		// web, mobile depend on api
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "utils", Path: "./utils", Ecosystem: config.EcosystemGo},
				{Name: "logging", Path: "./logging", Ecosystem: config.EcosystemGo},
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "utils", Strategy: "linked"},
						{Package: "logging", Strategy: "linked"},
					},
				},
				{Name: "api", Path: "./api", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "core", Strategy: "linked"},
					},
				},
				{Name: "web", Path: "./web", Ecosystem: config.EcosystemNPM,
					Dependencies: []config.Dependency{
						{Package: "api", Strategy: "linked"},
					},
				},
				{Name: "mobile", Path: "./mobile", Ecosystem: config.EcosystemNPM,
					Dependencies: []config.Dependency{
						{Package: "api", Strategy: "linked"},
					},
				},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)

		FindStronglyConnectedComponents(g)
		compressed := CompressGraph(g)

		sorted, err := TopologicalSort(compressed)
		require.NoError(t, err)
		assert.Len(t, sorted, 6)

		positions := make(map[string]int)
		for i, node := range sorted {
			positions[node.Name] = i
		}

		// Base packages first
		assert.Less(t, positions["utils"], positions["core"])
		assert.Less(t, positions["logging"], positions["core"])

		// Core before api
		assert.Less(t, positions["core"], positions["api"])

		// Api before web and mobile
		assert.Less(t, positions["api"], positions["web"])
		assert.Less(t, positions["api"], positions["mobile"])
	})

	t.Run("empty graph", func(t *testing.T) {
		compressed := NewCompressedGraph()
		sorted, err := TopologicalSort(compressed)

		require.NoError(t, err)
		assert.Empty(t, sorted)
	})

	t.Run("single node", func(t *testing.T) {
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "solo", Path: "./solo", Ecosystem: config.EcosystemGo},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)

		FindStronglyConnectedComponents(g)
		compressed := CompressGraph(g)

		sorted, err := TopologicalSort(compressed)
		require.NoError(t, err)
		assert.Len(t, sorted, 1)
		assert.Equal(t, "solo", sorted[0].Name)
	})

	t.Run("multiple cycles compressed correctly", func(t *testing.T) {
		// Cycle 1: a <-> b
		// Cycle 2: c <-> d
		// e depends on cycle1
		// cycle2 depends on e
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
				{Name: "e", Path: "./e", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "a", Strategy: "linked"},
					},
				},
				{Name: "c", Path: "./c", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "d", Strategy: "linked"},
						{Package: "e", Strategy: "linked"},
					},
				},
				{Name: "d", Path: "./d", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "c", Strategy: "linked"},
					},
				},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)

		FindStronglyConnectedComponents(g)
		compressed := CompressGraph(g)

		sorted, err := TopologicalSort(compressed)
		require.NoError(t, err)

		// Should have 3 nodes: 2 meta-nodes and e
		assert.Len(t, sorted, 3)

		positions := make(map[string]int)
		nodesByMembers := make(map[string]*CompressedNode)
		for i, node := range sorted {
			positions[node.Name] = i
			if len(node.Members) > 1 {
				// Store by first member for easier lookup
				nodesByMembers[node.Members[0]] = node
			}
		}

		// Cycle1 before e, e before cycle2
		cycle1Name := nodesByMembers["a"].Name
		cycle2Name := nodesByMembers["c"].Name

		assert.Less(t, positions[cycle1Name], positions["e"])
		assert.Less(t, positions["e"], positions[cycle2Name])
	})
}
