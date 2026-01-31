package version

import (
	"testing"
	"time"

	"github.com/NatoNathan/shipyard/internal/config"
	"github.com/NatoNathan/shipyard/internal/consignment"
	"github.com/NatoNathan/shipyard/internal/graph"
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/NatoNathan/shipyard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPropagator(t *testing.T) {
	t.Run("creates propagator with valid graph", func(t *testing.T) {
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "a", Path: "./a", Ecosystem: config.EcosystemGo},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		prop, err := NewPropagator(g)
		assert.NoError(t, err)
		assert.NotNil(t, prop)
	})

	t.Run("returns error for nil graph", func(t *testing.T) {
		prop, err := NewPropagator(nil)
		assert.Error(t, err)
		assert.Nil(t, prop)
	})
}

func TestPropagate(t *testing.T) {
	t.Run("single package with direct change", func(t *testing.T) {
		// Single package, one consignment with patch change
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		prop, err := NewPropagator(g)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"core": {Major: 1, Minor: 2, Patch: 3},
		}

		consignments := []*consignment.Consignment{
			{
				ID:         "test-1",
				Timestamp:  time.Now(),
				Packages:   []string{"core"},
				ChangeType: types.ChangeTypePatch,
				Summary:    "Fix bug",
			},
		}

		result, err := prop.Propagate(currentVersions, consignments)
		require.NoError(t, err)
		assert.Contains(t, result, "core")

		// core should be bumped from 1.2.3 to 1.2.4
		assert.Equal(t, semver.Version{Major: 1, Minor: 2, Patch: 4}, result["core"].NewVersion)
		assert.Equal(t, "patch", result["core"].ChangeType)
	})

	t.Run("no changes for packages without consignments", func(t *testing.T) {
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "a", Path: "./a", Ecosystem: config.EcosystemGo},
				{Name: "b", Path: "./b", Ecosystem: config.EcosystemGo},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		prop, err := NewPropagator(g)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"a": {Major: 1, Minor: 0, Patch: 0},
			"b": {Major: 2, Minor: 0, Patch: 0},
		}

		// No consignments
		result, err := prop.Propagate(currentVersions, []*consignment.Consignment{})
		require.NoError(t, err)

		// No packages should have version changes
		assert.Empty(t, result)
	})

	t.Run("linked dependency propagation", func(t *testing.T) {
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

		prop, err := NewPropagator(g)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"core": {Major: 1, Minor: 0, Patch: 0},
			"api":  {Major: 1, Minor: 0, Patch: 0},
		}

		consignments := []*consignment.Consignment{
			{
				ID:         "test-1",
				Timestamp:  time.Now(),
				Packages:   []string{"core"},
				ChangeType: types.ChangeTypeMinor,
				Summary:    "Add feature",
			},
		}

		result, err := prop.Propagate(currentVersions, consignments)
		require.NoError(t, err)

		// Both should be bumped
		assert.Contains(t, result, "core")
		assert.Contains(t, result, "api")

		// core: direct minor bump 1.0.0 -> 1.1.0
		assert.Equal(t, semver.Version{Major: 1, Minor: 1, Patch: 0}, result["core"].NewVersion)
		assert.Equal(t, "minor", result["core"].ChangeType)

		// api: propagated minor bump 1.0.0 -> 1.1.0
		assert.Equal(t, semver.Version{Major: 1, Minor: 1, Patch: 0}, result["api"].NewVersion)
		assert.Equal(t, "minor", result["api"].ChangeType)
	})

	t.Run("fixed dependency no propagation", func(t *testing.T) {
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

		prop, err := NewPropagator(g)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"core": {Major: 1, Minor: 0, Patch: 0},
			"api":  {Major: 1, Minor: 0, Patch: 0},
		}

		consignments := []*consignment.Consignment{
			{
				ID:         "test-1",
				Timestamp:  time.Now(),
				Packages:   []string{"core"},
				ChangeType: types.ChangeTypeMajor,
				Summary:    "Breaking change",
			},
		}

		result, err := prop.Propagate(currentVersions, consignments)
		require.NoError(t, err)

		// Only core should be bumped (fixed strategy blocks propagation)
		assert.Contains(t, result, "core")
		assert.NotContains(t, result, "api")

		// core: 1.0.0 -> 2.0.0
		assert.Equal(t, semver.Version{Major: 2, Minor: 0, Patch: 0}, result["core"].NewVersion)
	})

	t.Run("multiple consignments for same package", func(t *testing.T) {
		// core has two consignments: patch and minor
		// Should use highest priority (minor)
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		prop, err := NewPropagator(g)
		require.NoError(t, err)

		currentVersions := map[string]semver.Version{
			"core": {Major: 1, Minor: 0, Patch: 0},
		}

		consignments := []*consignment.Consignment{
			{
				ID:         "test-1",
				Timestamp:  time.Now(),
				Packages:   []string{"core"},
				ChangeType: types.ChangeTypePatch,
				Summary:    "Fix bug",
			},
			{
				ID:         "test-2",
				Timestamp:  time.Now(),
				Packages:   []string{"core"},
				ChangeType: types.ChangeTypeMinor,
				Summary:    "Add feature",
			},
		}

		result, err := prop.Propagate(currentVersions, consignments)
		require.NoError(t, err)

		// Should use minor (higher priority than patch)
		assert.Equal(t, semver.Version{Major: 1, Minor: 1, Patch: 0}, result["core"].NewVersion)
		assert.Equal(t, "minor", result["core"].ChangeType)
	})

	t.Run("returns error for missing current version", func(t *testing.T) {
		cfg := &config.Config{
			Packages: []config.Package{
				{Name: "core", Path: "./core", Ecosystem: config.EcosystemGo},
			},
		}

		g, err := graph.BuildGraph(cfg)
		require.NoError(t, err)

		prop, err := NewPropagator(g)
		require.NoError(t, err)

		// Missing version for core
		currentVersions := map[string]semver.Version{}

		consignments := []*consignment.Consignment{
			{
				ID:         "test-1",
				Timestamp:  time.Now(),
				Packages:   []string{"core"},
				ChangeType: types.ChangeTypePatch,
				Summary:    "Fix",
			},
		}

		result, err := prop.Propagate(currentVersions, consignments)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "core")
	})
}

func TestVersionBump(t *testing.T) {
	t.Run("contains expected fields", func(t *testing.T) {
		bump := VersionBump{
			Package:    "core",
			OldVersion: semver.Version{Major: 1, Minor: 0, Patch: 0},
			NewVersion: semver.Version{Major: 1, Minor: 1, Patch: 0},
			ChangeType: "minor",
			Source:     "direct",
		}

		assert.Equal(t, "core", bump.Package)
		assert.Equal(t, semver.Version{Major: 1, Minor: 0, Patch: 0}, bump.OldVersion)
		assert.Equal(t, semver.Version{Major: 1, Minor: 1, Patch: 0}, bump.NewVersion)
		assert.Equal(t, "minor", bump.ChangeType)
		assert.Equal(t, "direct", bump.Source)
	})
}
