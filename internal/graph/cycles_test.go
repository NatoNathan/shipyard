package graph

import (
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectCycles(t *testing.T) {
	t.Run("no cycles detected", func(t *testing.T) {
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

		hasCycles, cycles := DetectCycles(g)

		assert.False(t, hasCycles)
		assert.Empty(t, cycles)
	})

	t.Run("simple cycle detected - two nodes", func(t *testing.T) {
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

		hasCycles, cycles := DetectCycles(g)

		assert.True(t, hasCycles)
		assert.Len(t, cycles, 1)
		assert.Len(t, cycles[0], 2)

		// Verify both nodes are in the cycle
		cycleSet := make(map[string]bool)
		for _, node := range cycles[0] {
			cycleSet[node] = true
		}
		assert.True(t, cycleSet["a"])
		assert.True(t, cycleSet["b"])
	})

	t.Run("self-cycle detected", func(t *testing.T) {
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

		hasCycles, cycles := DetectCycles(g)

		assert.True(t, hasCycles)
		assert.Len(t, cycles, 1)
		assert.Len(t, cycles[0], 1)
		assert.Equal(t, "a", cycles[0][0])
	})

	t.Run("complex cycle detected - three nodes", func(t *testing.T) {
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

		hasCycles, cycles := DetectCycles(g)

		assert.True(t, hasCycles)
		assert.Len(t, cycles, 1)
		assert.Len(t, cycles[0], 3)

		// Verify all nodes are in the cycle
		cycleSet := make(map[string]bool)
		for _, node := range cycles[0] {
			cycleSet[node] = true
		}
		assert.True(t, cycleSet["a"])
		assert.True(t, cycleSet["b"])
		assert.True(t, cycleSet["c"])
	})

	t.Run("multiple independent cycles", func(t *testing.T) {
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

		hasCycles, cycles := DetectCycles(g)

		assert.True(t, hasCycles)
		assert.Len(t, cycles, 2)

		// Count cycle sizes
		cycleSizes := make(map[int]int)
		for _, cycle := range cycles {
			cycleSizes[len(cycle)]++
		}

		assert.Equal(t, 2, cycleSizes[2], "Should have 2 cycles of size 2")
	})

	t.Run("mixed graph - cycles and acyclic paths", func(t *testing.T) {
		// Graph structure:
		//   a -> b -> c (acyclic chain)
		//   d <-> e   (cycle)
		//   f -> d    (points to cycle but not part of it)
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

		hasCycles, cycles := DetectCycles(g)

		assert.True(t, hasCycles)
		assert.Len(t, cycles, 1, "Should detect exactly 1 cycle")
		assert.Len(t, cycles[0], 2, "Cycle should contain 2 nodes (d and e)")

		// Verify the cycle contains d and e
		cycleSet := make(map[string]bool)
		for _, node := range cycles[0] {
			cycleSet[node] = true
		}
		assert.True(t, cycleSet["d"])
		assert.True(t, cycleSet["e"])
	})

	t.Run("empty graph", func(t *testing.T) {
		g := NewGraph()
		hasCycles, cycles := DetectCycles(g)

		assert.False(t, hasCycles)
		assert.Empty(t, cycles)
	})

	t.Run("single node no edges", func(t *testing.T) {
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "solo", Path: "./solo", Ecosystem: config.EcosystemGo},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)

		hasCycles, cycles := DetectCycles(g)

		assert.False(t, hasCycles)
		assert.Empty(t, cycles)
	})

	t.Run("diamond structure - no cycle", func(t *testing.T) {
		// Diamond: a -> b -> d
		//          a -> c -> d
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
						{Package: "d", Strategy: "linked"},
					},
				},
				{Name: "c", Path: "./c", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "d", Strategy: "linked"},
					},
				},
				{Name: "d", Path: "./d", Ecosystem: config.EcosystemGo},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)

		hasCycles, cycles := DetectCycles(g)

		assert.False(t, hasCycles)
		assert.Empty(t, cycles)
	})
}
