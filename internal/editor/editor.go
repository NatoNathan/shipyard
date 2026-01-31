package editor

import (
	"fmt"
	"os"
	"os/exec"
)

// OpenEditor opens a text editor for the user to edit content
func OpenEditor(dir, initialContent string) (string, error) {
	return OpenEditorWithFunc(dir, initialContent, nil)
}

// OpenEditorWithFunc allows dependency injection for testing
func OpenEditorWithFunc(dir, initialContent string, editorFunc func(string) error) (string, error) {
	// Create temp file for editing
	f, err := os.CreateTemp(dir, "shipyard-edit-*.md")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := f.Name()
	defer os.Remove(tempPath) // Clean up temp file

	// Write initial content
	if initialContent != "" {
		if _, err := f.WriteString(initialContent); err != nil {
			f.Close()
			return "", fmt.Errorf("failed to write initial content: %w", err)
		}
	}
	f.Close()

	// Open editor
	if editorFunc != nil {
		// For testing
		if err := editorFunc(tempPath); err != nil {
			return "", fmt.Errorf("failed to open editor: %w", err)
		}
	} else {
		// For real usage - open user's editor
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim" // Default to vim
		}

		cmd := exec.Command(editor, tempPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to open editor: %w", err)
		}
	}

	// Read edited content
	content, err := os.ReadFile(tempPath)
	if err != nil {
		return "", fmt.Errorf("failed to read edited content: %w", err)
	}

	return string(content), nil
}
