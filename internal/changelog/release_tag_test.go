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

// TestGenerateReleaseTag_WithConsignments tests release tag with full consignment context
func TestGenerateReleaseTag_WithConsignments(t *testing.T) {
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
			Timestamp:  now,
			Packages:   []string{"api"},
			ChangeType: types.ChangeTypeMajor,
			Summary:    "Breaking API change",
		},
		{
			ID:         "c3",
			Timestamp:  now,
			Packages:   []string{"core", "api"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "Shared bug fix",
		},
	}

	versions := map[string]semver.Version{
		"core": {Major: 1, Minor: 2, Patch: 0},
		"api":  {Major: 2, Minor: 0, Patch: 0},
	}

	packages := []string{"core", "api"}

	// Template that includes consignments in release tag
	template := `release-{{ .Date | date "20060102" }}

# Release {{ .Date | date "2006-01-02" }}

Packages: core v1.2.0, api v2.0.0

{{ range .Consignments -}}
- [{{ .ChangeType }}] {{ .Summary }}
{{ end -}}`

	generator := NewChangelogGenerator()
	tagName, message, err := generator.GenerateReleaseTagWithContext(
		consignments,
		packages,
		versions,
		template,
	)

	require.NoError(t, err)
	assert.Contains(t, tagName, "release-")
	assert.Contains(t, message, "# Release")
	assert.Contains(t, message, "Added OAuth2 support")
	assert.Contains(t, message, "Breaking API change")
	assert.Contains(t, message, "Shared bug fix")
}

// TestGenerateReleaseTag_Lightweight tests simple release tag (date-based)
func TestGenerateReleaseTag_Lightweight(t *testing.T) {
	versions := map[string]semver.Version{
		"core": {Major: 1, Minor: 2, Patch: 0},
		"api":  {Major: 2, Minor: 0, Patch: 0},
	}

	packages := []string{"core", "api"}

	generator := NewChangelogGenerator()
	tagName, message, err := generator.GenerateReleaseTag(
		[]*consignment.Consignment{},
		packages,
		versions,
		"builtin:date",
	)

	require.NoError(t, err)
	assert.Contains(t, tagName, "release-")
	assert.Regexp(t, `^release-\d{8}-\d{6}$`, tagName)
	assert.Equal(t, "", message) // Lightweight tag
}
