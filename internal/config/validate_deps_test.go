package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateDependencies(t *testing.T) {
	t.Run("valid dependencies - all packages exist", func(t *testing.T) {
		cfg := &Config{
			Packages: []Package{
				{Name: "core", Path: "./core", Ecosystem: EcosystemGo},
				{Name: "api", Path: "./api", Ecosystem: EcosystemGo,
					Dependencies: []Dependency{
						{Package: "core", Strategy: "linked"},
					},
				},
			},
		}

		err := ValidateDependencies(cfg)
		assert.NoError(t, err)
	})

	t.Run("dangling reference - dependency not found", func(t *testing.T) {
		cfg := &Config{
			Packages: []Package{
				{Name: "api", Path: "./api", Ecosystem: EcosystemGo,
					Dependencies: []Dependency{
						{Package: "nonexistent", Strategy: "linked"},
					},
				},
			},
		}

		err := ValidateDependencies(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nonexistent")
		assert.Contains(t, err.Error(), "api")
	})

	t.Run("multiple dangling references", func(t *testing.T) {
		cfg := &Config{
			Packages: []Package{
				{Name: "api", Path: "./api", Ecosystem: EcosystemGo,
					Dependencies: []Dependency{
						{Package: "core", Strategy: "linked"},
						{Package: "utils", Strategy: "linked"},
					},
				},
			},
		}

		err := ValidateDependencies(cfg)
		assert.Error(t, err)
		// Should mention at least one of the missing packages
		errMsg := err.Error()
		assert.True(t,
			assert.ObjectsAreEqual(true, assert.Contains(t, errMsg, "core")) ||
			assert.ObjectsAreEqual(true, assert.Contains(t, errMsg, "utils")),
		)
	})

	t.Run("self-dependency is allowed", func(t *testing.T) {
		cfg := &Config{
			Packages: []Package{
				{Name: "pkg", Path: "./pkg", Ecosystem: EcosystemGo,
					Dependencies: []Dependency{
						{Package: "pkg", Strategy: "linked"},
					},
				},
			},
		}

		err := ValidateDependencies(cfg)
		assert.NoError(t, err)
	})

	t.Run("circular dependencies are allowed", func(t *testing.T) {
		cfg := &Config{
			Packages: []Package{
				{Name: "a", Path: "./a", Ecosystem: EcosystemGo,
					Dependencies: []Dependency{
						{Package: "b", Strategy: "linked"},
					},
				},
				{Name: "b", Path: "./b", Ecosystem: EcosystemGo,
					Dependencies: []Dependency{
						{Package: "a", Strategy: "linked"},
					},
				},
			},
		}

		err := ValidateDependencies(cfg)
		assert.NoError(t, err)
	})

	t.Run("empty config", func(t *testing.T) {
		cfg := &Config{
			Packages: []Package{},
		}

		err := ValidateDependencies(cfg)
		assert.NoError(t, err)
	})

	t.Run("nil config", func(t *testing.T) {
		err := ValidateDependencies(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil")
	})

	t.Run("package with no dependencies", func(t *testing.T) {
		cfg := &Config{
			Packages: []Package{
				{Name: "standalone", Path: "./standalone", Ecosystem: EcosystemGo},
			},
		}

		err := ValidateDependencies(cfg)
		assert.NoError(t, err)
	})

	t.Run("complex valid dependency graph", func(t *testing.T) {
		cfg := &Config{
			Packages: []Package{
				{Name: "utils", Path: "./utils", Ecosystem: EcosystemGo},
				{Name: "logging", Path: "./logging", Ecosystem: EcosystemGo},
				{Name: "core", Path: "./core", Ecosystem: EcosystemGo,
					Dependencies: []Dependency{
						{Package: "utils", Strategy: "linked"},
						{Package: "logging", Strategy: "linked"},
					},
				},
				{Name: "api", Path: "./api", Ecosystem: EcosystemGo,
					Dependencies: []Dependency{
						{Package: "core", Strategy: "linked"},
					},
				},
				{Name: "web", Path: "./web", Ecosystem: EcosystemNPM,
					Dependencies: []Dependency{
						{Package: "api", Strategy: "fixed"},
					},
				},
			},
		}

		err := ValidateDependencies(cfg)
		assert.NoError(t, err)
	})

	t.Run("mixed valid and invalid dependencies", func(t *testing.T) {
		cfg := &Config{
			Packages: []Package{
				{Name: "core", Path: "./core", Ecosystem: EcosystemGo},
				{Name: "api", Path: "./api", Ecosystem: EcosystemGo,
					Dependencies: []Dependency{
						{Package: "core", Strategy: "linked"},      // Valid
						{Package: "missing", Strategy: "linked"},   // Invalid
					},
				},
			},
		}

		err := ValidateDependencies(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing")
	})

	t.Run("validate returns all errors", func(t *testing.T) {
		cfg := &Config{
			Packages: []Package{
				{Name: "a", Path: "./a", Ecosystem: EcosystemGo,
					Dependencies: []Dependency{
						{Package: "x", Strategy: "linked"}, // Missing
					},
				},
				{Name: "b", Path: "./b", Ecosystem: EcosystemGo,
					Dependencies: []Dependency{
						{Package: "y", Strategy: "linked"}, // Missing
					},
				},
			},
		}

		err := ValidateDependencies(cfg)
		require.Error(t, err)

		// Should contain information about both errors
		errMsg := err.Error()
		containsX := assert.Contains(t, errMsg, "x") || assert.Contains(t, errMsg, "a")
		containsY := assert.Contains(t, errMsg, "y") || assert.Contains(t, errMsg, "b")

		assert.True(t, containsX || containsY, "Error should mention at least one missing dependency")
	})
}
