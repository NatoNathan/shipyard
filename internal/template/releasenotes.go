package template

import (
	"fmt"

	"github.com/NatoNathan/shipyard/internal/history"
)

// RenderReleaseNotes renders release notes using the default template
func RenderReleaseNotes(entries []history.Entry) (string, error) {
	return RenderReleaseNotesWithTemplate(entries, "builtin:default")
}

// RenderReleaseNotesWithTemplate renders release notes using a specific template
// templateSource can be:
// - "builtin:default" or "builtin:grouped" - builtin templates
// - "file:/path/to/template.tmpl" - file path
// - "https://example.com/template.tmpl" - remote URL
// - inline template content (multiline string)
// - "changelog" - auto-maps to changelog template for multi-version
// - "release-notes" - auto-maps to release-notes template for single-version
func RenderReleaseNotesWithTemplate(entries []history.Entry, templateSource string) (string, error) {
	if len(entries) == 0 {
		return "No releases found\n", nil
	}

	// Handle special auto-selection cases
	var templateType TemplateType
	var source string

	switch templateSource {
	case "changelog":
		// Multi-version changelog (always full history)
		templateType = TemplateTypeChangelog
		source = "builtin:default" // default.tmpl expects array of entries
	case "release-notes":
		// Single-version release notes
		templateType = TemplateTypeReleaseNotes
		source = "builtin:default"
	default:
		// User-specified template - determine type by entry count
		if len(entries) > 1 {
			templateType = TemplateTypeChangelog
		} else {
			templateType = TemplateTypeReleaseNotes
		}
		source = templateSource
	}

	// Create loader and renderer
	loader := NewTemplateLoader()
	renderer := NewTemplateRenderer()

	// Load template
	templateContent, err := loader.Load(source, templateType)
	if err != nil {
		return "", fmt.Errorf("failed to load template: %w", err)
	}

	// Prepare context based on template type
	var context interface{}

	if templateType == TemplateTypeChangelog {
		// Multi-entry context for changelog (array of entries)
		context = entries
	} else {
		// Single-entry context for release notes
		context = entries[0]
	}

	// Render template
	output, err := renderer.Render(templateContent, context)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return output, nil
}
