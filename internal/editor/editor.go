package editor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

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

		command, editorArgs, err := resolveEditorCommand(editor)
		if err != nil {
			return "", err
		}

		args := append(editorArgs, tempPath)
		cmd, err := newEditorCommand(command, args)
		if err != nil {
			return "", err
		}
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

func resolveEditorCommand(editor string) (string, []string, error) {
	editorParts, err := splitEditorCommand(editor)
	if err != nil {
		return "", nil, err
	}
	if len(editorParts) == 0 {
		return "", nil, fmt.Errorf("EDITOR is empty")
	}

	editorPath := editorParts[0]
	command, args, ok := allowedEditor(editorBaseName(editorPath), editorParts[1:])
	if !ok {
		return "", nil, fmt.Errorf("unsupported editor %q; set EDITOR to one of: code, code-insiders, emacs, nano, nvim, vi, vim", editorPath)
	}
	if hasPathSeparator(editorPath) {
		command = editorPath
	}
	if err := verifyEditorExecutable(command); err != nil {
		return "", nil, fmt.Errorf("editor executable not found: %w", err)
	}

	return command, args, nil
}

func splitEditorCommand(editor string) ([]string, error) {
	var parts []string
	var current strings.Builder
	var quote rune
	escaped := false
	inToken := false

	for _, r := range editor {
		if escaped {
			current.WriteRune(r)
			escaped = false
			inToken = true
			continue
		}
		if quote != 0 {
			if r == '\\' && quote == '"' {
				escaped = true
				continue
			}
			if r == quote {
				quote = 0
				inToken = true
				continue
			}
			current.WriteRune(r)
			inToken = true
			continue
		}
		if r == '\'' || r == '"' {
			quote = r
			inToken = true
			continue
		}
		if unicode.IsSpace(r) {
			if inToken {
				parts = append(parts, current.String())
				current.Reset()
				inToken = false
			}
			continue
		}
		current.WriteRune(r)
		inToken = true
	}
	if escaped || quote != 0 {
		return nil, fmt.Errorf("invalid EDITOR value: unterminated quote or escape")
	}
	if inToken {
		parts = append(parts, current.String())
	}
	return parts, nil
}

func editorBaseName(command string) string {
	name := filepath.Base(strings.ReplaceAll(command, "\\", string(filepath.Separator)))
	name = strings.ToLower(name)
	return strings.TrimSuffix(name, ".exe")
}

func hasPathSeparator(command string) bool {
	return filepath.IsAbs(command) || strings.ContainsAny(command, `/\`)
}

func verifyEditorExecutable(command string) error {
	if hasPathSeparator(command) {
		info, err := os.Stat(command)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return fmt.Errorf("%s is a directory", command)
		}
		return nil
	}
	_, err := exec.LookPath(command)
	return err
}

func allowedEditor(name string, requestedArgs []string) (string, []string, bool) {
	switch name {
	case "code":
		return "code", allowedCodeArgs(requestedArgs), true
	case "code-insiders":
		return "code-insiders", allowedCodeArgs(requestedArgs), true
	case "emacs":
		return "emacs", nil, true
	case "nano":
		return "nano", nil, true
	case "nvim":
		return "nvim", nil, true
	case "vi":
		return "vi", nil, true
	case "vim":
		return "vim", nil, true
	default:
		return "", nil, false
	}
}

func allowedCodeArgs(requestedArgs []string) []string {
	args := make([]string, 0, len(requestedArgs))
	for _, arg := range requestedArgs {
		switch arg {
		case "--wait":
			args = append(args, "--wait")
		case "--new-window":
			args = append(args, "--new-window")
		case "--reuse-window":
			args = append(args, "--reuse-window")
		}
	}
	return args
}

func newEditorCommand(command string, args []string) (*exec.Cmd, error) {
	switch editorBaseName(command) {
	case "code":
		return exec.Command(command, args...), nil // #nosec G204,G702 -- executable basename is allowlisted and args are sanitized flags plus a generated temp file path.
	case "code-insiders":
		return exec.Command(command, args...), nil // #nosec G204,G702 -- executable basename is allowlisted and args are sanitized flags plus a generated temp file path.
	case "emacs":
		return exec.Command(command, args...), nil // #nosec G204,G702 -- executable basename is allowlisted and args are sanitized flags plus a generated temp file path.
	case "nano":
		return exec.Command(command, args...), nil // #nosec G204,G702 -- executable basename is allowlisted and args are sanitized flags plus a generated temp file path.
	case "nvim":
		return exec.Command(command, args...), nil // #nosec G204,G702 -- executable basename is allowlisted and args are sanitized flags plus a generated temp file path.
	case "vi":
		return exec.Command(command, args...), nil // #nosec G204,G702 -- executable basename is allowlisted and args are sanitized flags plus a generated temp file path.
	case "vim":
		return exec.Command(command, args...), nil // #nosec G204,G702 -- executable basename is allowlisted and args are sanitized flags plus a generated temp file path.
	default:
		return nil, fmt.Errorf("unsupported editor %q", command)
	}
}
