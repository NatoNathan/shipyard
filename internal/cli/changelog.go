package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/glamour"
)

// renderMarkdown renders markdown content using Glamour with a nice style
func renderMarkdown(content string) (string, error) {
	// Create a glamour renderer with a nice style
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create markdown renderer: %w", err)
	}

	// Render the markdown
	rendered, err := r.Render(content)
	if err != nil {
		return "", fmt.Errorf("failed to render markdown: %w", err)
	}

	return rendered, nil
}

// writeChangelog writes the changelog content to a file
func writeChangelog(outputPath, content string) error {
	// Ensure the directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write changelog file: %w", err)
	}

	return nil
}
