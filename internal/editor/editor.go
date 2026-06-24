package editor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/NatoNathan/shipyard/internal/fileutil"
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
	defer func() { _ = os.Remove(tempPath) }() // Clean up temp file

	// Write initial content
	if initialContent != "" {
		if _, err := f.WriteString(initialContent); err != nil {
			_ = f.Close()
			return "", fmt.Errorf("failed to write initial content: %w", err)
		}
	}
	_ = f.Close()

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

		editorParts := strings.Fields(editor)
		if len(editorParts) == 0 {
			return "", fmt.Errorf("EDITOR is empty")
		}
		if _, err := exec.LookPath(editorParts[0]); err != nil {
			return "", fmt.Errorf("editor executable not found: %w", err)
		}
		args := append(editorParts[1:], tempPath)
		cmd := exec.Command(editorParts[0], args...) // #nosec G204,G702 -- EDITOR is a user-selected executable; exec.Command does not invoke a shell.
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to open editor: %w", err)
		}
	}

	// Read edited content
	content, err := fileutil.ReadFile(tempPath)
	if err != nil {
		return "", fmt.Errorf("failed to read edited content: %w", err)
	}

	return string(content), nil
}
