package consignment

import (
	"strings"
	"testing"
	"time"

	"github.com/NatoNathan/shipyard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSerialize_BasicConsignment tests serializing a basic consignment
func TestSerialize_BasicConsignment(t *testing.T) {
	cons := &Consignment{
		ID:         "20260130-143022-a1b2c3",
		Timestamp:  time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC),
		Packages:   []string{"core"},
		ChangeType: types.ChangeTypePatch,
		Summary:    "Fixed a bug",
		Metadata:   nil,
	}

	content, err := Serialize(cons)
	require.NoError(t, err, "Should serialize without error")

	// Verify frontmatter is present
	assert.Contains(t, content, "---\n", "Should contain frontmatter delimiters")
	assert.Contains(t, content, "id: 20260130-143022-a1b2c3", "Should contain ID")
	assert.Contains(t, content, "timestamp:", "Should contain timestamp field")
	assert.Contains(t, content, "2026-01-30T14:30:22Z", "Should contain timestamp value")
	assert.Contains(t, content, "packages:", "Should contain packages")
	assert.Contains(t, content, "- core", "Should contain package name")
	assert.Contains(t, content, "changeType: patch", "Should contain change type")

	// Verify markdown body
	assert.Contains(t, content, "Fixed a bug", "Should contain summary")
}

// TestSerialize_WithMetadata tests serializing a consignment with metadata
func TestSerialize_WithMetadata(t *testing.T) {
	cons := &Consignment{
		ID:         "20260130-143022-a1b2c3",
		Timestamp:  time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC),
		Packages:   []string{"core", "api"},
		ChangeType: types.ChangeTypeMinor,
		Summary:    "Added new feature",
		Metadata: map[string]interface{}{
			"author": "dev@example.com",
			"issue":  "JIRA-123",
		},
	}

	content, err := Serialize(cons)
	require.NoError(t, err, "Should serialize without error")

	// Verify metadata
	assert.Contains(t, content, "metadata:", "Should contain metadata")
	assert.Contains(t, content, "author: dev@example.com", "Should contain author metadata")
	assert.Contains(t, content, "issue: JIRA-123", "Should contain issue metadata")
}

// TestSerialize_MultiplePackages tests serializing with multiple packages
func TestSerialize_MultiplePackages(t *testing.T) {
	cons := &Consignment{
		ID:         "20260130-143022-a1b2c3",
		Timestamp:  time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC),
		Packages:   []string{"core", "api", "web"},
		ChangeType: types.ChangeTypeMajor,
		Summary:    "Breaking change",
		Metadata:   nil,
	}

	content, err := Serialize(cons)
	require.NoError(t, err, "Should serialize without error")

	// Verify all packages are present
	assert.Contains(t, content, "- core", "Should contain core package")
	assert.Contains(t, content, "- api", "Should contain api package")
	assert.Contains(t, content, "- web", "Should contain web package")
}

// TestSerialize_MultilineSummary tests serializing with multiline markdown
func TestSerialize_MultilineSummary(t *testing.T) {
	summary := `# Fixed Authentication Bug

The authentication module was failing when users had empty profiles.

## Details

- Added null check
- Added unit tests
- Verified in staging`

	cons := &Consignment{
		ID:         "20260130-143022-a1b2c3",
		Timestamp:  time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC),
		Packages:   []string{"core"},
		ChangeType: types.ChangeTypePatch,
		Summary:    summary,
		Metadata:   nil,
	}

	content, err := Serialize(cons)
	require.NoError(t, err, "Should serialize without error")

	// Verify markdown structure is preserved
	assert.Contains(t, content, "# Fixed Authentication Bug", "Should preserve heading")
	assert.Contains(t, content, "## Details", "Should preserve subheading")
	assert.Contains(t, content, "- Added null check", "Should preserve list items")
}

// TestSerialize_FrontmatterDelimiters tests proper frontmatter formatting
func TestSerialize_FrontmatterDelimiters(t *testing.T) {
	cons := &Consignment{
		ID:         "20260130-143022-a1b2c3",
		Timestamp:  time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC),
		Packages:   []string{"core"},
		ChangeType: types.ChangeTypePatch,
		Summary:    "Test",
		Metadata:   nil,
	}

	content, err := Serialize(cons)
	require.NoError(t, err, "Should serialize without error")

	// Verify frontmatter delimiters
	lines := strings.Split(content, "\n")
	assert.Equal(t, "---", lines[0], "Should start with ---")

	// Find closing delimiter
	foundClosing := false
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			foundClosing = true
			break
		}
	}
	assert.True(t, foundClosing, "Should have closing --- delimiter")
}

// TestSerialize_EmptyMetadata tests serialization with nil metadata
func TestSerialize_EmptyMetadata(t *testing.T) {
	cons := &Consignment{
		ID:         "20260130-143022-a1b2c3",
		Timestamp:  time.Date(2026, 1, 30, 14, 30, 22, 0, time.UTC),
		Packages:   []string{"core"},
		ChangeType: types.ChangeTypePatch,
		Summary:    "Test",
		Metadata:   nil,
	}

	content, err := Serialize(cons)
	require.NoError(t, err, "Should serialize without error")

	// Metadata should not appear if nil
	assert.NotContains(t, content, "metadata:", "Should not contain empty metadata field")
}
