package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles for different message types
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)

	// Key-value styles
	keyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))

	// Section header style
	sectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("13")).
			Bold(true).
			Underline(true).
			MarginTop(1).
			MarginBottom(1)

	// List bullet style
	bulletStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

	// Dimmed text style for secondary info
	dimmedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	// Header style - bold magenta without underline/margins (for command output headers)
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("13")).
			Bold(true)

	// Change type badge styles
	majorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	minorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	patchStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)

	// Version arrow styles
	oldVersionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	newVersionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	arrowStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
)

// SuccessMessage returns a styled success message with check mark
func SuccessMessage(message string) string {
	return successStyle.Render("✓ " + message)
}

// ErrorMessage returns a styled error message with X mark
func ErrorMessage(message string) string {
	return errorStyle.Render("✗ " + message)
}

// InfoMessage returns a styled info message with info symbol
func InfoMessage(message string) string {
	return infoStyle.Render("ℹ " + message)
}

// WarningMessage returns a styled warning message with warning symbol
func WarningMessage(message string) string {
	return warningStyle.Render("⚠ " + message)
}

// KeyValue returns a styled key-value pair
func KeyValue(key, value string) string {
	return fmt.Sprintf("%s: %s", keyStyle.Render(key), valueStyle.Render(value))
}

// Section returns a styled section header
func Section(title string) string {
	return sectionStyle.Render(title)
}

// List returns a styled bulleted list
func List(items []string) string {
	var lines []string
	for _, item := range items {
		lines = append(lines, bulletStyle.Render("  • ")+item)
	}
	return strings.Join(lines, "\n")
}

// Header returns a bold magenta header with an icon prefix.
// Unlike Section(), it has no underline or margins, making it suitable for command output headers.
func Header(icon, title string) string {
	return headerStyle.Render(icon + " " + title)
}

// Dimmed returns gray text for secondary or skipped information.
func Dimmed(text string) string {
	return dimmedStyle.Render(text)
}

// ChangeTypeBadge returns a colorized change type string:
// major=red, minor=yellow, patch=green.
func ChangeTypeBadge(changeType string) string {
	switch changeType {
	case "major":
		return majorStyle.Render(changeType)
	case "minor":
		return minorStyle.Render(changeType)
	case "patch":
		return patchStyle.Render(changeType)
	default:
		return changeType
	}
}

// VersionArrow returns a styled "old -> new" string with red old, green new, and cyan arrow.
func VersionArrow(old, newVer string) string {
	return oldVersionStyle.Render(old) + arrowStyle.Render(" \u2192 ") + newVersionStyle.Render(newVer)
}

// NewSpinner creates a new progress spinner with the given message
func NewSpinner(message string) *spinner.Spinner {
	return spinner.New().Title(message)
}
