package prompt

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type confirmModel struct {
	message  string
	selected bool // true = yes, false = no
	done     bool
	err      error
}

func (m confirmModel) Init() tea.Cmd {
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.err = fmt.Errorf("cancelled")
			m.done = true
			return m, tea.Quit

		case "y", "Y":
			m.selected = true
			m.done = true
			return m, tea.Quit

		case "n", "N":
			m.selected = false
			m.done = true
			return m, tea.Quit

		case "enter":
			m.done = true
			return m, tea.Quit

		case "left", "h":
			m.selected = true

		case "right", "l":
			m.selected = false
		}
	}

	return m, nil
}

func (m confirmModel) View() string {
	if m.done {
		return ""
	}

	s := titleStyle.Render(m.message) + "\n\n"

	// Show Yes/No options
	yesStyle := ""
	noStyle := ""
	if m.selected {
		yesStyle = selectedStyle.Render("[Yes]")
		noStyle = "No"
	} else {
		yesStyle = "Yes"
		noStyle = selectedStyle.Render("[No]")
	}

	s += fmt.Sprintf("  %s  %s\n", yesStyle, noStyle)
	s += "\n" + helpStyle.Render("←/→: select • enter/y/n: confirm • q: quit")

	return s
}

// PromptConfirm prompts the user for yes/no confirmation
func PromptConfirm(message string, defaultYes bool) (bool, error) {
	return PromptConfirmFunc(message, defaultYes, nil)
}

// PromptConfirmFunc allows dependency injection for testing
func PromptConfirmFunc(message string, defaultYes bool, inputFunc func() (bool, error)) (bool, error) {
	// If inputFunc provided (for testing), use it
	if inputFunc != nil {
		return inputFunc()
	}

	// Interactive prompt using Bubble Tea
	m := confirmModel{
		message:  message,
		selected: defaultYes,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("confirmation prompt failed: %w", err)
	}

	result := finalModel.(confirmModel)
	if result.err != nil {
		return false, result.err
	}

	return result.selected, nil
}
