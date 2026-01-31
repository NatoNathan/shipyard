package graph

import (
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildGraph(t *testing.T) {
	t.Run("empty config returns empty graph", func(t *testing.T) {
		cfg := &config.Config{
			Packages: []config.Package{},
		}

		g, err := BuildGraph(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, g)
		assert.Equal(t, 0, g.GetNodeCount())
		assert.Equal(t, 0, g.GetEdgeCount())
	})

	t.Run("single package with no dependencies", func(t *testing.T) {
		cfg := &config.Config{
			Packages: []config.Package{
				{
					Name:         "core",
					Path:         "./core",
					Ecosystem:    config.EcosystemGo,
					Dependencies: []config.Dependency{},
				},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)
		assert.Equal(t, 1, g.GetNodeCount())
		assert.Equal(t, 0, g.GetEdgeCount())

		node, exists := g.GetNode("core")
		assert.True(t, exists)
		assert.Equal(t, "core", node.Package.Name)
	})

	t.Run("multiple packages with dependencies", func(t *testing.T) {
		cfg := &config.Config{
			Packages: []config.Package{
				{
					Name:      "utils",
					Path:      "./utils",
					Ecosystem: config.EcosystemGo,
				},
				{
					Name:      "core",
					Path:      "./core",
					Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{
							Package:  "utils",
							Strategy: "linked",
						},
					},
				},
				{
					Name:      "api",
					Path:      "./api",
					Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{
							Package:  "core",
							Strategy: "linked",
						},
					},
				},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)
		assert.Equal(t, 3, g.GetNodeCount())
		assert.Equal(t, 2, g.GetEdgeCount())

		// Verify edges
		coreEdges := g.GetEdgesFrom("core")
		assert.Len(t, coreEdges, 1)
		assert.Equal(t, "utils", coreEdges[0].To)
		assert.Equal(t, "linked", coreEdges[0].Strategy)

		apiEdges := g.GetEdgesFrom("api")
		assert.Len(t, apiEdges, 1)
		assert.Equal(t, "core", apiEdges[0].To)
	})

	t.Run("package with custom bump mapping", func(t *testing.T) {
		cfg := &config.Config{
			Packages: []config.Package{
				{
					Name:      "core",
					Path:      "./core",
					Ecosystem: config.EcosystemGo,
				},
				{
					Name:      "web",
					Path:      "./web",
					Ecosystem: config.EcosystemNPM,
					Dependencies: []config.Dependency{
						{
							Package:  "core",
							Strategy: "linked",
							BumpMapping: map[string]string{
								"major": "patch",
								"minor": "patch",
								"patch": "patch",
							},
						},
					},
				},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)

		edges := g.GetEdgesFrom("web")
		assert.Len(t, edges, 1)
		assert.NotNil(t, edges[0].BumpMap)
		assert.Equal(t, "patch", edges[0].BumpMap["major"])
		assert.Equal(t, "patch", edges[0].BumpMap["minor"])
		assert.Equal(t, "patch", edges[0].BumpMap["patch"])
	})

	t.Run("error on missing dependency reference", func(t *testing.T) {
		cfg := &config.Config{
			Packages: []config.Package{
				{
					Name:      "api",
					Path:      "./api",
					Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{
							Package:  "nonexistent",
							Strategy: "linked",
						},
					},
				},
			},
		}

		g, err := BuildGraph(cfg)
		assert.Error(t, err)
		assert.Nil(t, g)
		assert.Contains(t, err.Error(), "nonexistent")
	})

	t.Run("multiple dependencies from one package", func(t *testing.T) {
		cfg := &config.Config{
			Packages: []config.Package{
				{
					Name:      "utils",
					Path:      "./utils",
					Ecosystem: config.EcosystemGo,
				},
				{
					Name:      "logging",
					Path:      "./logging",
					Ecosystem: config.EcosystemGo,
				},
				{
					Name:      "core",
					Path:      "./core",
					Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{
							Package:  "utils",
							Strategy: "linked",
						},
						{
							Package:  "logging",
							Strategy: "fixed",
						},
					},
				},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)
		assert.Equal(t, 3, g.GetNodeCount())
		assert.Equal(t, 2, g.GetEdgeCount())

		edges := g.GetEdgesFrom("core")
		assert.Len(t, edges, 2)

		// Verify both edges exist
		targetNames := make(map[string]bool)
		for _, edge := range edges {
			targetNames[edge.To] = true
		}
		assert.True(t, targetNames["utils"])
		assert.True(t, targetNames["logging"])
	})

	t.Run("complex monorepo structure", func(t *testing.T) {
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "utils", Path: "./packages/utils", Ecosystem: config.EcosystemGo},
				{Name: "logging", Path: "./packages/logging", Ecosystem: config.EcosystemGo},
				{
					Name:      "core",
					Path:      "./packages/core",
					Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "utils", Strategy: "linked"},
						{Package: "logging", Strategy: "linked"},
					},
				},
				{
					Name:      "api",
					Path:      "./services/api",
					Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "core", Strategy: "linked"},
					},
				},
				{
					Name:      "web",
					Path:      "./apps/web",
					Ecosystem: config.EcosystemNPM,
					Dependencies: []config.Dependency{
						{
							Package:  "api",
							Strategy: "linked",
							BumpMapping: map[string]string{
								"major": "minor",
								"minor": "patch",
								"patch": "patch",
							},
						},
					},
				},
				{
					Name:      "mobile",
					Path:      "./apps/mobile",
					Ecosystem: config.EcosystemNPM,
					Dependencies: []config.Dependency{
						{Package: "api", Strategy: "fixed"},
					},
				},
			},
		}

		g, err := BuildGraph(cfg)
		require.NoError(t, err)
		assert.Equal(t, 6, g.GetNodeCount())
		// core->utils, core->logging, api->core, web->api, mobile->api = 5 edges
		assert.Equal(t, 5, g.GetEdgeCount())

		// Verify all nodes exist
		for _, pkgName := range []string{"utils", "logging", "core", "api", "web", "mobile"} {
			_, exists := g.GetNode(pkgName)
			assert.True(t, exists, "node %s should exist", pkgName)
		}

		// Verify core has 2 dependencies
		assert.Len(t, g.GetEdgesFrom("core"), 2)

		// Verify web has custom bump mapping
		webEdges := g.GetEdgesFrom("web")
		assert.Len(t, webEdges, 1)
		assert.NotNil(t, webEdges[0].BumpMap)
		assert.Equal(t, "minor", webEdges[0].BumpMap["major"])

		// Verify mobile has fixed strategy
		mobileEdges := g.GetEdgesFrom("mobile")
		assert.Len(t, mobileEdges, 1)
		assert.Equal(t, "fixed", mobileEdges[0].Strategy)
	})
}
