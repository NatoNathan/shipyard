package github

import (
	"testing"

	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/stretchr/testify/assert"
)

func TestExtractTitleFromNotes(t *testing.T) {
	version := semver.Version{Major: 1, Minor: 2, Patch: 3}

	tests := []struct {
		name     string
		notes    string
		pkg      string
		expected string
	}{
		{
			name:     "empty string",
			notes:    "",
			pkg:      "myapp",
			expected: "myapp v1.2.3",
		},
		{
			name:     "markdown heading prefix",
			notes:    "# Release Notes\nSome details",
			pkg:      "myapp",
			expected: "myapp v1.2.3",
		},
		{
			name:     "normal title string",
			notes:    "New release with bugfixes\nMore details here",
			pkg:      "myapp",
			expected: "New release with bugfixes",
		},
		{
			name:     "whitespace only first line",
			notes:    "   \nActual content",
			pkg:      "myapp",
			expected: "myapp v1.2.3",
		},
		{
			name:     "multiple heading levels",
			notes:    "## Minor heading\nContent",
			pkg:      "myapp",
			expected: "myapp v1.2.3",
		},
		{
			name:     "single line notes",
			notes:    "Quick bugfix release",
			pkg:      "core",
			expected: "Quick bugfix release",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTitleFromNotes(tt.notes, tt.pkg, version)
			assert.Equal(t, tt.expected, result)
		})
	}
}
