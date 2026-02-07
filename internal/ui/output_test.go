package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSuccessMessage tests rendering success messages
func TestSuccessMessage(t *testing.T) {
	msg := SuccessMessage("Operation completed")

	// Should contain the success indicator and message
	assert.Contains(t, msg, "✓", "Success message should contain check mark")
	assert.Contains(t, msg, "Operation completed", "Success message should contain the text")
}

// TestErrorMessage tests rendering error messages
func TestErrorMessage(t *testing.T) {
	msg := ErrorMessage("Operation failed")

	// Should contain the error indicator and message
	assert.Contains(t, msg, "✗", "Error message should contain X mark")
	assert.Contains(t, msg, "Operation failed", "Error message should contain the text")
}

// TestInfoMessage tests rendering info messages
func TestInfoMessage(t *testing.T) {
	msg := InfoMessage("Processing...")

	// Should contain the info indicator and message
	assert.Contains(t, msg, "ℹ", "Info message should contain info symbol")
	assert.Contains(t, msg, "Processing...", "Info message should contain the text")
}

// TestWarningMessage tests rendering warning messages
func TestWarningMessage(t *testing.T) {
	msg := WarningMessage("Be careful")

	// Should contain the warning indicator and message
	assert.Contains(t, msg, "⚠", "Warning message should contain warning symbol")
	assert.Contains(t, msg, "Be careful", "Warning message should contain the text")
}

// TestKeyValue tests rendering key-value pairs
func TestKeyValue(t *testing.T) {
	msg := KeyValue("Package", "core")

	// Should contain both key and value
	assert.Contains(t, msg, "Package", "Should contain the key")
	assert.Contains(t, msg, "core", "Should contain the value")

	// Key and value should be on same line
	lines := strings.Split(msg, "\n")
	assert.Equal(t, 1, len(lines), "Key-value should be on single line")
}

// TestSection tests rendering section headers
func TestSection(t *testing.T) {
	msg := Section("Configuration")

	// Should contain the section title
	assert.Contains(t, msg, "Configuration", "Should contain section title")

	// Should have some styling to make it stand out
	assert.NotEqual(t, "Configuration", msg, "Section should be styled differently than plain text")
}

// TestList tests rendering bulleted lists
func TestList(t *testing.T) {
	items := []string{"Item 1", "Item 2", "Item 3"}
	msg := List(items)

	// Should contain all items
	for _, item := range items {
		assert.Contains(t, msg, item, "List should contain all items")
	}

	// Should have bullet points
	assert.Contains(t, msg, "•", "List should have bullet points")
}

// TestHeader tests rendering header with icon
func TestHeader(t *testing.T) {
	msg := Header("\U0001F4E6", "Pending Changes")

	assert.Contains(t, msg, "\U0001F4E6", "Header should contain icon")
	assert.Contains(t, msg, "Pending Changes", "Header should contain title")
}

// TestDimmed tests rendering dimmed text
func TestDimmed(t *testing.T) {
	msg := Dimmed("Skipped git commit")

	assert.Contains(t, msg, "Skipped git commit", "Dimmed should contain text")
}

// TestChangeTypeBadge tests rendering change type badges
func TestChangeTypeBadge(t *testing.T) {
	tests := []struct {
		changeType string
	}{
		{"major"},
		{"minor"},
		{"patch"},
	}

	for _, tt := range tests {
		t.Run(tt.changeType, func(t *testing.T) {
			badge := ChangeTypeBadge(tt.changeType)
			assert.Contains(t, badge, tt.changeType, "Badge should contain change type")
		})
	}
}

// TestChangeTypeBadge_Unknown tests that unknown change types are returned unstyled
func TestChangeTypeBadge_Unknown(t *testing.T) {
	badge := ChangeTypeBadge("custom")
	assert.Equal(t, "custom", badge, "Unknown change types should be returned as-is")
}

// TestVersionArrow tests rendering version arrows
func TestVersionArrow(t *testing.T) {
	msg := VersionArrow("1.0.0", "1.1.0")

	assert.Contains(t, msg, "1.0.0", "Should contain old version")
	assert.Contains(t, msg, "1.1.0", "Should contain new version")
	assert.Contains(t, msg, "\u2192", "Should contain arrow")
}

// TestProgressSpinner tests creating a progress spinner
func TestProgressSpinner(t *testing.T) {
	// Test that we can create a spinner without errors
	spinner := NewSpinner("Loading...")
	assert.NotNil(t, spinner, "Should create spinner")
}
