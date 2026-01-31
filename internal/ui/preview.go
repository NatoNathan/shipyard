package ui

import (
	"fmt"
	"strings"

	"github.com/NatoNathan/shipyard/pkg/semver"
	"github.com/charmbracelet/lipgloss"
)

// PackageChange represents a version change for preview
type PackageChange struct {
	Name       string
	OldVersion semver.Version
	NewVersion semver.Version
	ChangeType string
	Changes    []string
}

var (
	// Preview styles
	packageNameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("13")).
				Bold(true)

	versionOldStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9"))

	versionNewStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true)

	changeTypeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Italic(true)

	changeItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")).
			MarginLeft(4)
)

// RenderPreview renders a preview of version changes
func RenderPreview(changes []PackageChange) string {
	if len(changes) == 0 {
		return InfoMessage("No changes to preview")
	}

	var sections []string

	sections = append(sections, Section("Version Preview"))

	for _, change := range changes {
		pkgSection := renderPackageChange(change)
		sections = append(sections, pkgSection)
	}

	return strings.Join(sections, "\n\n")
}

// renderPackageChange renders a single package change
func renderPackageChange(change PackageChange) string {
	var lines []string

	// Package name and version diff
	pkgName := packageNameStyle.Render(change.Name)
	versionDiff := RenderVersionDiff(change.OldVersion, change.NewVersion)
	changeType := changeTypeStyle.Render(fmt.Sprintf("(%s)", change.ChangeType))

	lines = append(lines, fmt.Sprintf("%s: %s %s", pkgName, versionDiff, changeType))

	// Changes list
	if len(change.Changes) > 0 {
		lines = append(lines, "  Changes:")
		for _, item := range change.Changes {
			lines = append(lines, changeItemStyle.Render("• "+item))
		}
	}

	return strings.Join(lines, "\n")
}

// RenderVersionDiff renders a version diff with arrow
func RenderVersionDiff(oldVer, newVer semver.Version) string {
	old := versionOldStyle.Render(oldVer.String())
	new := versionNewStyle.Render(newVer.String())
	arrow := lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Render("→")

	return fmt.Sprintf("%s %s %s", old, arrow, new)
}
