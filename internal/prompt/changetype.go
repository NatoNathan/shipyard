package prompt

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/NatoNathan/shipyard/pkg/types"
)

type changeTypeModel struct {
	options  []changeTypeOption
	cursor   int
	selected types.ChangeType
	done     bool
	err      error
}

type changeTypeOption struct {
	value       types.ChangeType
	label       string
	description string
}

func (m changeTypeModel) Init() tea.Cmd {
	return nil
}

func (m changeTypeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.err = fmt.Errorf("cancelled")
			m.done = true
			return m, tea.Quit

		case "enter":
			m.selected = m.options[m.cursor].value
			m.done = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		}
	}

	return m, nil
}

func (m changeTypeModel) View() string {
	if m.done {
		return ""
	}

	s := titleStyle.Render("Select change type:") + "\n\n"

	for i, opt := range m.options {
		cursor := "  "
		if m.cursor == i {
			cursor = cursorStyle.Render("> ")
			s += fmt.Sprintf("%s%s - %s\n", cursor, selectedStyle.Render(opt.label), opt.description)
		} else {
			s += fmt.Sprintf("%s%s - %s\n", cursor, opt.label, opt.description)
		}
	}

	s += "\n" + helpStyle.Render("↑/↓: navigate • enter: confirm • q: quit")

	return s
}

// PromptForChangeType prompts the user to select a change type
func PromptForChangeType() (types.ChangeType, error) {
	return PromptForChangeTypeFunc(nil)
}

// PromptForChangeTypeFunc allows dependency injection for testing
func PromptForChangeTypeFunc(inputFunc func() (types.ChangeType, error)) (types.ChangeType, error) {
	// If inputFunc provided (for testing), use it
	if inputFunc != nil {
		selected, err := inputFunc()
		if err != nil {
			return "", err
		}
		if selected == "" {
			return "", fmt.Errorf("must select a change type")
		}
		return selected, nil
	}

	// Interactive prompt using Bubble Tea
	m := changeTypeModel{
		options: []changeTypeOption{
			{types.ChangeTypePatch, "patch", "Backwards compatible bug fixes"},
			{types.ChangeTypeMinor, "minor", "Backwards compatible new features"},
			{types.ChangeTypeMajor, "major", "Breaking changes"},
		},
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("change type selection failed: %w", err)
	}

	result := finalModel.(changeTypeModel)
	if result.err != nil {
		return "", result.err
	}

	if result.selected == "" {
		return "", fmt.Errorf("must select a change type")
	}

	return result.selected, nil
}
