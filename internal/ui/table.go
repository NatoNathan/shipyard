package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var (
	// Table border style - cyan to match bulletStyle
	tableBorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

	// Table header style - magenta bold to match sectionStyle
	tableHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("13")).
				Bold(true)

	// Table row style - white to match valueStyle
	tableRowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
)

// Table renders a styled table with the given headers and rows.
// Returns a string that callers can print.
func Table(headers []string, rows [][]string) string {
	headerRow := make([]string, len(headers))
	copy(headerRow, headers)

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(tableBorderStyle).
		Headers(headerRow...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return tableHeaderStyle
			}
			return tableRowStyle
		})

	for _, row := range rows {
		t.Row(row...)
	}

	return t.Render()
}
