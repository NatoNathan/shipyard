package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPromptConfirm_Yes tests confirming with yes
func TestPromptConfirm_Yes(t *testing.T) {
	// Mock: Simulate user confirming
	mockInput := func() (bool, error) {
		return true, nil
	}

	// Test: Prompt for confirmation
	result, err := PromptConfirmFunc("Proceed?", false, mockInput)

	// Verify: Should return true
	require.NoError(t, err)
	assert.True(t, result)
}

// TestPromptConfirm_No tests declining with no
func TestPromptConfirm_No(t *testing.T) {
	// Mock: Simulate user declining
	mockInput := func() (bool, error) {
		return false, nil
	}

	// Test: Prompt for confirmation
	result, err := PromptConfirmFunc("Proceed?", true, mockInput)

	// Verify: Should return false
	require.NoError(t, err)
	assert.False(t, result)
}

// TestPromptConfirm_DefaultYes tests default value of yes
func TestPromptConfirm_DefaultYes(t *testing.T) {
	// Mock: Simulate user just pressing enter (accepting default)
	mockInput := func() (bool, error) {
		return true, nil // Default was yes
	}

	// Test: Prompt with default yes
	result, err := PromptConfirmFunc("Continue?", true, mockInput)

	// Verify: Should return true (the default)
	require.NoError(t, err)
	assert.True(t, result)
}

// TestPromptConfirm_DefaultNo tests default value of no
func TestPromptConfirm_DefaultNo(t *testing.T) {
	// Mock: Simulate user just pressing enter (accepting default)
	mockInput := func() (bool, error) {
		return false, nil // Default was no
	}

	// Test: Prompt with default no
	result, err := PromptConfirmFunc("Overwrite?", false, mockInput)

	// Verify: Should return false (the default)
	require.NoError(t, err)
	assert.False(t, result)
}
