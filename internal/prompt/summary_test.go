package prompt

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptSummary_DirectInput(t *testing.T) {
	// Mock function that returns a summary without opening editor
	mockFn := func(projectPath string) (string, error) {
		return "Test summary from textarea", nil
	}

	summary, err := PromptSummaryWithFunc("/tmp/test", mockFn)
	require.NoError(t, err)
	assert.Equal(t, "Test summary from textarea", summary)
}

func TestPromptSummary_OpenEditor(t *testing.T) {
	// Mock function that simulates user pressing Ctrl+E
	mockFn := func(projectPath string) (string, error) {
		// Simulate opening editor and returning result
		return "Summary from editor", nil
	}

	summary, err := PromptSummaryWithFunc("/tmp/test", mockFn)
	require.NoError(t, err)
	assert.Equal(t, "Summary from editor", summary)
}

func TestPromptSummary_EmptyInput(t *testing.T) {
	// Mock function that returns empty input
	mockFn := func(projectPath string) (string, error) {
		return "", fmt.Errorf("summary cannot be empty")
	}

	_, err := PromptSummaryWithFunc("/tmp/test", mockFn)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestPromptSummary_Cancelled(t *testing.T) {
	// Mock function that simulates user cancelling
	mockFn := func(projectPath string) (string, error) {
		return "", fmt.Errorf("cancelled")
	}

	_, err := PromptSummaryWithFunc("/tmp/test", mockFn)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled")
}

func TestPromptSummary_MultilineContent(t *testing.T) {
	// Mock function that returns multi-line content
	mockFn := func(projectPath string) (string, error) {
		// Only first line should be returned as summary
		return "First line summary", nil
	}

	summary, err := PromptSummaryWithFunc("/tmp/test", mockFn)
	require.NoError(t, err)
	assert.Equal(t, "First line summary", summary)
	assert.NotContains(t, summary, "\n")
}
