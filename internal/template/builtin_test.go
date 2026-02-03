package template

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuiltinTemplate_Changelog(t *testing.T) {
	now := time.Date(2026, 1, 30, 14, 30, 0, 0, time.UTC)

	type Consignment struct {
		Packages   []string
		ChangeType string
		Summary    string
		Metadata   map[string]interface{}
	}

	type Entry struct {
		Package      string
		Version      string
		Timestamp    time.Time
		Consignments []Consignment
	}

	// Changelog template expects array of entries (multi-version)
	context := []Entry{
		{
			Package:   "core",
			Version:   "1.2.0",
			Timestamp: now,
			Consignments: []Consignment{
				{
					Packages:   []string{"core"},
					ChangeType: "minor",
					Summary:    "Added OAuth2 support",
					Metadata: map[string]interface{}{
						"author": "alice@example.com",
						"issue":  "FEAT-123",
					},
				},
				{
					Packages:   []string{"core"},
					ChangeType: "patch",
					Summary:    "Fixed bug in validation",
					Metadata:   map[string]interface{}{},
				},
			},
		},
	}

	loader := NewTemplateLoader()
	content, err := loader.Load("builtin:default", TemplateTypeChangelog)
	require.NoError(t, err)

	renderer := NewTemplateRenderer()
	result, err := renderer.Render(content, context)

	require.NoError(t, err)
	assert.Contains(t, result, "# Changelog")
	assert.Contains(t, result, "[1.2.0]")
	assert.Contains(t, result, "OAuth2")
}

func TestBuiltinTemplate_TagName(t *testing.T) {
	context := map[string]interface{}{
		"Package": "core",
		"Version": "2.5.10",
	}

	loader := NewTemplateLoader()
	content, err := loader.Load("builtin:default", TemplateTypeTag)
	require.NoError(t, err)

	renderer := NewTemplateRenderer()
	result, err := renderer.Render(content, context)

	require.NoError(t, err)
	assert.Equal(t, "v2.5.10", result)
}

func TestBuiltinTemplate_ReleaseNotes(t *testing.T) {
	now := time.Date(2026, 1, 30, 14, 30, 0, 0, time.UTC)

	type Consignment struct {
		ChangeType string
		Summary    string
	}

	context := map[string]interface{}{
		"Package":   "core",
		"Version":   "1.5.0",
		"Timestamp": now, // Changed from Date to Timestamp
		"Consignments": []Consignment{
			{
				ChangeType: "minor",
				Summary:    "Added new feature",
			},
			{
				ChangeType: "patch",
				Summary:    "Fixed bug",
			},
		},
	}

	loader := NewTemplateLoader()
	content, err := loader.Load("builtin:default", TemplateTypeReleaseNotes)
	require.NoError(t, err)

	renderer := NewTemplateRenderer()
	result, err := renderer.Render(content, context)

	require.NoError(t, err)
	assert.Contains(t, result, "core v1.5.0") // Updated format
	assert.Contains(t, result, "2026-01-30")  // Updated date format
	assert.Contains(t, result, "Added new feature")
	assert.Contains(t, result, "Fixed bug")
}

func TestBuiltinTemplate_AllTemplatesValid(t *testing.T) {
	allTemplates, err := GetAllBuiltinTemplates()
	require.NoError(t, err)

	parser := NewTemplateParser()

	for templateType, templates := range allTemplates {
		for name, content := range templates {
			testName := fmt.Sprintf("%s/%s", templateType, name)
			t.Run(testName, func(t *testing.T) {
				_, err := parser.Parse(testName, content)
				assert.NoError(t, err, "builtin template %s should parse without error", testName)
			})
		}
	}
}

func TestBuiltinTemplate_ChangelogWithSharedConsignments(t *testing.T) {
	now := time.Date(2026, 1, 30, 14, 30, 0, 0, time.UTC)

	type Consignment struct {
		Packages   []string
		ChangeType string
		Summary    string
		Metadata   map[string]interface{}
	}

	type Entry struct {
		Package      string
		Version      string
		Timestamp    time.Time
		Consignments []Consignment
	}

	// Test that consignments affecting multiple packages can appear in single-package changelog
	context := []Entry{
		{
			Package:   "core",
			Version:   "1.2.0",
			Timestamp: now,
			Consignments: []Consignment{
				{
					Packages:   []string{"core"},
					ChangeType: "minor",
					Summary:    "Core feature",
					Metadata:   map[string]interface{}{},
				},
				{
					Packages:   []string{"core", "api"},
					ChangeType: "patch",
					Summary:    "Shared fix",
					Metadata:   map[string]interface{}{},
				},
			},
		},
	}

	loader := NewTemplateLoader()
	content, err := loader.Load("builtin:default", TemplateTypeChangelog)
	require.NoError(t, err)

	renderer := NewTemplateRenderer()
	result, err := renderer.Render(content, context)

	require.NoError(t, err)
	assert.Contains(t, result, "[1.2.0]")
	assert.Contains(t, result, "Core feature")
	assert.Contains(t, result, "Shared fix")
}

func TestBuiltinTemplate_EmptyConsignments(t *testing.T) {
	now := time.Date(2026, 1, 30, 14, 30, 0, 0, time.UTC)

	type Entry struct {
		Package      string
		Version      string
		Timestamp    time.Time
		Consignments []interface{}
	}

	context := []Entry{
		{
			Package:      "core",
			Version:      "1.0.0",
			Timestamp:    now,
			Consignments: []interface{}{}, // Empty
		},
	}

	loader := NewTemplateLoader()
	content, err := loader.Load("builtin:default", TemplateTypeChangelog)
	require.NoError(t, err)

	renderer := NewTemplateRenderer()
	result, err := renderer.Render(content, context)

	// Should render without error even with empty consignments
	require.NoError(t, err)
	assert.Contains(t, result, "# Changelog")
	// Entry with no consignments should be skipped by template
	assert.NotContains(t, result, "[1.0.0]")
}

func TestBuiltinTemplate_ConsistentFormatting(t *testing.T) {
	// Test that builtin templates produce consistent, well-formatted output
	t.Run("changelog has proper markdown structure", func(t *testing.T) {
		template, err := GetDefaultChangelogTemplate()
		require.NoError(t, err)

		// Should have markdown headers
		assert.Contains(t, template, "# Changelog")
		assert.Contains(t, template, "##")

		// Should use proper template syntax
		assert.Contains(t, template, "{{")
		assert.Contains(t, template, "}}")
	})

	t.Run("tagname is simple and clean", func(t *testing.T) {
		template, err := GetDefaultTagTemplate()
		require.NoError(t, err)

		// Should be a simple one-liner
		assert.NotContains(t, template, "\n")
		assert.Contains(t, template, "v{{")
	})

	t.Run("release notes has proper structure", func(t *testing.T) {
		template, err := GetDefaultReleaseNotesTemplate()
		require.NoError(t, err)

		// Should have title
		assert.Contains(t, template, "# Release")

		// Should have date (updated format)
		assert.Contains(t, template, "Released:")

		// Should have changes section
		assert.Contains(t, template, "## Changes")
	})
}
