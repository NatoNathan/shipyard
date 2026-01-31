package ui

import (
	"testing"

	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/stretchr/testify/assert"
)

// TestRenderPreview tests rendering version preview
func TestRenderPreview(t *testing.T) {
	changes := []PackageChange{
		{
			Name:       "core",
			OldVersion: semver.MustParse("1.0.0"),
			NewVersion: semver.MustParse("1.1.0"),
			ChangeType: "minor",
			Changes:    []string{"Add new feature", "Fix bug"},
		},
		{
			Name:       "api",
			OldVersion: semver.MustParse("2.0.0"),
			NewVersion: semver.MustParse("2.0.1"),
			ChangeType: "patch",
			Changes:    []string{"Update dependency"},
		},
	}

	output := RenderPreview(changes)

	// Should contain package names
	assert.Contains(t, output, "core", "Should show core package")
	assert.Contains(t, output, "api", "Should show api package")

	// Should contain versions
	assert.Contains(t, output, "1.0.0", "Should show old version")
	assert.Contains(t, output, "1.1.0", "Should show new version")
	assert.Contains(t, output, "2.0.0", "Should show old version")
	assert.Contains(t, output, "2.0.1", "Should show new version")

	// Should contain change types
	assert.Contains(t, output, "minor", "Should show change type")
	assert.Contains(t, output, "patch", "Should show change type")

	// Should contain changes
	assert.Contains(t, output, "Add new feature", "Should list changes")
	assert.Contains(t, output, "Fix bug", "Should list changes")
	assert.Contains(t, output, "Update dependency", "Should list changes")
}

// TestRenderPreviewEmpty tests rendering with no changes
func TestRenderPreviewEmpty(t *testing.T) {
	changes := []PackageChange{}

	output := RenderPreview(changes)

	// Should indicate no changes
	assert.Contains(t, output, "No changes", "Should indicate no changes")
}

// TestRenderVersionDiff tests rendering version diff
func TestRenderVersionDiff(t *testing.T) {
	oldVer := semver.MustParse("1.2.3")
	newVer := semver.MustParse("2.0.0")

	output := RenderVersionDiff(oldVer, newVer)

	// Should show both versions
	assert.Contains(t, output, "1.2.3", "Should show old version")
	assert.Contains(t, output, "2.0.0", "Should show new version")

	// Should have some indicator of direction/change
	assert.True(t, len(output) > len("1.2.3 2.0.0"), "Should have formatting beyond just versions")
}
