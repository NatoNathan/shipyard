package prompt

import (
	"fmt"
	"strings"

	"github.com/NatoNathan/shipyard/internal/editor"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles for the summary prompt
var (
	summaryTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))

	summaryHelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

	summaryFocusedStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	summaryBlurredStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#999999")).
		Padding(0, 1)
)

// summaryModel is the Bubble Tea model for summary input
type summaryModel struct {
	textarea    textarea.Model
	err         error
	submitted   bool
	openEditor  bool
	projectPath string
}

// PromptSummaryFunc is the function signature for prompting summary
type PromptSummaryFunc func(projectPath string) (string, error)

// PromptSummary prompts for a consignment summary with text area
func PromptSummary(projectPath string) (string, error) {
	return PromptSummaryWithFunc(projectPath, runSummaryPrompt)
}

// PromptSummaryWithFunc allows injection of custom prompt function for testing
func PromptSummaryWithFunc(projectPath string, fn PromptSummaryFunc) (string, error) {
	return fn(projectPath)
}

// runSummaryPrompt runs the actual Bubble Tea program
func runSummaryPrompt(projectPath string) (string, error) {
	ta := textarea.New()
	ta.Placeholder = "Enter summary (first line) and optional description..."
	ta.SetWidth(80)
	ta.SetHeight(8)
	ta.ShowLineNumbers = false
	ta.CharLimit = 2000 // Reasonable limit for consignment
	ta.Focus()

	// Configure styles
	ta.FocusedStyle.Base = summaryFocusedStyle
	ta.BlurredStyle.Base = summaryBlurredStyle

	m := summaryModel{
		textarea:    ta,
		projectPath: projectPath,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run prompt: %w", err)
	}

	result := finalModel.(summaryModel)
	if result.err != nil {
		return "", result.err
	}

	if result.openEditor {
		// User pressed Ctrl+E - open editor with current content
		return openEditorForSummary(projectPath, result.textarea.Value())
	}

	// Extract first line as summary
	value := strings.TrimSpace(result.textarea.Value())
	if value == "" {
		return "", fmt.Errorf("summary cannot be empty")
	}

	lines := strings.Split(value, "\n")
	return strings.TrimSpace(lines[0]), nil
}

// openEditorForSummary opens the system editor with initial content
func openEditorForSummary(projectPath, initialContent string) (string, error) {
	// Prepare initial content with instructions
	template := "# Enter your change description here\n# First line: summary\n# Remaining lines: detailed description\n\n"
	if initialContent != "" {
		template += initialContent
	}

	content, err := editor.OpenEditor(projectPath, template)
	if err != nil {
		return "", err
	}

	// Parse first non-comment line as summary
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			return trimmed, nil
		}
	}

	return "", fmt.Errorf("no summary provided")
}

// Init implements tea.Model
func (m summaryModel) Init() tea.Cmd {
	return textarea.Blink
}

// Update implements tea.Model
func (m summaryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.err = fmt.Errorf("cancelled")
			m.submitted = true
			return m, tea.Quit

		case "ctrl+e":
			// Open editor
			m.openEditor = true
			m.submitted = true
			return m, tea.Quit

		case "enter":
			// Check if we have content on first line
			value := strings.TrimSpace(m.textarea.Value())
			if value != "" {
				lines := strings.Split(value, "\n")
				if strings.TrimSpace(lines[0]) != "" {
					m.submitted = true
					return m, tea.Quit
				}
			}
			// Otherwise, insert newline (default textarea behavior)

		case "esc":
			if m.textarea.Focused() {
				m.textarea.Blur()
			} else {
				m.textarea.Focus()
			}
			return m, nil
		}

	case error:
		m.err = msg
		return m, tea.Quit
	}

	// Update textarea
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

// View implements tea.Model
func (m summaryModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(summaryTitleStyle.Render("Change Summary"))
	b.WriteString("\n\n")

	// Textarea
	b.WriteString(m.textarea.View())
	b.WriteString("\n\n")

	// Help text
	help := summaryHelpStyle.Render("First line = summary • Ctrl+E = open editor • Enter = submit • Ctrl+C = cancel")
	b.WriteString(help)
	b.WriteString("\n")

	return b.String()
}
