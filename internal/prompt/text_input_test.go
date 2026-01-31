package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPromptTextInput_WithValue tests entering a value
func TestPromptTextInput_WithValue(t *testing.T) {
	// Mock: Simulate user entering a value
	mockInput := func() (string, error) {
		return "my-package", nil
	}

	// Test: Prompt for text input
	result, err := PromptTextInputFunc("Package name:", "default", mockInput)

	// Verify: Should return entered value
	require.NoError(t, err)
	assert.Equal(t, "my-package", result)
}

// TestPromptTextInput_UseDefault tests accepting default
func TestPromptTextInput_UseDefault(t *testing.T) {
	// Mock: Simulate user accepting default (empty input)
	mockInput := func() (string, error) {
		return "", nil
	}

	// Test: Prompt with default
	result, err := PromptTextInputFunc("Package name:", "default-pkg", mockInput)

	// Verify: Should return default value
	require.NoError(t, err)
	assert.Equal(t, "default-pkg", result)
}

// TestPromptTextInput_EmptyDefault tests with no default
func TestPromptTextInput_EmptyDefault(t *testing.T) {
	// Mock: Simulate user entering a value
	mockInput := func() (string, error) {
		return "custom", nil
	}

	// Test: Prompt with no default
	result, err := PromptTextInputFunc("Enter value:", "", mockInput)

	// Verify: Should return entered value
	require.NoError(t, err)
	assert.Equal(t, "custom", result)
}
