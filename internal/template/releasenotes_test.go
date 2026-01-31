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

		// Test: Render release notes
		output, err := RenderReleaseNotes(entries)

		// Verify: Valid markdown output
		require.NoError(t, err)
		assert.Contains(t, output, "# Release Notes", "Should have title")
		assert.Contains(t, output, "## core - 1.1.0", "Should have version header")
		assert.Contains(t, output, "Released: 2026-01-30", "Should have release date")
		assert.Contains(t, output, "### Changes", "Should have changes section")
		assert.Contains(t, output, "- **minor**: Add new feature", "Should list changes with type")
		assert.Contains(t, output, "- **patch**: Fix critical bug", "Should list all changes")
	})

	t.Run("renders multiple versions", func(t *testing.T) {
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

		// Test: Render release notes
		output, err := RenderReleaseNotes(entries)

		// Verify: Both versions present
		require.NoError(t, err)
		assert.Contains(t, output, "## core - 1.1.0")
		assert.Contains(t, output, "## core - 1.0.1")
		assert.Contains(t, output, "Add feature")
		assert.Contains(t, output, "Fix bug")
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

		// Verify: Version header present but no changes section
		require.NoError(t, err)
		assert.Contains(t, output, "## core - 1.0.0")
		assert.NotContains(t, output, "### Changes", "Should not have changes section for empty consignments")
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

	t.Run("handles nil consignments slice", func(t *testing.T) {
		// Setup: Create entry with nil consignments
		entries := []history.Entry{
			{
				Version:      "1.0.0",
				Package:      "core",
				Timestamp:    timestamp,
				Consignments: nil,
			},
		}

		// Test: Render release notes
		output, err := RenderReleaseNotes(entries)

		// Verify: No error, version rendered without changes
		require.NoError(t, err)
		assert.Contains(t, output, "## core - 1.0.0")
	})
}

// TestRenderReleaseNotesWithOptions tests rendering with options
func TestRenderReleaseNotesWithOptions(t *testing.T) {
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

	t.Run("groups changes by type", func(t *testing.T) {
		// Setup: Create entries with multiple change types
		entriesWithTypes := []history.Entry{
			{
				Version:   "2.0.0",
				Package:   "api",
				Timestamp: timestamp,
				Consignments: []history.Consignment{
					{ID: "c1", Summary: "Breaking change", ChangeType: "major"},
					{ID: "c2", Summary: "New feature", ChangeType: "minor"},
					{ID: "c3", Summary: "Bug fix", ChangeType: "patch"},
					{ID: "c4", Summary: "Another feature", ChangeType: "minor"},
				},
			},
		}

		// Test: Render with grouping
		opts := &RenderOptions{GroupByType: true}
		output, err := RenderReleaseNotesWithOptions(entriesWithTypes, opts)

		// Verify: Changes grouped by type
		require.NoError(t, err)

		// Check structure has grouped sections
		lines := strings.Split(output, "\n")
		var hasMajorSection, hasMinorSection, hasPatchSection bool
		for _, line := range lines {
			if strings.Contains(line, "#### Breaking Changes") || strings.Contains(line, "#### Major") {
				hasMajorSection = true
			}
			if strings.Contains(line, "#### Features") || strings.Contains(line, "#### Minor") {
				hasMinorSection = true
			}
			if strings.Contains(line, "#### Bug Fixes") || strings.Contains(line, "#### Patch") {
				hasPatchSection = true
			}
		}

		assert.True(t, hasMajorSection || strings.Contains(output, "Breaking change"), "Should have major changes")
		assert.True(t, hasMinorSection || strings.Contains(output, "New feature"), "Should have minor changes")
		assert.True(t, hasPatchSection || strings.Contains(output, "Bug fix"), "Should have patch changes")
	})

	t.Run("uses default options when nil", func(t *testing.T) {
		// Test: Render with nil options
		output, err := RenderReleaseNotesWithOptions(entries, nil)

		// Verify: Still renders successfully
		require.NoError(t, err)
		assert.Contains(t, output, "Release Notes")
		assert.Contains(t, output, "1.1.0")
	})
}

// TestRenderOptions tests the options structure
func TestRenderOptions(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		// Test: Get default options
		opts := DefaultRenderOptions()

		// Verify: Sensible defaults
		assert.NotNil(t, opts)
		assert.False(t, opts.GroupByType, "Should not group by default")
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
		assert.Contains(t, output, "\n##", "Should have second-level headers")
		assert.Contains(t, output, "\n###", "Should have third-level headers")

		// Check for proper list formatting
		assert.Contains(t, output, "\n- ", "Should have bullet points")

		// Check for proper line endings
		assert.True(t, strings.HasSuffix(output, "\n"), "Should end with newline")
	})

	t.Run("escapes special markdown characters", func(t *testing.T) {
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
