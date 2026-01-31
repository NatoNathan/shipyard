package version

import (
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/graph"
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixedDependencyHandling(t *testing.T) {
	t.Run("fixed dependency blocks propagation", func(t *testing.T) {
		// api -> core (fixed)
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
				{Name: "api", Path: "./api", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "core", Strategy: "fixed"},
					},
				},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"core": {Major: 1, Minor: 0, Patch: 0},
			"api":  {Major: 2, Minor: 0, Patch: 0},
		}

		directBumps := map[string]string{
			"core": "major", // Breaking change in core
		}

		result, err := PropagateLinked(g, currentVersions, directBumps)
		require.NoError(t, err)

		// core gets bump, api does not (fixed strategy)
		assert.Contains(t, result, "core")
		assert.NotContains(t, result, "api")

		// core version changes
		assert.Equal(t, semver.Version{Major: 2, Minor: 0, Patch: 0}, result["core"].NewVersion)

		// api stays at 2.0.0 (no bump)
	})

	t.Run("fixed strategy isolates packages from dependency changes", func(t *testing.T) {
		// Multiple packages with fixed dependencies on core
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
				{Name: "api", Path: "./api", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "core", Strategy: "fixed"},
					},
				},
				{Name: "web", Path: "./web", Ecosystem: config.EcosystemNPM,
					Dependencies: []config.Dependency{
						{Package: "core", Strategy: "fixed"},
					},
				},
				{Name: "mobile", Path: "./mobile", Ecosystem: config.EcosystemNPM,
					Dependencies: []config.Dependency{
						{Package: "core", Strategy: "fixed"},
					},
				},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"core":   {Major: 1, Minor: 0, Patch: 0},
			"api":    {Major: 1, Minor: 0, Patch: 0},
			"web":    {Major: 1, Minor: 0, Patch: 0},
			"mobile": {Major: 1, Minor: 0, Patch: 0},
		}

		directBumps := map[string]string{
			"core": "minor",
		}

		result, err := PropagateLinked(g, currentVersions, directBumps)
		require.NoError(t, err)

		// Only core gets bump
		assert.Len(t, result, 1)
		assert.Contains(t, result, "core")
		assert.NotContains(t, result, "api")
		assert.NotContains(t, result, "web")
		assert.NotContains(t, result, "mobile")
	})

	t.Run("mixed strategies - fixed and linked", func(t *testing.T) {
		// api -> core (linked), web -> core (fixed)
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
				{Name: "api", Path: "./api", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "core", Strategy: "linked"},
					},
				},
				{Name: "web", Path: "./web", Ecosystem: config.EcosystemNPM,
					Dependencies: []config.Dependency{
						{Package: "core", Strategy: "fixed"},
					},
				},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"core": {Major: 1, Minor: 0, Patch: 0},
			"api":  {Major: 1, Minor: 0, Patch: 0},
			"web":  {Major: 1, Minor: 0, Patch: 0},
		}

		directBumps := map[string]string{
			"core": "patch",
		}

		result, err := PropagateLinked(g, currentVersions, directBumps)
		require.NoError(t, err)

		// core and api get bumps (linked), web does not (fixed)
		assert.Len(t, result, 2)
		assert.Contains(t, result, "core")
		assert.Contains(t, result, "api")
		assert.NotContains(t, result, "web")

		assert.Equal(t, "patch", result["core"].ChangeType)
		assert.Equal(t, "patch", result["api"].ChangeType)
	})

	t.Run("fixed blocks transitive propagation", func(t *testing.T) {
		// app -> service (fixed) -> core (linked)
		// Change to core should propagate to service but NOT to app
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
				{Name: "service", Path: "./service", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "core", Strategy: "linked"},
					},
				},
				{Name: "app", Path: "./app", Ecosystem: config.EcosystemNPM,
					Dependencies: []config.Dependency{
						{Package: "service", Strategy: "fixed"},
					},
				},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"core":    {Major: 1, Minor: 0, Patch: 0},
			"service": {Major: 1, Minor: 0, Patch: 0},
			"app":     {Major: 1, Minor: 0, Patch: 0},
		}

		directBumps := map[string]string{
			"core": "minor",
		}

		result, err := PropagateLinked(g, currentVersions, directBumps)
		require.NoError(t, err)

		// core and service get bumps, app does not
		assert.Contains(t, result, "core")
		assert.Contains(t, result, "service")
		assert.NotContains(t, result, "app")
	})

	t.Run("empty strategy defaults to linked", func(t *testing.T) {
		// api -> core (no strategy specified, should default to linked)
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
				{Name: "api", Path: "./api", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "core", Strategy: ""},
					},
				},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		// Check that empty strategy is stored as "linked" in graph
		edges := g.GetEdgesFrom("api")
		require.Len(t, edges, 1)
		assert.Equal(t, "linked", edges[0].Strategy)
	})

	t.Run("fixed strategy with major bump in dependency", func(t *testing.T) {
		// Test that even major breaking changes don't propagate through fixed
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "lib", Path: "./lib", Ecosystem: config.EcosystemGo},
				{Name: "consumer", Path: "./consumer", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "lib", Strategy: "fixed"},
					},
				},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"lib":      {Major: 1, Minor: 5, Patch: 10},
			"consumer": {Major: 2, Minor: 3, Patch: 7},
		}

		directBumps := map[string]string{
			"lib": "major", // Breaking change
		}

		result, err := PropagateLinked(g, currentVersions, directBumps)
		require.NoError(t, err)

		// Only lib changes, consumer is isolated
		assert.Len(t, result, 1)
		assert.Equal(t, semver.Version{Major: 2, Minor: 0, Patch: 0}, result["lib"].NewVersion)
	})
}
