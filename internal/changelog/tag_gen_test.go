package changelog

import (
	"testing"
	"time"

	"github.com/NatoNathan/shipyard/internal/consignment"
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/NatoNathan/shipyard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGeneratePackageTag_Simple tests generating simple tag name (lightweight tag)
func TestGeneratePackageTag_Simple(t *testing.T) {
	now := time.Now()
	consignments := []*consignment.Consignment{
		{
			ID:         "c1",
			Timestamp:  now,
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypeMinor,
			Summary:    "Added OAuth2",
		},
	}

	version := semver.Version{Major: 1, Minor: 2, Patch: 0}

	// Use simple template that returns just tag name
	template := "v{{ .Version }}"

	generator := NewChangelogGenerator()
	tagName, message, err := generator.GeneratePackageTagWithContext(
		consignments,
		"core",
		version,
		template,
	)

	require.NoError(t, err)
	assert.Equal(t, "v1.2.0", tagName)
	assert.Equal(t, "", message) // No message = lightweight tag
}

// TestGeneratePackageTag_Annotated tests generating annotated tag with message
func TestGeneratePackageTag_Annotated(t *testing.T) {
	now := time.Now()
	consignments := []*consignment.Consignment{
		{
			ID:         "c1",
			Timestamp:  now,
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypeMinor,
			Summary:    "Added OAuth2 support",
		},
		{
			ID:         "c2",
			Timestamp:  now.Add(time.Hour),
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "Fixed validation bug",
		},
	}

	version := semver.Version{Major: 1, Minor: 2, Patch: 0}

	// Template that returns tag name + message
	template := `core/v{{ .Version }}

# Release core v{{ .Version }}

{{ range .Consignments -}}
- {{ .Summary }}
{{ end -}}`

	generator := NewChangelogGenerator()
	tagName, message, err := generator.GeneratePackageTagWithContext(
		consignments,
		"core",
		version,
		template,
	)

	require.NoError(t, err)
	assert.Equal(t, "core/v1.2.0", tagName)
	assert.Contains(t, message, "# Release core v1.2.0")
	assert.Contains(t, message, "Added OAuth2 support")
	assert.Contains(t, message, "Fixed validation bug")
}

// TestGeneratePackageTag_PackageFiltering tests that consignments are filtered
func TestGeneratePackageTag_PackageFiltering(t *testing.T) {
	now := time.Now()
	consignments := []*consignment.Consignment{
		{
			ID:         "c1",
			Timestamp:  now,
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypeMinor,
			Summary:    "Core feature",
		},
		{
			ID:         "c2",
			Timestamp:  now,
			Packages:   []string{"api"},
			ChangeType: types.ChangeTypeMajor,
			Summary:    "API breaking change",
		},
		{
			ID:         "c3",
			Timestamp:  now,
			Packages:   []string{"core", "api"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "Shared fix",
		},
	}

	version := semver.Version{Major: 1, Minor: 1, Patch: 0}

	// Template that lists consignments
	template := `core/v{{ .Version }}

{{ range .Consignments -}}
- {{ .Summary }}
{{ end -}}`

	generator := NewChangelogGenerator()
	tagName, message, err := generator.GeneratePackageTagWithContext(
		consignments,
		"core",
		version,
		template,
	)

	require.NoError(t, err)
	assert.Equal(t, "core/v1.1.0", tagName)
	// Should include core-only and shared
	assert.Contains(t, message, "Core feature")
	assert.Contains(t, message, "Shared fix")
	// Should NOT include api-only
	assert.NotContains(t, message, "API breaking change")
}

// TestGeneratePackageTag_GoMonorepo tests Go-style package tag
func TestGeneratePackageTag_GoMonorepo(t *testing.T) {
	now := time.Now()
	consignments := []*consignment.Consignment{
		{
			ID:         "c1",
			Timestamp:  now,
			Packages:   []string{"api"},
			ChangeType: types.ChangeTypeMajor,
			Summary:    "Breaking: New API",
		},
	}

	version := semver.Version{Major: 2, Minor: 0, Patch: 0}

	generator := NewChangelogGenerator()
	tagName, message, err := generator.GeneratePackageTag(
		consignments,
		"api",
		version,
		"builtin:go",
	)

	require.NoError(t, err)
	assert.Equal(t, "api/v2.0.0", tagName)
	// builtin:go should produce lightweight tag
	assert.Equal(t, "", message)
}

func TestBuiltinGoAnnotated(t *testing.T) {
	now := time.Now()
	consignments := []*consignment.Consignment{
		{
			ID:         "c1",
			Timestamp:  now,
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypeMinor,
			Summary:    "Added OAuth2 support",
			Metadata:   map[string]interface{}{"author": "alice@example.com"},
		},
	}

	version := semver.Version{Major: 1, Minor: 2, Patch: 0}

	generator := NewChangelogGenerator()
	tagName, message, err := generator.GeneratePackageTag(
		consignments,
		"core",
		version,
		"builtin:go-annotated",
	)

	require.NoError(t, err)
	assert.Equal(t, "core/v1.2.0", tagName)
	assert.Contains(t, message, "# Release core v1.2.0")
	assert.Contains(t, message, "Added OAuth2 support")
	assert.Contains(t, message, "alice@example.com")
}
