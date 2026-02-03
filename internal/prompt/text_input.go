package prompt

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type textInputModel struct {
	message      string
	defaultValue string
	value        string
	done         bool
	err          error
}

func (m textInputModel) Init() tea.Cmd {
	return nil
}

func (m textInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.err = fmt.Errorf("cancelled")
			m.done = true
			return m, tea.Quit

		case "enter":
			if m.value == "" {
				m.value = m.defaultValue
			}
			m.done = true
			return m, tea.Quit

		case "backspace":
			if len(m.value) > 0 {
				m.value = m.value[:len(m.value)-1]
			}

		default:
			// Add character to input
			if len(msg.String()) == 1 {
				m.value += msg.String()
			}
		}
	}

	return m, nil
}

func (m textInputModel) View() string {
	if m.done {
		return ""
	}

	s := titleStyle.Render(m.message) + "\n\n"

	// Show input field with cursor
	inputValue := m.value
	if inputValue == "" && m.defaultValue != "" {
		inputValue = helpStyle.Render(m.defaultValue)
	} else {
		inputValue = selectedStyle.Render(m.value + "█")
	}

	s += "  " + inputValue + "\n"
	s += "\n" + helpStyle.Render("enter: confirm • esc: cancel")

	return s
}

// PromptTextInput prompts the user for text input
func PromptTextInput(message, defaultValue string) (string, error) {
	return PromptTextInputFunc(message, defaultValue, nil)
}

// PromptTextInputFunc allows dependency injection for testing
func PromptTextInputFunc(message, defaultValue string, inputFunc func() (string, error)) (string, error) {
	// If inputFunc provided (for testing), use it
	if inputFunc != nil {
		value, err := inputFunc()
		if err != nil {
			return "", err
		}
		if value == "" {
			value = defaultValue
		}
		return value, nil
	}

	// Interactive prompt using Bubble Tea
	m := textInputModel{
		message:      message,
		defaultValue: defaultValue,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("text input failed: %w", err)
	}

	result := finalModel.(textInputModel)
	if result.err != nil {
		return "", result.err
	}

	value := strings.TrimSpace(result.value)
	if value == "" {
		value = defaultValue
	}

	return value, nil
}
