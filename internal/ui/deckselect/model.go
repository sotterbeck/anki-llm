package deckselect

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sotterbeck/anki-llm/internal/ui/shared"
)

type Model struct {
	app *shared.AppState
}

func NewModel(app *shared.AppState) Model {
	return Model{app: app}
}

func (m Model) Init() tea.Cmd {
	return spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmd, handled := shared.HandleSpinnerTick(m.app, msg); handled {
		return m, cmd
	}
	if handled := shared.HandleAsyncMsg(m.app, msg); handled {
		return m, nil
	}

	decks := shared.VisibleDecks(m.app)

	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "q", "ctrl+c", "esc":
			m.app.Status = ""
			next := m.app.Factory.NewNotes(m.app)
			return next, next.Init()
		case "up", "k":
			m.app.DeckCursor = maxInt(0, m.app.DeckCursor-1)
		case "down", "j":
			m.app.DeckCursor = minInt(len(decks)-1, m.app.DeckCursor+1)
		case "enter":
			selected := decks[m.app.DeckCursor]
			if selected == "+ Create new deck" {
				next := m.app.Factory.NewDeckCreate(m.app)
				return next, next.Init()
			}
			m.app.DeckName = selected
			m.app.Status = "deck changed to " + selected
			next := m.app.Factory.NewNotes(m.app)
			return next, next.Init()
		}
	}

	return m, nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
