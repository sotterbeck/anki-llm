package deckcreate

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/sotterbeck/anki-llm/internal/ui/shared"
)

func (m Model) View() string {
	if errView := shared.RenderError(m.app); errView != "" {
		return errView
	}
	hints := "enter:confirm  esc:cancel  q:quit"
	return shared.TitleStyle.Render("Create New Deck") + "\n" + m.renderNewDeckInput() + "\n" + shared.RenderFooter(m.app, hints)
}

func (m Model) renderNewDeckInput() string {
	return lipgloss.NewStyle().Padding(0, 1).Render("Deck name: " + m.app.DeckInput.View())
}
