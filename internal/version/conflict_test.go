package version

import (
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/graph"
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationWithPropagation(t *testing.T) {
	t.Run("full propagation pipeline with conflicts - higher priority wins", func(t *testing.T) {
		// Complex scenario:
		// a (patch bump) -> b (linked) -> d (linked)
		// c (minor bump) -> d (linked)
		// d receives both patch (from b/a) and minor (from c)
		// Should get minor (higher priority)
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "a", Path: "./a", Ecosystem: config.EcosystemGo},
				{Name: "b", Path: "./b", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "a", Strategy: "linked"},
					},
				},
				{Name: "c", Path: "./c", Ecosystem: config.EcosystemGo},
				{Name: "d", Path: "./d", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "b", Strategy: "linked"},
						{Package: "c", Strategy: "linked"},
					},
				},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"a": {Major: 1, Minor: 0, Patch: 0},
			"b": {Major: 1, Minor: 0, Patch: 0},
			"c": {Major: 1, Minor: 0, Patch: 0},
			"d": {Major: 1, Minor: 0, Patch: 0},
		}

		directBumps := map[string]string{
			"a": "patch",
			"c": "minor",
		}

		// Run propagation
		result, err := PropagateLinked(g, currentVersions, directBumps)
		require.NoError(t, err)

		resolved := result

		// a: patch (direct)
		assert.Equal(t, "patch", resolved["a"].ChangeType)
		assert.Equal(t, semver.Version{Major: 1, Minor: 0, Patch: 1}, resolved["a"].NewVersion)

		// b: patch (from a)
		assert.Equal(t, "patch", resolved["b"].ChangeType)

		// c: minor (direct)
		assert.Equal(t, "minor", resolved["c"].ChangeType)
		assert.Equal(t, semver.Version{Major: 1, Minor: 1, Patch: 0}, resolved["c"].NewVersion)

		// d: minor (higher priority wins from c path)
		assert.Equal(t, "minor", resolved["d"].ChangeType)
		assert.Equal(t, semver.Version{Major: 1, Minor: 1, Patch: 0}, resolved["d"].NewVersion)
	})

	t.Run("cycle resolution then conflict resolution", func(t *testing.T) {
		// a <-> b (cycle, both have bumps)
		// c depends on a
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
						{Package: "a", Strategy: "linked"},
					},
				},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		graph.FindStronglyConnectedComponents(g)

		currentVersions := map[string]semver.Version{
			"a": {Major: 1, Minor: 0, Patch: 0},
			"b": {Major: 1, Minor: 0, Patch: 0},
			"c": {Major: 1, Minor: 0, Patch: 0},
		}

		directBumps := map[string]string{
			"a": "patch",
			"b": "minor",
		}

		// Resolve cycles first
		cycleResolved, err := ResolveCycleBumps(g, currentVersions, directBumps)
		require.NoError(t, err)

		// Both a and b should have minor (highest in cycle)
		assert.Equal(t, "minor", cycleResolved["a"].ChangeType)
		assert.Equal(t, "minor", cycleResolved["b"].ChangeType)

		// Then propagate
		// Need to extract just the change types for propagation
		resolvedBumps := make(map[string]string)
		for pkg, bump := range cycleResolved {
			resolvedBumps[pkg] = bump.ChangeType
		}

		propagated, err := PropagateLinked(g, currentVersions, resolvedBumps)
		require.NoError(t, err)

		final := propagated

		// c should get minor (propagated from the cycle)
		assert.Contains(t, final, "c")
		assert.Equal(t, "minor", final["c"].ChangeType)
	})
}
