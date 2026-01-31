package template

import (
	"fmt"
	"strings"

	"github.com/NatoNathan/shipyard/internal/history"
)

// RenderOptions configures release notes rendering
type RenderOptions struct {
	GroupByType bool // Group changes by type (major, minor, patch)
}

// DefaultRenderOptions returns default rendering options
func DefaultRenderOptions() *RenderOptions {
	return &RenderOptions{
		GroupByType: false,
	}
}

// RenderReleaseNotes renders release notes using the builtin markdown template
func RenderReleaseNotes(entries []history.Entry) (string, error) {
	return RenderReleaseNotesWithOptions(entries, nil)
}

// RenderReleaseNotesWithOptions renders release notes with custom options
func RenderReleaseNotesWithOptions(entries []history.Entry, opts *RenderOptions) (string, error) {
	if opts == nil {
		opts = DefaultRenderOptions()
	}

	if len(entries) == 0 {
		return "No releases found\n", nil
	}

	var sb strings.Builder

	// Main title
	sb.WriteString("# Release Notes\n\n")

	// Render each version
	for _, entry := range entries {
		// Version header
		sb.WriteString(fmt.Sprintf("## %s - %s\n\n", entry.Package, entry.Version))
		sb.WriteString(fmt.Sprintf("Released: %s\n\n", entry.Timestamp.Format("2006-01-02")))

		// Render changes
		if len(entry.Consignments) > 0 {
			if opts.GroupByType {
				renderGroupedChanges(&sb, entry.Consignments)
			} else {
				renderFlatChanges(&sb, entry.Consignments)
			}
		}
	}

	return sb.String(), nil
}

// renderFlatChanges renders changes as a flat list
func renderFlatChanges(sb *strings.Builder, consignments []history.Consignment) {
	sb.WriteString("### Changes\n\n")
	for _, c := range consignments {
		sb.WriteString(fmt.Sprintf("- **%s**: %s\n", c.ChangeType, c.Summary))
	}
	sb.WriteString("\n")
}

// renderGroupedChanges renders changes grouped by type
func renderGroupedChanges(sb *strings.Builder, consignments []history.Consignment) {
	// Group by type
	groups := make(map[string][]history.Consignment)
	for _, c := range consignments {
		groups[c.ChangeType] = append(groups[c.ChangeType], c)
	}

	// Render in order: major, minor, patch
	typeOrder := []struct {
		key   string
		title string
	}{
		{"major", "#### Breaking Changes"},
		{"minor", "#### Features"},
		{"patch", "#### Bug Fixes"},
	}

	for _, typeInfo := range typeOrder {
		if changes, ok := groups[typeInfo.key]; ok && len(changes) > 0 {
			sb.WriteString(fmt.Sprintf("%s\n\n", typeInfo.title))
			for _, c := range changes {
				sb.WriteString(fmt.Sprintf("- %s\n", c.Summary))
			}
			sb.WriteString("\n")
		}
	}

	// Render any other types not in the standard order
	for changeType, changes := range groups {
		if changeType != "major" && changeType != "minor" && changeType != "patch" {
			sb.WriteString(fmt.Sprintf("#### %s\n\n", strings.Title(changeType)))
			for _, c := range changes {
				sb.WriteString(fmt.Sprintf("- %s\n", c.Summary))
			}
			sb.WriteString("\n")
		}
	}
}
