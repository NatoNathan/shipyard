package integration

import (
	"testing"
	"time"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/consignment"
	"github.com/NatoNathan/shipyard/internal/graph"
	"github.com/NatoNathan/shipyard/internal/version"
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/NatoNathan/shipyard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndToEndPropagation tests the complete version propagation pipeline
// from consignments through graph analysis to final version bumps.
func TestEndToEndPropagation(t *testing.T) {
	t.Run("simple monorepo - linked propagation", func(t *testing.T) {
		// Scenario: Simple monorepo with linked dependencies
		// utils (patch) -> core (linked) -> api (linked)
		// All should get patch bump
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "utils", Path: "./packages/utils", Ecosystem: config.EcosystemGo},
				{Name: "core", Path: "./packages/core", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "utils", Strategy: "linked"},
					},
				},
				{Name: "api", Path: "./services/api", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "core", Strategy: "linked"},
					},
				},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		prop, err := version.NewPropagator(g)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"utils": {Major: 1, Minor: 2, Patch: 0},
			"core":  {Major: 2, Minor: 0, Patch: 0},
			"api":   {Major: 3, Minor: 1, Patch: 5},
		}

		consignments := []*consignment.Consignment{
			{
				ID:         "20250130-120000-abc123",
				Timestamp:  time.Now(),
				Packages:   []string{"utils"},
				ChangeType: types.ChangeTypePatch,
				Summary:    "Fix bug in utils",
			},
		}

		result, err := prop.Propagate(currentVersions, consignments)
		require.NoError(t, err)

		// All packages should be bumped
		assert.Len(t, result, 3)

		// Verify versions
		assert.Equal(t, semver.Version{Major: 1, Minor: 2, Patch: 1}, result["utils"].NewVersion)
		assert.Equal(t, semver.Version{Major: 2, Minor: 0, Patch: 1}, result["core"].NewVersion)
		assert.Equal(t, semver.Version{Major: 3, Minor: 1, Patch: 6}, result["api"].NewVersion)
	})

	t.Run("monorepo with fixed dependencies", func(t *testing.T) {
		// Scenario: API depends on core (linked), Web depends on API (fixed)
		// core gets minor bump, api propagates it, web stays unchanged
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./packages/core", Ecosystem: config.EcosystemGo},
				{Name: "api", Path: "./services/api", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "core", Strategy: "linked"},
					},
				},
				{Name: "web", Path: "./apps/web", Ecosystem: config.EcosystemNPM,
					Dependencies: []config.Dependency{
						{Package: "api", Strategy: "fixed"},
					},
				},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		prop, err := version.NewPropagator(g)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"core": {Major: 1, Minor: 0, Patch: 0},
			"api":  {Major: 1, Minor: 0, Patch: 0},
			"web":  {Major: 2, Minor: 3, Patch: 0},
		}

		consignments := []*consignment.Consignment{
			{
				ID:         "20250130-120100-def456",
				Timestamp:  time.Now(),
				Packages:   []string{"core"},
				ChangeType: types.ChangeTypeMinor,
				Summary:    "Add new feature to core",
			},
		}

		result, err := prop.Propagate(currentVersions, consignments)
		require.NoError(t, err)

		// Only core and api should be bumped (web has fixed strategy)
		assert.Len(t, result, 2)
		assert.Contains(t, result, "core")
		assert.Contains(t, result, "api")
		assert.NotContains(t, result, "web")

		assert.Equal(t, semver.Version{Major: 1, Minor: 1, Patch: 0}, result["core"].NewVersion)
		assert.Equal(t, semver.Version{Major: 1, Minor: 1, Patch: 0}, result["api"].NewVersion)
	})

	t.Run("bump mapping - major to patch", func(t *testing.T) {
		// Scenario: Go package gets major bump, NPM consumer maps to patch
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "go-sdk", Path: "./sdk", Ecosystem: config.EcosystemGo},
				{Name: "web-app", Path: "./web", Ecosystem: config.EcosystemNPM,
					Dependencies: []config.Dependency{
						{
							Package:  "go-sdk",
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

		prop, err := version.NewPropagator(g)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"go-sdk":  {Major: 1, Minor: 0, Patch: 0},
			"web-app": {Major: 5, Minor: 2, Patch: 1},
		}

		consignments := []*consignment.Consignment{
			{
				ID:         "20250130-120200-ghi789",
				Timestamp:  time.Now(),
				Packages:   []string{"go-sdk"},
				ChangeType: types.ChangeTypeMajor,
				Summary:    "Breaking change in SDK",
			},
		}

		result, err := prop.Propagate(currentVersions, consignments)
		require.NoError(t, err)

		// SDK gets major, web-app gets patch (mapped)
		assert.Equal(t, semver.Version{Major: 2, Minor: 0, Patch: 0}, result["go-sdk"].NewVersion)
		assert.Equal(t, semver.Version{Major: 5, Minor: 2, Patch: 2}, result["web-app"].NewVersion)
		assert.Equal(t, "major", result["go-sdk"].ChangeType)
		assert.Equal(t, "patch", result["web-app"].ChangeType)
	})

	t.Run("multiple consignments - priority resolution", func(t *testing.T) {
		// Scenario: Package has patch and minor consignments, should get minor
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		prop, err := version.NewPropagator(g)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"core": {Major: 1, Minor: 5, Patch: 3},
		}

		consignments := []*consignment.Consignment{
			{
				ID:         "20250130-120300-jkl012",
				Timestamp:  time.Now(),
				Packages:   []string{"core"},
				ChangeType: types.ChangeTypePatch,
				Summary:    "Fix typo",
			},
			{
				ID:         "20250130-120400-mno345",
				Timestamp:  time.Now().Add(time.Minute),
				Packages:   []string{"core"},
				ChangeType: types.ChangeTypeMinor,
				Summary:    "Add feature",
			},
		}

		result, err := prop.Propagate(currentVersions, consignments)
		require.NoError(t, err)

		// Should get minor (higher priority)
		assert.Equal(t, semver.Version{Major: 1, Minor: 6, Patch: 0}, result["core"].NewVersion)
		assert.Equal(t, "minor", result["core"].ChangeType)
	})

	t.Run("diamond dependency - all paths converge", func(t *testing.T) {
		// Scenario:
		//       d
		//      / \
		//     b   c
		//      \ /
		//       a (patch)
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

		prop, err := version.NewPropagator(g)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"a": {Major: 1, Minor: 0, Patch: 0},
			"b": {Major: 1, Minor: 0, Patch: 0},
			"c": {Major: 1, Minor: 0, Patch: 0},
			"d": {Major: 1, Minor: 0, Patch: 0},
		}

		consignments := []*consignment.Consignment{
			{
				ID:         "20250130-120500-pqr678",
				Timestamp:  time.Now(),
				Packages:   []string{"a"},
				ChangeType: types.ChangeTypePatch,
				Summary:    "Fix in base package",
			},
		}

		result, err := prop.Propagate(currentVersions, consignments)
		require.NoError(t, err)

		// All packages should get patch
		for _, pkg := range []string{"a", "b", "c", "d"} {
			assert.Contains(t, result, pkg)
			assert.Equal(t, semver.Version{Major: 1, Minor: 0, Patch: 1}, result[pkg].NewVersion)
		}
	})

	t.Run("complex monorepo - real-world scenario", func(t *testing.T) {
		// Scenario: Typical monorepo structure
		// - Shared libraries: utils, logging
		// - Core package depends on both
		// - API service depends on core
		// - Web and Mobile apps depend on API (different strategies)
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "utils", Path: "./packages/utils", Ecosystem: config.EcosystemGo},
				{Name: "logging", Path: "./packages/logging", Ecosystem: config.EcosystemGo},
				{Name: "core", Path: "./packages/core", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "utils", Strategy: "linked"},
						{Package: "logging", Strategy: "linked"},
					},
				},
				{Name: "api", Path: "./services/api", Ecosystem: config.EcosystemGo,
					Dependencies: []config.Dependency{
						{Package: "core", Strategy: "linked"},
					},
				},
				{Name: "web", Path: "./apps/web", Ecosystem: config.EcosystemNPM,
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
				{Name: "mobile", Path: "./apps/mobile", Ecosystem: config.EcosystemNPM,
					Dependencies: []config.Dependency{
						{Package: "api", Strategy: "fixed"},
					},
				},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		prop, err := version.NewPropagator(g)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"utils":   {Major: 1, Minor: 0, Patch: 0},
			"logging": {Major: 1, Minor: 0, Patch: 0},
			"core":    {Major: 2, Minor: 0, Patch: 0},
			"api":     {Major: 3, Minor: 0, Patch: 0},
			"web":     {Major: 4, Minor: 0, Patch: 0},
			"mobile":  {Major: 5, Minor: 0, Patch: 0},
		}

		consignments := []*consignment.Consignment{
			{
				ID:         "20250130-120600-stu901",
				Timestamp:  time.Now(),
				Packages:   []string{"utils"},
				ChangeType: types.ChangeTypeMinor,
				Summary:    "Add helper function",
			},
		}

		result, err := prop.Propagate(currentVersions, consignments)
		require.NoError(t, err)

		// Verify propagation path
		assert.Contains(t, result, "utils")
		assert.Contains(t, result, "core")
		assert.Contains(t, result, "api")
		assert.Contains(t, result, "web")
		assert.NotContains(t, result, "mobile") // Fixed strategy
		assert.NotContains(t, result, "logging") // No consignment

		// Check version bumps
		assert.Equal(t, semver.Version{Major: 1, Minor: 1, Patch: 0}, result["utils"].NewVersion)
		assert.Equal(t, semver.Version{Major: 2, Minor: 1, Patch: 0}, result["core"].NewVersion)
		assert.Equal(t, semver.Version{Major: 3, Minor: 1, Patch: 0}, result["api"].NewVersion)

		// Web gets patch (minor mapped to patch)
		assert.Equal(t, semver.Version{Major: 4, Minor: 0, Patch: 1}, result["web"].NewVersion)
		assert.Equal(t, "patch", result["web"].ChangeType)
	})

	t.Run("validation - missing package in consignment", func(t *testing.T) {
		// Scenario: Consignment references package not in config
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		prop, err := version.NewPropagator(g)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"core": {Major: 1, Minor: 0, Patch: 0},
		}

		consignments := []*consignment.Consignment{
			{
				ID:         "20250130-120700-vwx234",
				Timestamp:  time.Now(),
				Packages:   []string{"nonexistent"},
				ChangeType: types.ChangeTypePatch,
				Summary:    "Invalid consignment",
			},
		}

		result, err := prop.Propagate(currentVersions, consignments)

		// Should fail with error about missing version
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "nonexistent")
	})

	t.Run("empty consignments - no bumps", func(t *testing.T) {
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		prop, err := version.NewPropagator(g)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"core": {Major: 1, Minor: 0, Patch: 0},
		}

		result, err := prop.Propagate(currentVersions, []*consignment.Consignment{})
		require.NoError(t, err)

		// No bumps
		assert.Empty(t, result)
	})
}
