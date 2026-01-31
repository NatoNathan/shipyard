package graph

import (
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraphCache(t *testing.T) {
	t.Run("cache miss - builds and stores graph", func(t *testing.T) {
		cache := NewGraphCache()

		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
			},
		}

		g, err := cache.GetOrBuild(cfg)
		require.NoError(t, err)
		assert.NotNil(t, g)

		// Verify graph was built correctly
		assert.Equal(t, 1, g.GetNodeCount())
		node, exists := g.GetNode("core")
		assert.True(t, exists)
		assert.Equal(t, "core", node.Package.Name)
	})

	t.Run("cache hit - returns cached graph", func(t *testing.T) {
		cache := NewGraphCache()

		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
			},
		}

		// First call - builds graph
		g1, err := cache.GetOrBuild(cfg)
		require.NoError(t, err)

		// Second call - should return cached
		g2, err := cache.GetOrBuild(cfg)
		require.NoError(t, err)

		// Should be the same graph instance
		assert.True(t, g1 == g2, "Should return cached graph instance")
	})

	t.Run("different configs - separate cache entries", func(t *testing.T) {
		cache := NewGraphCache()

		cfg1 := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
			},
		}

		cfg2 := &config.Config{
			Packages: []config.Package{
				{Name: "api", Path: "./api", Ecosystem: config.EcosystemGo},
			},
		}

		g1, err := cache.GetOrBuild(cfg1)
		require.NoError(t, err)

		g2, err := cache.GetOrBuild(cfg2)
		require.NoError(t, err)

		// Should be different graphs
		assert.False(t, g1 == g2)

		// Verify each graph has correct content
		_, exists1 := g1.GetNode("core")
		assert.True(t, exists1)

		_, exists2 := g2.GetNode("api")
		assert.True(t, exists2)
	})

	t.Run("clear cache", func(t *testing.T) {
		cache := NewGraphCache()

		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
			},
		}

		// Build and cache
		g1, err := cache.GetOrBuild(cfg)
		require.NoError(t, err)

		// Clear cache
		cache.Clear()

		// Next call should rebuild
		g2, err := cache.GetOrBuild(cfg)
		require.NoError(t, err)

		// Should be different instances (rebuilt)
		assert.False(t, g1 == g2)
	})

	t.Run("invalidate specific config", func(t *testing.T) {
		cache := NewGraphCache()

		cfg1 := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
			},
		}

		cfg2 := &config.Config{
			Packages: []config.Package{
				{Name: "api", Path: "./api", Ecosystem: config.EcosystemGo},
			},
		}

		// Cache both
		g1a, _ := cache.GetOrBuild(cfg1)
		g2a, _ := cache.GetOrBuild(cfg2)

		// Invalidate cfg1
		cache.Invalidate(cfg1)

		// cfg1 should rebuild, cfg2 should be cached
		g1b, _ := cache.GetOrBuild(cfg1)
		g2b, _ := cache.GetOrBuild(cfg2)

		assert.False(t, g1a == g1b, "cfg1 should be rebuilt")
		assert.True(t, g2a == g2b, "cfg2 should be cached")
	})

	t.Run("cache preserves SCC information", func(t *testing.T) {
		cache := NewGraphCache()

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

		g1, err := cache.GetOrBuild(cfg)
		require.NoError(t, err)

		// Run SCC detection on first graph
		FindStronglyConnectedComponents(g1)

		nodeA1, _ := g1.GetNode("a")
		nodeB1, _ := g1.GetNode("b")
		sccID1 := nodeA1.SCC

		// Get cached graph
		g2, err := cache.GetOrBuild(cfg)
		require.NoError(t, err)

		// Cached graph should have same SCC information
		nodeA2, _ := g2.GetNode("a")
		nodeB2, _ := g2.GetNode("b")

		assert.Equal(t, sccID1, nodeA2.SCC)
		assert.Equal(t, nodeB1.SCC, nodeB2.SCC)
	})

	t.Run("cache handles build errors", func(t *testing.T) {
		cache := NewGraphCache()

		// Config with invalid dependency
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "api", Path: "./api", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "nonexistent", Strategy: "linked"},
					},
				},
			},
		}

		g, err := cache.GetOrBuild(cfg)
		assert.Error(t, err)
		assert.Nil(t, g)

		// Error shouldn't be cached - next call should also try to build
		g2, err2 := cache.GetOrBuild(cfg)
		assert.Error(t, err2)
		assert.Nil(t, g2)
	})

	t.Run("nil config returns error", func(t *testing.T) {
		cache := NewGraphCache()

		g, err := cache.GetOrBuild(nil)
		assert.Error(t, err)
		assert.Nil(t, g)
	})

	t.Run("empty config caches empty graph", func(t *testing.T) {
		cache := NewGraphCache()

		cfg := &config.Config{
			Packages: []config.Package{},
		}

		g, err := cache.GetOrBuild(cfg)
		require.NoError(t, err)
		assert.NotNil(t, g)
		assert.Equal(t, 0, g.GetNodeCount())

		// Should be cached
		g2, err := cache.GetOrBuild(cfg)
		require.NoError(t, err)
		assert.True(t, g == g2)
	})
}
