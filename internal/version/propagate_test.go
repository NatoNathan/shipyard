package version

import (
	"testing"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/graph"
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPropagateLinked(t *testing.T) {
	t.Run("simple linked propagation - two packages", func(t *testing.T) {
		// api -> core (linked)
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
				{Name: "api", Path: "./api", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "core", Strategy: "linked"},
					},
				},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"core": {Major: 1, Minor: 0, Patch: 0},
			"api":  {Major: 1, Minor: 0, Patch: 0},
		}

		directBumps := map[string]string{
			"core": "minor",
		}

		result, err := PropagateLinked(g, currentVersions, directBumps)
		require.NoError(t, err)

		// Both should receive bumps
		assert.Contains(t, result, "core")
		assert.Contains(t, result, "api")

		// core: direct minor bump
		assert.Equal(t, semver.Version{Major: 1, Minor: 1, Patch: 0}, result["core"].NewVersion)
		assert.Equal(t, "minor", result["core"].ChangeType)
		assert.Equal(t, "direct", result["core"].Source)

		// api: propagated minor bump
		assert.Equal(t, semver.Version{Major: 1, Minor: 1, Patch: 0}, result["api"].NewVersion)
		assert.Equal(t, "minor", result["api"].ChangeType)
		assert.Equal(t, "propagated", result["api"].Source)
	})

	t.Run("chain propagation - three packages", func(t *testing.T) {
		// web -> api -> core (all linked)
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
						{Package: "api", Strategy: "linked"},
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

		// All three should receive bumps
		assert.Len(t, result, 3)
		assert.Equal(t, "patch", result["core"].ChangeType)
		assert.Equal(t, "patch", result["api"].ChangeType)
		assert.Equal(t, "patch", result["web"].ChangeType)
	})

	t.Run("bump mapping - major to patch", func(t *testing.T) {
		// web -> api (linked with bump mapping)
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "api", Path: "./api", Ecosystem: config.EcosystemGo},
				{Name: "web", Path: "./web", Ecosystem: config.EcosystemNPM,
					Dependencies: []config.Dependency{
						{
							Package:  "api",
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

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"api": {Major: 1, Minor: 0, Patch: 0},
			"web": {Major: 2, Minor: 0, Patch: 0},
		}

		directBumps := map[string]string{
			"api": "major", // Major change in API
		}

		result, err := PropagateLinked(g, currentVersions, directBumps)
		require.NoError(t, err)

		// API gets major bump
		assert.Equal(t, semver.Version{Major: 2, Minor: 0, Patch: 0}, result["api"].NewVersion)
		assert.Equal(t, "major", result["api"].ChangeType)

		// Web gets patch bump (mapped from major)
		assert.Equal(t, semver.Version{Major: 2, Minor: 0, Patch: 1}, result["web"].NewVersion)
		assert.Equal(t, "patch", result["web"].ChangeType)
	})

	t.Run("diamond dependency - multiple paths", func(t *testing.T) {
		// d depends on both b and c, both depend on a
		// a (bump) -> b -> d
		//          -> c -> d
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "a", Path: "./a", Ecosystem: config.EcosystemGo},
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
			"a": "minor",
		}

		result, err := PropagateLinked(g, currentVersions, directBumps)
		require.NoError(t, err)

		// All packages should receive minor bump
		assert.Len(t, result, 4)
		for _, pkg := range []string{"a", "b", "c", "d"} {
			assert.Equal(t, "minor", result[pkg].ChangeType, "Package %s should have minor bump", pkg)
			assert.Equal(t, semver.Version{Major: 1, Minor: 1, Patch: 0}, result[pkg].NewVersion)
		}
	})

	t.Run("direct bump overrides propagation", func(t *testing.T) {
		// Both packages have direct bumps
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
				{Name: "api", Path: "./api", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "core", Strategy: "linked"},
					},
				},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"core": {Major: 1, Minor: 0, Patch: 0},
			"api":  {Major: 1, Minor: 0, Patch: 0},
		}

		directBumps := map[string]string{
			"core": "patch",
			"api":  "major", // Direct major bump
		}

		result, err := PropagateLinked(g, currentVersions, directBumps)
		require.NoError(t, err)

		// core: patch
		assert.Equal(t, semver.Version{Major: 1, Minor: 0, Patch: 1}, result["core"].NewVersion)
		assert.Equal(t, "direct", result["core"].Source)

		// api: major (not propagated patch from core)
		assert.Equal(t, semver.Version{Major: 2, Minor: 0, Patch: 0}, result["api"].NewVersion)
		assert.Equal(t, "major", result["api"].ChangeType)
		assert.Equal(t, "direct", result["api"].Source)
	})

	t.Run("propagation respects only linked strategy", func(t *testing.T) {
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
			"core": "minor",
		}

		result, err := PropagateLinked(g, currentVersions, directBumps)
		require.NoError(t, err)

		// core and api get bumps, web does not (fixed strategy)
		assert.Contains(t, result, "core")
		assert.Contains(t, result, "api")
		assert.NotContains(t, result, "web")
	})

	t.Run("no propagation for empty direct bumps", func(t *testing.T) {
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
				{Name: "api", Path: "./api", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "core", Strategy: "linked"},
					},
				},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"core": {Major: 1, Minor: 0, Patch: 0},
			"api":  {Major: 1, Minor: 0, Patch: 0},
		}

		directBumps := map[string]string{} // Empty

		result, err := PropagateLinked(g, currentVersions, directBumps)
		require.NoError(t, err)

		assert.Empty(t, result)
	})

	t.Run("error for missing current version", func(t *testing.T) {
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
				{Name: "api", Path: "./api", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "core", Strategy: "linked"},
					},
				},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"core": {Major: 1, Minor: 0, Patch: 0},
			// Missing "api" version
		}

		directBumps := map[string]string{
			"core": "minor",
		}

		result, err := PropagateLinked(g, currentVersions, directBumps)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "api")
	})
}
