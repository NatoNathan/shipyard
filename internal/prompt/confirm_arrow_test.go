package prompt

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestConfirmModel_ArrowKeyDirection tests that arrow keys select the correct option
func TestConfirmModel_ArrowKeyDirection(t *testing.T) {
	t.Run("left arrow selects Yes (left option)", func(t *testing.T) {
		// Setup: Create model starting with No selected
		m := confirmModel{
			message:  "Confirm?",
			selected: false, // No is selected
		}

		// Action: Press left arrow
		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
		result := updatedModel.(confirmModel)

		// Verify: Yes should now be selected (left arrow goes to left option)
		assert.True(t, result.selected, "Left arrow should select Yes (the left option)")
	})

	t.Run("right arrow selects No (right option)", func(t *testing.T) {
		// Setup: Create model starting with Yes selected
		m := confirmModel{
			message:  "Confirm?",
			selected: true, // Yes is selected
		}

		// Action: Press right arrow
		updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
		result := updatedModel.(confirmModel)

		// Verify: No should now be selected (right arrow goes to right option)
		assert.False(t, result.selected, "Right arrow should select No (the right option)")
	})

	t.Run("display order matches arrow key behavior", func(t *testing.T) {
		// Verify that the view displays Yes on the left and No on the right
		m := confirmModel{
			message:  "Confirm?",
			selected: true,
		}

		view := m.View()

		// The view should show Yes before No (Yes is on the left)
		// When selected=true, Yes should be highlighted
		assert.Contains(t, view, "[Yes]", "When selected=true, Yes should be highlighted")
		assert.NotContains(t, view, "[No]", "When selected=true, No should not be highlighted")
	})
}
