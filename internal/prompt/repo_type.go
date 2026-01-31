package prompt

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// RepoType represents the type of repository
type RepoType string

const (
	RepoTypeMonorepo RepoType = "monorepo"
	RepoTypeSingle   RepoType = "single"
)

type repoTypeModel struct {
	options  []repoTypeOption
	cursor   int
	selected RepoType
	done     bool
	err      error
}

type repoTypeOption struct {
	value       RepoType
	label       string
	description string
}

func (m repoTypeModel) Init() tea.Cmd {
	return nil
}

func (m repoTypeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m repoTypeModel) View() string {
	if m.done {
		return ""
	}

	s := titleStyle.Render("What type of repository is this?") + "\n\n"

	for i, opt := range m.options {
		cursor := "  "
		if m.cursor == i {
			cursor = cursorStyle.Render("> ")
			s += fmt.Sprintf("%s%s\n", cursor, selectedStyle.Render(opt.label))
			s += fmt.Sprintf("   %s\n\n", helpStyle.Render(opt.description))
		} else {
			s += fmt.Sprintf("%s%s\n", cursor, opt.label)
			s += fmt.Sprintf("   %s\n\n", helpStyle.Render(opt.description))
		}
	}

	s += helpStyle.Render("↑/↓: navigate • enter: confirm • q: quit")

	return s
}

// PromptRepoType prompts the user to select repository type
func PromptRepoType() (RepoType, error) {
	return PromptRepoTypeFunc(nil)
}

// PromptRepoTypeFunc allows dependency injection for testing
func PromptRepoTypeFunc(inputFunc func() (RepoType, error)) (RepoType, error) {
	// If inputFunc provided (for testing), use it
	if inputFunc != nil {
		return inputFunc()
	}

	// Interactive prompt using Bubble Tea
	m := repoTypeModel{
		options: []repoTypeOption{
			{
				value:       RepoTypeSingle,
				label:       "Single repository",
				description: "One package/project in this repository",
			},
			{
				value:       RepoTypeMonorepo,
				label:       "Monorepo",
				description: "Multiple packages/projects in this repository",
			},
		},
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("repository type selection failed: %w", err)
	}

	result := finalModel.(repoTypeModel)
	if result.err != nil {
		return "", result.err
	}

	return result.selected, nil
}
