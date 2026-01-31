package prompt

import (
	"testing"

	"github.com/NatoNathan/shipyard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPromptForChangeType_Patch tests selecting patch
func TestPromptForChangeType_Patch(t *testing.T) {
	// Mock: Simulate user selecting "patch"
	mockInput := func() (types.ChangeType, error) {
		return types.ChangeTypePatch, nil
	}

	// Test: Prompt for change type
	selected, err := promptForChangeTypeWithInput(mockInput)

	// Verify: Should return patch
	require.NoError(t, err)
	assert.Equal(t, types.ChangeTypePatch, selected)
}

// TestPromptForChangeType_Minor tests selecting minor
func TestPromptForChangeType_Minor(t *testing.T) {
	// Mock: Simulate user selecting "minor"
	mockInput := func() (types.ChangeType, error) {
		return types.ChangeTypeMinor, nil
	}

	// Test: Prompt for change type
	selected, err := promptForChangeTypeWithInput(mockInput)

	// Verify: Should return minor
	require.NoError(t, err)
	assert.Equal(t, types.ChangeTypeMinor, selected)
}

// TestPromptForChangeType_Major tests selecting major
func TestPromptForChangeType_Major(t *testing.T) {
	// Mock: Simulate user selecting "major"
	mockInput := func() (types.ChangeType, error) {
		return types.ChangeTypeMajor, nil
	}

	// Test: Prompt for change type
	selected, err := promptForChangeTypeWithInput(mockInput)

	// Verify: Should return major
	require.NoError(t, err)
	assert.Equal(t, types.ChangeTypeMajor, selected)
}

// TestPromptForChangeType_Empty tests empty selection
func TestPromptForChangeType_Empty(t *testing.T) {
	// Mock: Simulate empty selection
	mockInput := func() (types.ChangeType, error) {
		return "", nil
	}

	// Test: Prompt for change type
	selected, err := promptForChangeTypeWithInput(mockInput)

	// Verify: Should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must select a change type")
	assert.Empty(t, selected)
}

// promptForChangeTypeWithInput is a helper that allows testing with mocked input
func promptForChangeTypeWithInput(inputFunc func() (types.ChangeType, error)) (types.ChangeType, error) {
	return PromptForChangeTypeFunc(inputFunc)
}
