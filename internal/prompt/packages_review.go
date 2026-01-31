package prompt

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/NatoNathan/shipyard/internal/config"
)

type packageReviewModel struct {
	packages []config.Package
	selected map[int]bool
	cursor   int
	done     bool
	err      error
}

func (m packageReviewModel) Init() tea.Cmd {
	return nil
}

func (m packageReviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m packageReviewModel) View() string {
	if m.done {
		return ""
	}

	s := titleStyle.Render("Detected packages - select which to include:") + "\n\n"

	for i, pkg := range m.packages {
		cursor := "  "
		if m.cursor == i {
			cursor = cursorStyle.Render("> ")
		}

		checked := "[ ]"
		if m.selected[i] {
			checked = selectedStyle.Render("[✓]")
		}

		// Show package info: name, ecosystem, path
		info := fmt.Sprintf("%s (%s) at %s", pkg.Name, pkg.Ecosystem, pkg.Path)
		s += fmt.Sprintf("%s%s %s\n", cursor, checked, info)
	}

	s += "\n" + helpStyle.Render("space: toggle • enter: confirm • q: quit")

	return s
}

// PromptReviewPackages prompts the user to review and select detected packages
func PromptReviewPackages(packages []config.Package) ([]config.Package, error) {
	return PromptReviewPackagesFunc(packages, nil)
}

// PromptReviewPackagesFunc allows dependency injection for testing
func PromptReviewPackagesFunc(packages []config.Package, inputFunc func() ([]config.Package, error)) ([]config.Package, error) {
	// Validate packages
	if len(packages) == 0 {
		return nil, fmt.Errorf("no packages to review")
	}

	// If inputFunc provided (for testing), use it
	if inputFunc != nil {
		return inputFunc()
	}

	// Interactive prompt using Bubble Tea
	// Start with all packages selected by default
	selected := make(map[int]bool)
	for i := range packages {
		selected[i] = true
	}

	m := packageReviewModel{
		packages: packages,
		selected: selected,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("package review failed: %w", err)
	}

	result := finalModel.(packageReviewModel)
	if result.err != nil {
		return nil, result.err
	}

	// Build selected packages list
	var selectedPackages []config.Package
	for i := range result.selected {
		selectedPackages = append(selectedPackages, result.packages[i])
	}

	if len(selectedPackages) == 0 {
		return nil, fmt.Errorf("must select at least one package")
	}

	return selectedPackages, nil
}
