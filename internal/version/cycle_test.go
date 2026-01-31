package version

import (
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/graph"
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveCycleBumps(t *testing.T) {
	t.Run("simple cycle - both get highest bump", func(t *testing.T) {
		// a <-> b (cycle)
		// a has patch bump, b has minor bump
		// Both should get minor (highest)
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

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		// Run SCC detection to identify the cycle
		graph.FindStronglyConnectedComponents(g)

		currentVersions := map[string]semver.Version{
			"a": {Major: 1, Minor: 0, Patch: 0},
			"b": {Major: 1, Minor: 0, Patch: 0},
		}

		directBumps := map[string]string{
			"a": "patch",
			"b": "minor",
		}

		result, err := ResolveCycleBumps(g, currentVersions, directBumps)
		require.NoError(t, err)

		// Both should get minor bump (highest priority)
		assert.Contains(t, result, "a")
		assert.Contains(t, result, "b")

		assert.Equal(t, "minor", result["a"].ChangeType)
		assert.Equal(t, "minor", result["b"].ChangeType)

		assert.Equal(t, semver.Version{Major: 1, Minor: 1, Patch: 0}, result["a"].NewVersion)
		assert.Equal(t, semver.Version{Major: 1, Minor: 1, Patch: 0}, result["b"].NewVersion)

		assert.Equal(t, "cycle", result["a"].Source)
		assert.Equal(t, "cycle", result["b"].Source)
	})

	t.Run("three-node cycle - all get highest bump", func(t *testing.T) {
		// a -> b -> c -> a (cycle)
		// Different bumps: a=patch, b=minor, c=major
		// All should get major
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
			"c": "major",
		}

		result, err := ResolveCycleBumps(g, currentVersions, directBumps)
		require.NoError(t, err)

		// All should get major bump
		for _, pkg := range []string{"a", "b", "c"} {
			assert.Equal(t, "major", result[pkg].ChangeType)
			assert.Equal(t, semver.Version{Major: 2, Minor: 0, Patch: 0}, result[pkg].NewVersion)
			assert.Equal(t, "cycle", result[pkg].Source)
		}
	})

	t.Run("multiple separate cycles", func(t *testing.T) {
		// Cycle 1: a <-> b (patch, minor) -> both get minor
		// Cycle 2: c <-> d (patch, major) -> both get major
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
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		graph.FindStronglyConnectedComponents(g)

		currentVersions := map[string]semver.Version{
			"a": {Major: 1, Minor: 0, Patch: 0},
			"b": {Major: 1, Minor: 0, Patch: 0},
			"c": {Major: 2, Minor: 0, Patch: 0},
			"d": {Major: 2, Minor: 0, Patch: 0},
		}

		directBumps := map[string]string{
			"a": "patch",
			"b": "minor",
			"c": "patch",
			"d": "major",
		}

		result, err := ResolveCycleBumps(g, currentVersions, directBumps)
		require.NoError(t, err)

		// Cycle 1: both get minor
		assert.Equal(t, "minor", result["a"].ChangeType)
		assert.Equal(t, "minor", result["b"].ChangeType)

		// Cycle 2: both get major
		assert.Equal(t, "major", result["c"].ChangeType)
		assert.Equal(t, "major", result["d"].ChangeType)
	})

	t.Run("cycle with only one package having bump", func(t *testing.T) {
		// a <-> b (cycle)
		// Only a has a direct bump, b should get the same
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

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		graph.FindStronglyConnectedComponents(g)

		currentVersions := map[string]semver.Version{
			"a": {Major: 1, Minor: 0, Patch: 0},
			"b": {Major: 1, Minor: 0, Patch: 0},
		}

		directBumps := map[string]string{
			"a": "minor",
			// b has no direct bump, but should get minor from cycle resolution
		}

		result, err := ResolveCycleBumps(g, currentVersions, directBumps)
		require.NoError(t, err)

		// Both should get minor
		assert.Equal(t, "minor", result["a"].ChangeType)
		assert.Equal(t, "minor", result["b"].ChangeType)
	})

	t.Run("no cycles - returns direct bumps unchanged", func(t *testing.T) {
		// Linear: a -> b (no cycle)
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "a", Path: "./a", Ecosystem: config.EcosystemGo},
				{Name: "b", Path: "./b", Ecosystem: config.EcosystemGo,
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
		}

		directBumps := map[string]string{
			"a": "patch",
			"b": "minor",
		}

		result, err := ResolveCycleBumps(g, currentVersions, directBumps)
		require.NoError(t, err)

		// No cycles, so direct bumps preserved with "direct" source
		assert.Equal(t, "patch", result["a"].ChangeType)
		assert.Equal(t, "minor", result["b"].ChangeType)
		assert.Equal(t, "direct", result["a"].Source)
		assert.Equal(t, "direct", result["b"].Source)
	})

	t.Run("self-cycle", func(t *testing.T) {
		// a -> a (self-cycle)
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "a", Path: "./a", Ecosystem: config.EcosystemGo,
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
		}

		directBumps := map[string]string{
			"a": "minor",
		}

		result, err := ResolveCycleBumps(g, currentVersions, directBumps)
		require.NoError(t, err)

		// Single node cycle still gets "cycle" source
		assert.Equal(t, "minor", result["a"].ChangeType)
		assert.Equal(t, "cycle", result["a"].Source)
	})

	t.Run("empty direct bumps", func(t *testing.T) {
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "a", Path: "./a", Ecosystem: config.EcosystemGo},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		graph.FindStronglyConnectedComponents(g)

		currentVersions := map[string]semver.Version{
			"a": {Major: 1, Minor: 0, Patch: 0},
		}

		directBumps := map[string]string{} // Empty

		result, err := ResolveCycleBumps(g, currentVersions, directBumps)
		require.NoError(t, err)

		assert.Empty(t, result)
	})

	t.Run("cycle mixed with non-cycle packages", func(t *testing.T) {
		// a <-> b (cycle with patch, minor) -> both get minor
		// c (independent with major) -> gets major
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
				{Name: "c", Path: "./c", Ecosystem: config.EcosystemGo},
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
			"c": "major",
		}

		result, err := ResolveCycleBumps(g, currentVersions, directBumps)
		require.NoError(t, err)

		// Cycle members get minor
		assert.Equal(t, "minor", result["a"].ChangeType)
		assert.Equal(t, "minor", result["b"].ChangeType)
		assert.Equal(t, "cycle", result["a"].Source)
		assert.Equal(t, "cycle", result["b"].Source)

		// Non-cycle member keeps its direct bump
		assert.Equal(t, "major", result["c"].ChangeType)
		assert.Equal(t, "direct", result["c"].Source)
	})
}
