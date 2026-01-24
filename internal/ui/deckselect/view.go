package deckselect

import (
	"fmt"
	"strings"

	"github.com/sotterbeck/anki-llm/internal/ui/shared"
)

func (m Model) View() string {
	if errView := shared.RenderError(m.app); errView != "" {
		return errView
	}
	hints := "j/k:move  enter:select  esc:cancel  q:quit"
	return shared.TitleStyle.Render("Select Deck") + "\n" + m.renderDeckSelector() + "\n" + shared.RenderFooter(m.app, hints)
}

func (m Model) renderDeckSelector() string {
	var b strings.Builder
	decks := shared.VisibleDecks(m.app)
	for i, deck := range decks {
		cursor := " "
		if i == m.app.DeckCursor {
			cursor = ">"
		}
		line := fmt.Sprintf("%s %s", cursor, deck)
		if i == m.app.DeckCursor {
			b.WriteString(shared.SelStyle.Render(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}
	return shared.ListStyle.Render(b.String())
}
