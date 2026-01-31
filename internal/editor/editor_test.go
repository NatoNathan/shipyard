package editor

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenEditor_WithContent tests opening editor with initial content
func TestOpenEditor_WithContent(t *testing.T) {
	// Setup: Create temp file with initial content
	tempDir := t.TempDir()
	initialContent := "# Initial content\n\nEdit this file"

	// Mock: Simulate editor execution (just append text)
	mockEditor := func(path string) error {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		modified := append(content, []byte("\n\nUser added this")...)
		return os.WriteFile(path, modified, 0644)
	}

	// Test: Open editor
	result, err := openEditorWithFunc(tempDir, initialContent, mockEditor)

	// Verify: Should return modified content
	require.NoError(t, err)
	assert.Contains(t, result, "Initial content")
	assert.Contains(t, result, "User added this")
}

// TestOpenEditor_EmptyContent tests opening editor with no initial content
func TestOpenEditor_EmptyContent(t *testing.T) {
	// Setup: No initial content
	tempDir := t.TempDir()

	// Mock: Simulate editor execution
	mockEditor := func(path string) error {
		return os.WriteFile(path, []byte("New content"), 0644)
	}

	// Test: Open editor
	result, err := openEditorWithFunc(tempDir, "", mockEditor)

	// Verify: Should return new content
	require.NoError(t, err)
	assert.Equal(t, "New content", result)
}

// TestOpenEditor_NoChanges tests when user makes no changes
func TestOpenEditor_NoChanges(t *testing.T) {
	// Setup: Initial content
	tempDir := t.TempDir()
	initialContent := "Unchanged content"

	// Mock: Simulate editor execution (no changes)
	mockEditor := func(path string) error {
		return nil // No changes
	}

	// Test: Open editor
	result, err := openEditorWithFunc(tempDir, initialContent, mockEditor)

	// Verify: Should return original content
	require.NoError(t, err)
	assert.Equal(t, initialContent, result)
}

// TestOpenEditor_EditorFails tests when editor command fails
func TestOpenEditor_EditorFails(t *testing.T) {
	// Setup: Initial content
	tempDir := t.TempDir()
	initialContent := "Content"

	// Mock: Simulate editor failure
	mockEditor := func(path string) error {
		return os.ErrPermission
	}

	// Test: Open editor
	result, err := openEditorWithFunc(tempDir, initialContent, mockEditor)

	// Verify: Should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open editor")
	assert.Empty(t, result)
}

// TestOpenEditor_TempFileCreation tests temp file is created correctly
func TestOpenEditor_TempFileCreation(t *testing.T) {
	// Setup: Initial content with special characters
	tempDir := t.TempDir()
	initialContent := "Line 1\nLine 2\n\nLine 4 with special chars: !@#$%"

	// Mock: Verify temp file has correct content
	var capturedPath string
	mockEditor := func(path string) error {
		capturedPath = path
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		// Verify initial content is correct
		assert.Equal(t, initialContent, string(content))
		return nil
	}

	// Test: Open editor
	_, err := openEditorWithFunc(tempDir, initialContent, mockEditor)

	// Verify: Temp file was created and cleaned up
	require.NoError(t, err)
	assert.NotEmpty(t, capturedPath)
	assert.Contains(t, capturedPath, tempDir)
	assert.Contains(t, capturedPath, ".md")

	// Verify temp file was cleaned up
	_, err = os.Stat(capturedPath)
	assert.True(t, os.IsNotExist(err), "Temp file should be cleaned up")
}

// TestOpenEditor_PreservesFormatting tests multiline content preservation
func TestOpenEditor_PreservesFormatting(t *testing.T) {
	// Setup: Multiline content with formatting
	tempDir := t.TempDir()
	initialContent := `---
title: Test
---

# Header

- List item 1
- List item 2

Code block:
` + "```" + `
code here
` + "```"

	// Mock: Simulate editor execution (no changes)
	mockEditor := func(path string) error {
		return nil
	}

	// Test: Open editor
	result, err := openEditorWithFunc(tempDir, initialContent, mockEditor)

	// Verify: Formatting preserved
	require.NoError(t, err)
	assert.Equal(t, initialContent, result)
}

// openEditorWithFunc is a helper that allows testing with mocked editor command
func openEditorWithFunc(dir, initialContent string, editorFunc func(string) error) (string, error) {
	return OpenEditorWithFunc(dir, initialContent, editorFunc)
}
