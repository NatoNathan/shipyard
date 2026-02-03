package changelog

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NatoNathan/shipyard/internal/consignment"
	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/NatoNathan/shipyard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateChangelog_Basic(t *testing.T) {
	now := time.Date(2026, 1, 30, 14, 30, 0, 0, time.UTC)

	consignments := []*consignment.Consignment{
		{
			ID:         "c1",
			Timestamp:  now,
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypeMinor,
			Summary:    "Added OAuth2 support",
			Metadata: map[string]interface{}{
				"author": "alice@example.com",
			},
		},
		{
			ID:         "c2",
			Timestamp:  now.Add(time.Hour),
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "Fixed validation bug",
			Metadata:   map[string]interface{}{},
		},
	}

	version := semver.Version{Major: 1, Minor: 2, Patch: 0}

	generator := NewChangelogGenerator()
	result, err := generator.GenerateForPackage(consignments, "core", version, "builtin:default")

	require.NoError(t, err)
	assert.Contains(t, result, "# Changelog")
	assert.Contains(t, result, "1.2.0")
	assert.Contains(t, result, "OAuth2")
	assert.Contains(t, result, "validation bug")
}

func TestGenerateChangelog_MultiplePackages(t *testing.T) {
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
	}

	versions := map[string]semver.Version{
		"core": {Major: 1, Minor: 1, Patch: 0},
		"api":  {Major: 2, Minor: 0, Patch: 0},
	}

	generator := NewChangelogGenerator()
	results, err := generator.GenerateAll(consignments, versions, "builtin:default")

	require.NoError(t, err)
	require.Len(t, results, 2)

	// Check core changelog
	coreChangelog := results["core"]
	assert.Contains(t, coreChangelog, "1.1.0")
	assert.Contains(t, coreChangelog, "Core feature")
	assert.NotContains(t, coreChangelog, "API breaking change")

	// Check api changelog
	apiChangelog := results["api"]
	assert.Contains(t, apiChangelog, "2.0.0")
	assert.Contains(t, apiChangelog, "API breaking change")
	assert.NotContains(t, apiChangelog, "Core feature")
}

func TestGenerateChangelog_PackageFiltering(t *testing.T) {
	now := time.Now()

	consignments := []*consignment.Consignment{
		{
			ID:         "c1",
			Timestamp:  now,
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypeMinor,
			Summary:    "Core-only feature",
		},
		{
			ID:         "c2",
			Timestamp:  now,
			Packages:   []string{"api"},
			ChangeType: types.ChangeTypeMajor,
			Summary:    "API-only breaking change",
		},
		{
			ID:         "c3",
			Timestamp:  now,
			Packages:   []string{"core", "api"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "Shared fix",
		},
	}

	coreVersion := semver.Version{Major: 1, Minor: 1, Patch: 0}
	apiVersion := semver.Version{Major: 2, Minor: 0, Patch: 0}

	generator := NewChangelogGenerator()

	// Generate core changelog - should include c1 and c3
	coreChangelog, err := generator.GenerateForPackage(consignments, "core", coreVersion, "builtin:default")
	require.NoError(t, err)
	assert.Contains(t, coreChangelog, "Core-only feature")
	assert.Contains(t, coreChangelog, "Shared fix")
	assert.NotContains(t, coreChangelog, "API-only breaking change")

	// Generate api changelog - should include c2 and c3
	apiChangelog, err := generator.GenerateForPackage(consignments, "api", apiVersion, "builtin:default")
	require.NoError(t, err)
	assert.Contains(t, apiChangelog, "API-only breaking change")
	assert.Contains(t, apiChangelog, "Shared fix")
	assert.NotContains(t, apiChangelog, "Core-only feature")
}

func TestGenerateChangelog_CustomTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create custom template (must iterate over array of entries)
	customTemplate := `# My Custom Changelog

{{- range . }}
{{- range .Consignments }}
* {{ .Summary }} ({{ .ChangeType }})
{{- end }}
{{- end }}
`

	templatePath := filepath.Join(tmpDir, "custom.tmpl")
	require.NoError(t, os.WriteFile(templatePath, []byte(customTemplate), 0644))

	consignments := []*consignment.Consignment{
		{
			ID:         "c1",
			Timestamp:  time.Now(),
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "Fixed issue",
		},
	}

	version := semver.Version{Major: 1, Minor: 0, Patch: 1}

	generator := NewChangelogGenerator()
	generator.SetBaseDir(tmpDir)

	result, err := generator.GenerateForPackage(consignments, "core", version, "file:custom.tmpl")

	require.NoError(t, err)
	assert.Contains(t, result, "# My Custom Changelog")
	assert.Contains(t, result, "Fixed issue")
	assert.Contains(t, result, "patch")
}

func TestGenerateChangelog_WithMetadata(t *testing.T) {
	consignments := []*consignment.Consignment{
		{
			ID:         "c1",
			Timestamp:  time.Now(),
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypeMinor,
			Summary:    "New feature",
			Metadata: map[string]interface{}{
				"author":   "bob@example.com",
				"issue":    "JIRA-456",
				"breaking": false,
			},
		},
	}

	version := semver.Version{Major: 1, Minor: 3, Patch: 0}

	generator := NewChangelogGenerator()
	result, err := generator.GenerateForPackage(consignments, "core", version, "builtin:default")

	require.NoError(t, err)
	// New changelog template groups by change type and doesn't show metadata
	// Metadata is stored but not displayed in the default template
	assert.Contains(t, result, "New feature")
	assert.Contains(t, result, "# Changelog")
}

func TestGenerateChangelog_EmptyConsignments(t *testing.T) {
	consignments := []*consignment.Consignment{}
	version := semver.Version{Major: 1, Minor: 0, Patch: 0}

	generator := NewChangelogGenerator()
	result, err := generator.GenerateForPackage(consignments, "core", version, "builtin:default")

	require.NoError(t, err)
	assert.Contains(t, result, "# Changelog")
	// Should render structure even without consignments
}

func TestGenerateChangelog_GroupedByChangeType(t *testing.T) {
	now := time.Now()

	consignments := []*consignment.Consignment{
		{
			ID:         "c1",
			Timestamp:  now,
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypeMajor,
			Summary:    "Breaking change 1",
		},
		{
			ID:         "c2",
			Timestamp:  now,
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypeMinor,
			Summary:    "Feature 1",
		},
		{
			ID:         "c3",
			Timestamp:  now,
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "Fix 1",
		},
	}

	version := semver.Version{Major: 2, Minor: 0, Patch: 0}

	generator := NewChangelogGenerator()
	result, err := generator.GenerateForPackage(consignments, "core", version, "builtin:default")

	require.NoError(t, err)
	// All change types should be present
	assert.Contains(t, result, "Breaking change 1")
	assert.Contains(t, result, "Feature 1")
	assert.Contains(t, result, "Fix 1")
}

func TestGenerateChangelog_InvalidTemplate(t *testing.T) {
	consignments := []*consignment.Consignment{
		{
			ID:         "c1",
			Timestamp:  time.Now(),
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "Fix",
		},
	}

	version := semver.Version{Major: 1, Minor: 0, Patch: 1}

	generator := NewChangelogGenerator()
	_, err := generator.GenerateForPackage(consignments, "core", version, "file:/nonexistent/template.tmpl")

	assert.Error(t, err)
}

func TestWriteChangelog_ToFile(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

	consignments := []*consignment.Consignment{
		{
			ID:         "c1",
			Timestamp:  time.Now(),
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "Fixed bug",
		},
	}

	version := semver.Version{Major: 1, Minor: 0, Patch: 1}

	generator := NewChangelogGenerator()
	err := generator.WriteChangelogToFile(consignments, "core", version, "builtin:default", changelogPath)

	require.NoError(t, err)

	// Verify file was created
	content, err := os.ReadFile(changelogPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Changelog")
	assert.Contains(t, string(content), "Fixed bug")
}

func TestGenerateChangelog_PreserveExisting(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

	// Create existing changelog
	existingContent := `# Changelog

## [1.0.0] - 2026-01-15

- Initial release
`
	require.NoError(t, os.WriteFile(changelogPath, []byte(existingContent), 0644))

	consignments := []*consignment.Consignment{
		{
			ID:         "c1",
			Timestamp:  time.Now(),
			Packages:   []string{"core"},
			ChangeType: types.ChangeTypePatch,
			Summary:    "New fix",
		},
	}

	version := semver.Version{Major: 1, Minor: 0, Patch: 1}

	generator := NewChangelogGenerator()
	generator.SetPreserveExisting(true)

	err := generator.WriteChangelogToFile(consignments, "core", version, "builtin:default", changelogPath)

	require.NoError(t, err)

	// Read updated changelog
	content, err := os.ReadFile(changelogPath)
	require.NoError(t, err)

	// Should contain both old and new content
	contentStr := string(content)
	assert.Contains(t, contentStr, "Initial release")
	assert.Contains(t, contentStr, "New fix")
	assert.Contains(t, contentStr, "1.0.0")
	assert.Contains(t, contentStr, "1.0.1")
}

func TestGeneratePackageTag(t *testing.T) {
	version := semver.Version{Major: 1, Minor: 5, Patch: 2}

	generator := NewChangelogGenerator()
	tagName, message, err := generator.GeneratePackageTag([]*consignment.Consignment{}, "core", version, "builtin:default")

	require.NoError(t, err)
	assert.Equal(t, "v1.5.2", tagName)
	assert.Equal(t, "", message) // Lightweight tag
}

func TestGeneratePackageTag_Go(t *testing.T) {
	version := semver.Version{Major: 2, Minor: 0, Patch: 0}

	generator := NewChangelogGenerator()
	tagName, message, err := generator.GeneratePackageTag([]*consignment.Consignment{}, "api", version, "builtin:go")

	require.NoError(t, err)
	assert.Equal(t, "api/v2.0.0", tagName)
	assert.Equal(t, "", message) // Lightweight tag
}

func TestGeneratePackageTag_NPM(t *testing.T) {
	version := semver.Version{Major: 1, Minor: 3, Patch: 5}

	generator := NewChangelogGenerator()
	tagName, message, err := generator.GeneratePackageTag([]*consignment.Consignment{}, "@scope/package", version, "builtin:npm")

	require.NoError(t, err)
	assert.Equal(t, "@scope/package@1.3.5", tagName)
	assert.Equal(t, "", message) // Lightweight tag
}

func TestGenerateAllPackageTags(t *testing.T) {
	versions := map[string]semver.Version{
		"core": {Major: 1, Minor: 5, Patch: 2},
		"api":  {Major: 2, Minor: 0, Patch: 0},
	}

	generator := NewChangelogGenerator()
	tags, err := generator.GenerateAllPackageTags([]*consignment.Consignment{}, versions, "builtin:go")

	require.NoError(t, err)
	require.Len(t, tags, 2)
	assert.Equal(t, "core/v1.5.2", tags["core"].Name)
	assert.Equal(t, "", tags["core"].Message) // Lightweight
	assert.Equal(t, "api/v2.0.0", tags["api"].Name)
	assert.Equal(t, "", tags["api"].Message) // Lightweight
}

func TestGenerateReleaseTag_Date(t *testing.T) {
	versions := map[string]semver.Version{
		"core": {Major: 1, Minor: 5, Patch: 2},
		"api":  {Major: 2, Minor: 0, Patch: 0},
	}

	packages := []string{"core", "api"}

	generator := NewChangelogGenerator()
	tagName, message, err := generator.GenerateReleaseTag([]*consignment.Consignment{}, packages, versions, "builtin:date")

	require.NoError(t, err)
	assert.Contains(t, tagName, "release-")
	// Should contain date in format YYYYMMDD-HHMMSS
	assert.Regexp(t, `^release-\d{8}-\d{6}$`, tagName)
	assert.Equal(t, "", message) // Lightweight tag
}

func TestGenerateReleaseTag_Versions(t *testing.T) {
	versions := map[string]semver.Version{
		"core": {Major: 1, Minor: 5, Patch: 2},
		"api":  {Major: 2, Minor: 0, Patch: 0},
	}

	packages := []string{"core", "api"}

	generator := NewChangelogGenerator()
	tagName, message, err := generator.GenerateReleaseTag([]*consignment.Consignment{}, packages, versions, "builtin:versions")

	require.NoError(t, err)
	// Should contain both package names and versions
	assert.Contains(t, tagName, "release-")
	assert.Contains(t, tagName, "core")
	assert.Contains(t, tagName, "api")
	assert.Contains(t, tagName, "1.5.2")
	assert.Contains(t, tagName, "2.0.0")
	assert.Equal(t, "", message) // Lightweight tag
}

func TestGeneratePackageTag_CustomTemplate(t *testing.T) {
	version := semver.Version{Major: 2, Minor: 0, Patch: 0}

	// Custom template with package name prefix
	customTemplate := `{{ .Package }}/v{{ .Version }}`

	generator := NewChangelogGenerator()
	tagName, message, err := generator.GeneratePackageTagWithContext([]*consignment.Consignment{}, "api", version, customTemplate)

	require.NoError(t, err)
	assert.Equal(t, "api/v2.0.0", tagName)
	assert.Equal(t, "", message) // Lightweight tag
}
