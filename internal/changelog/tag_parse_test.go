package changelog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseTagOutput_SingleLine tests parsing lightweight tag (single line)
func TestParseTagOutput_SingleLine(t *testing.T) {
	output := "v1.0.0"

	name, message, err := ParseTagOutput(output)

	require.NoError(t, err)
	assert.Equal(t, "v1.0.0", name)
	assert.Equal(t, "", message) // Empty message = lightweight tag
}

// TestParseTagOutput_SingleLineWithWhitespace tests trimming whitespace
func TestParseTagOutput_SingleLineWithWhitespace(t *testing.T) {
	output := "  v1.0.0  \n"

	name, message, err := ParseTagOutput(output)

	require.NoError(t, err)
	assert.Equal(t, "v1.0.0", name)
	assert.Equal(t, "", message)
}

// TestParseTagOutput_AnnotatedTag tests parsing annotated tag with message
func TestParseTagOutput_AnnotatedTag(t *testing.T) {
	output := `core/v1.2.0

# Release core v1.2.0

## Changes

- Added OAuth2 support
- Fixed validation bug`

	name, message, err := ParseTagOutput(output)

	require.NoError(t, err)
	assert.Equal(t, "core/v1.2.0", name)
	assert.Equal(t, `# Release core v1.2.0

## Changes

- Added OAuth2 support
- Fixed validation bug`, message)
}

// TestParseTagOutput_AnnotatedTagWithoutBlankLine tests error when missing blank line
func TestParseTagOutput_AnnotatedTagWithoutBlankLine(t *testing.T) {
	output := `core/v1.2.0
# Release core v1.2.0`

	_, _, err := ParseTagOutput(output)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "blank line")
}

// TestParseTagOutput_EmptyOutput tests error handling for empty output
func TestParseTagOutput_EmptyOutput(t *testing.T) {
	output := ""

	_, _, err := ParseTagOutput(output)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

// TestParseTagOutput_OnlyWhitespace tests error handling for whitespace-only output
func TestParseTagOutput_OnlyWhitespace(t *testing.T) {
	output := "   \n\n  "

	_, _, err := ParseTagOutput(output)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

// TestParseTagOutput_MultiLineNameOnly tests two lines without blank separator
func TestParseTagOutput_MultiLineNameOnly(t *testing.T) {
	output := `v1.0.0
v2.0.0`

	_, _, err := ParseTagOutput(output)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "blank line")
}
