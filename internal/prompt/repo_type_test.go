package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPromptRepoType_Single tests selecting single repository
func TestPromptRepoType_Single(t *testing.T) {
	// Mock: Simulate user selecting single repo
	mockInput := func() (RepoType, error) {
		return RepoTypeSingle, nil
	}

	// Test: Prompt for repo type
	result, err := PromptRepoTypeFunc(mockInput)

	// Verify: Should return single
	require.NoError(t, err)
	assert.Equal(t, RepoTypeSingle, result)
}

// TestPromptRepoType_Monorepo tests selecting monorepo
func TestPromptRepoType_Monorepo(t *testing.T) {
	// Mock: Simulate user selecting monorepo
	mockInput := func() (RepoType, error) {
		return RepoTypeMonorepo, nil
	}

	// Test: Prompt for repo type
	result, err := PromptRepoTypeFunc(mockInput)

	// Verify: Should return monorepo
	require.NoError(t, err)
	assert.Equal(t, RepoTypeMonorepo, result)
}
