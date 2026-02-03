package template

import (
	"strings"
	"testing"
	"time"

	"github.com/NatoNathan/shipyard/internal/history"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRenderReleaseNotes tests the builtin release notes template
func TestRenderReleaseNotes(t *testing.T) {
	timestamp := time.Date(2026, 1, 30, 10, 0, 0, 0, time.UTC)

	t.Run("renders single version with changes", func(t *testing.T) {
		// Setup: Create history entry
		entries := []history.Entry{
			{
				Version:   "1.1.0",
				Package:   "core",
				Timestamp: timestamp,
				Consignments: []history.Consignment{
					{ID: "c1", Summary: "Add new feature", ChangeType: "minor"},
					{ID: "c2", Summary: "Fix critical bug", ChangeType: "patch"},
				},
			},
		}

		// Test: Render release notes using default template
		output, err := RenderReleaseNotes(entries)

		// Verify: Valid markdown output from default template
		require.NoError(t, err)
		assert.Contains(t, output, "# Release Notes", "Should have title")
		assert.Contains(t, output, "core v1.1.0", "Should have package and version")
		assert.Contains(t, output, "2026-01-30", "Should have release date")
		assert.Contains(t, output, "## Changes", "Should have changes section")
		assert.Contains(t, output, "Minor", "Should show change type")
		assert.Contains(t, output, "Add new feature", "Should list changes")
		assert.Contains(t, output, "Fix critical bug", "Should list all changes")
	})

	t.Run("renders version without consignments", func(t *testing.T) {
		// Setup: Create entry with no consignments
		entries := []history.Entry{
			{
				Version:      "1.0.0",
				Package:      "core",
				Timestamp:    timestamp,
				Consignments: []history.Consignment{},
			},
		}

		// Test: Render release notes
		output, err := RenderReleaseNotes(entries)

		// Verify: Shows "no changes" message from template
		require.NoError(t, err)
		assert.Contains(t, output, "core v1.0.0")
		assert.Contains(t, output, "No changes", "Should show no changes message")
	})

	t.Run("returns message for empty history", func(t *testing.T) {
		// Setup: Empty entries
		entries := []history.Entry{}

		// Test: Render release notes
		output, err := RenderReleaseNotes(entries)

		// Verify: Returns appropriate message
		require.NoError(t, err)
		assert.Contains(t, output, "No releases found", "Should indicate no releases")
	})
}

// TestRenderReleaseNotesWithTemplate tests template selection
func TestRenderReleaseNotesWithTemplate(t *testing.T) {
	timestamp := time.Date(2026, 1, 30, 10, 0, 0, 0, time.UTC)

	t.Run("uses builtin grouped template", func(t *testing.T) {
		// Setup: Create entries with multiple change types
		entries := []history.Entry{
			{
				Version:   "2.0.0",
				Package:   "api",
				Timestamp: timestamp,
				Consignments: []history.Consignment{
					{ID: "c1", Summary: "Breaking change", ChangeType: "major"},
					{ID: "c2", Summary: "New feature", ChangeType: "minor"},
					{ID: "c3", Summary: "Bug fix", ChangeType: "patch"},
				},
			},
		}

		// Test: Render with grouped template
		output, err := RenderReleaseNotesWithTemplate(entries, "builtin:grouped")

		// Verify: Changes grouped by type
		require.NoError(t, err)
		assert.Contains(t, output, "## Breaking Changes")
		assert.Contains(t, output, "## Features")
		assert.Contains(t, output, "## Bug Fixes")
		assert.Contains(t, output, "Breaking change")
		assert.Contains(t, output, "New feature")
		assert.Contains(t, output, "Bug fix")
	})

	t.Run("auto-selects changelog template for multiple entries", func(t *testing.T) {
		// Setup: Create multiple history entries
		entries := []history.Entry{
			{
				Version:   "1.1.0",
				Package:   "core",
				Timestamp: timestamp,
				Consignments: []history.Consignment{
					{ID: "c1", Summary: "Add feature", ChangeType: "minor"},
				},
			},
			{
				Version:   "1.0.1",
				Package:   "core",
				Timestamp: timestamp.Add(-24 * time.Hour),
				Consignments: []history.Consignment{
					{ID: "c2", Summary: "Fix bug", ChangeType: "patch"},
				},
			},
		}

		// Test: Render with multiple entries (should use changelog template)
		output, err := RenderReleaseNotesWithTemplate(entries, "builtin:default")

		// Verify: Both versions present in changelog format
		require.NoError(t, err)
		assert.Contains(t, output, "# Changelog")
		assert.Contains(t, output, "[1.1.0]")
		assert.Contains(t, output, "[1.0.1]")
		assert.Contains(t, output, "Add feature")
		assert.Contains(t, output, "Fix bug")
	})

	t.Run("uses changelog keyword", func(t *testing.T) {
		// Setup: Create multiple entries
		entries := []history.Entry{
			{
				Version:   "1.1.0",
				Package:   "core",
				Timestamp: timestamp,
				Consignments: []history.Consignment{
					{ID: "c1", Summary: "Add feature", ChangeType: "minor"},
				},
			},
		}

		// Test: Use "changelog" keyword
		output, err := RenderReleaseNotesWithTemplate(entries, "changelog")

		// Verify: Uses changelog template
		require.NoError(t, err)
		assert.Contains(t, output, "# Changelog")
	})

	t.Run("uses release-notes keyword", func(t *testing.T) {
		// Setup: Create single entry
		entries := []history.Entry{
			{
				Version:   "1.1.0",
				Package:   "core",
				Timestamp: timestamp,
				Consignments: []history.Consignment{
					{ID: "c1", Summary: "Add feature", ChangeType: "minor"},
				},
			},
		}

		// Test: Use "release-notes" keyword
		output, err := RenderReleaseNotesWithTemplate(entries, "release-notes")

		// Verify: Uses release notes template
		require.NoError(t, err)
		assert.Contains(t, output, "# Release Notes")
	})
}

// TestMarkdownFormatting tests markdown output quality
func TestMarkdownFormatting(t *testing.T) {
	timestamp := time.Date(2026, 1, 30, 10, 0, 0, 0, time.UTC)

	entries := []history.Entry{
		{
			Version:   "1.1.0",
			Package:   "core",
			Timestamp: timestamp,
			Consignments: []history.Consignment{
				{ID: "c1", Summary: "Add feature", ChangeType: "minor"},
			},
		},
	}

	t.Run("produces valid markdown", func(t *testing.T) {
		// Test: Render release notes
		output, err := RenderReleaseNotes(entries)

		// Verify: Valid markdown structure
		require.NoError(t, err)

		// Check for proper markdown headers
		assert.True(t, strings.HasPrefix(output, "#"), "Should start with markdown header")
		assert.Contains(t, output, "##", "Should have second-level headers")

		// Check for proper list formatting
		assert.Contains(t, output, "- ", "Should have bullet points")
	})

	t.Run("preserves special markdown characters", func(t *testing.T) {
		// Setup: Create entry with special characters
		specialEntries := []history.Entry{
			{
				Version:   "1.0.0",
				Package:   "core",
				Timestamp: timestamp,
				Consignments: []history.Consignment{
					{ID: "c1", Summary: "Fix issue with `code` and *emphasis*", ChangeType: "patch"},
				},
			},
		}

		// Test: Render release notes
		output, err := RenderReleaseNotes(specialEntries)

		// Verify: Special characters preserved (markdown supports them in text)
		require.NoError(t, err)
		assert.Contains(t, output, "Fix issue with `code` and *emphasis*")
	})
}
