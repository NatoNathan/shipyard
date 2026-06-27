package template

import (
	"fmt"
	"slices"

	"github.com/NatoNathan/shipyard/internal/history"
)

// RenderReleaseNotes renders release notes using the default template
func RenderReleaseNotes(entries []history.Entry) (string, error) {
	return RenderReleaseNotesWithTemplate(entries, "builtin:default")
}

// RenderReleaseNotesWithTemplate renders a single-version release note.
// Custom templates always receive a single history.Entry as context.
// templateSource can be:
// - "builtin:default" or "builtin:grouped" - builtin templates
// - "file:/path/to/template.tmpl" - file path
// - "https://example.com/template.tmpl" - remote URL
// - inline template content (multiline string)
// - "changelog" - auto-maps to changelog template for multi-version
// - "release-notes" - auto-maps to release-notes template for single-version
func RenderReleaseNotesWithTemplate(entries []history.Entry, templateSource string) (string, error) {
	return renderWithMode(entries, templateSource, TemplateTypeReleaseNotes)
}

// RenderChangelogWithTemplate renders a multi-version changelog.
// Custom templates always receive a ChangelogContext as context.
func RenderChangelogWithTemplate(entries []history.Entry, templateSource string) (string, error) {
	return renderWithMode(entries, templateSource, TemplateTypeChangelog)
}

// renderWithMode is the shared implementation. mode controls the context type used
// for custom (non-alias) template sources.
func renderWithMode(entries []history.Entry, templateSource string, mode TemplateType) (string, error) {
	if len(entries) == 0 {
		return "No releases found\n", nil
	}

	var templateType TemplateType
	var source string

	switch templateSource {
	case "changelog":
		// Named alias always uses all-versions (array) context
		templateType = TemplateTypeChangelog
		source = "builtin:default"
	case "release-notes":
		// Named alias always uses single-version context
		templateType = TemplateTypeReleaseNotes
		source = "builtin:default"
	default:
		// Custom template: honour the caller's stated mode
		templateType = mode
		source = templateSource
	}

	loader := NewTemplateLoader()
	renderer := NewTemplateRenderer()

	templateContent, err := loader.Load(source, templateType)
	if err != nil {
		return "", fmt.Errorf("failed to load template: %w", err)
	}

	var context interface{}
	if templateType == TemplateTypeChangelog {
		sorted := make([]history.Entry, len(entries))
		copy(sorted, entries)
		slices.SortFunc(sorted, func(a, b history.Entry) int {
			return b.Timestamp.Compare(a.Timestamp)
		})
		context = newChangelogContext(sorted)
	} else {
		context = entries[0]
	}

	output, err := renderer.Render(templateContent, context)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return output, nil
}
