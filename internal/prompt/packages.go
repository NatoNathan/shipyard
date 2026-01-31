package prompt

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

type packageModel struct {
	packages []string
	selected map[int]bool
	cursor   int
	done     bool
	err      error
}

func (m packageModel) Init() tea.Cmd {
	return nil
}

func (m packageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.err = fmt.Errorf("cancelled")
			m.done = true
			return m, tea.Quit

		case "enter":
			if len(m.selected) == 0 {
				m.err = fmt.Errorf("must select at least one package")
			}
			m.done = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.packages)-1 {
				m.cursor++
			}

		case " ":
			if m.selected[m.cursor] {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = true
			}
		}
	}

	return m, nil
}

func (m packageModel) View() string {
	if m.done {
		return ""
	}

	s := titleStyle.Render("Select package(s) affected by this change:") + "\n\n"

	for i, pkg := range m.packages {
		cursor := "  "
		if m.cursor == i {
			cursor = cursorStyle.Render("> ")
		}

		checked := "[ ]"
		if m.selected[i] {
			checked = selectedStyle.Render("[✓]")
		}

		s += fmt.Sprintf("%s%s %s\n", cursor, checked, pkg)
	}

	s += "\n" + helpStyle.Render("space: select • enter: confirm • q: quit")

	return s
}

// PromptForPackages prompts the user to select one or more packages
func PromptForPackages(available []string) ([]string, error) {
	return PromptForPackagesFunc(available, nil)
}

// PromptForPackagesFunc allows dependency injection for testing
func PromptForPackagesFunc(available []string, inputFunc func() ([]string, error)) ([]string, error) {
	// Validate available packages
	if len(available) == 0 {
		return nil, fmt.Errorf("no packages available")
	}

	// If inputFunc provided (for testing), use it
	if inputFunc != nil {
		selected, err := inputFunc()
		if err != nil {
			return nil, err
		}
		if len(selected) == 0 {
			return nil, fmt.Errorf("must select at least one package")
		}
		return selected, nil
	}

	// Interactive prompt using Bubble Tea
	m := packageModel{
		packages: available,
		selected: make(map[int]bool),
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("package selection failed: %w", err)
	}

	result := finalModel.(packageModel)
	if result.err != nil {
		return nil, result.err
	}

	// Build selected list
	var selected []string
	for i := range result.selected {
		selected = append(selected, result.packages[i])
	}

	if len(selected) == 0 {
		return nil, fmt.Errorf("must select at least one package")
	}

	return selected, nil
}
