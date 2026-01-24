package deckcreate

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
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
	m.app.DeckInput.Reset()
	m.app.DeckInput.Focus()
	return tea.Batch(spinner.Tick, textinput.Blink)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmd, handled := shared.HandleSpinnerTick(m.app, msg); handled {
		return m, cmd
	}
	if handled := shared.HandleAsyncMsg(m.app, msg); handled {
		return m, nil
	}

	if mt, ok := msg.(shared.DeckCreatedMsg); ok {
		if mt.Err != nil {
			m.app.Status = "error creating deck: " + mt.Err.Error()
		} else {
			m.app.DeckName = mt.DeckName
			m.app.DeckList = append(m.app.DeckList, mt.DeckName)
			m.app.Status = "deck created"
		}
		next := m.app.Factory.NewNotes(m.app)
		return next, next.Init()
	}

	var cmd tea.Cmd
	m.app.DeckInput, cmd = m.app.DeckInput.Update(msg)

	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "esc":
			m.app.DeckInput.Blur()
			next := m.app.Factory.NewNotes(m.app)
			return next, next.Init()
		case "enter":
			deckName := m.app.DeckInput.Value()
			if deckName == "" {
				m.app.Status = "deck name cannot be empty"
				return m, nil
			}
			m.app.Status = "creating deck..."
			m.app.DeckInput.Blur()
			return m, shared.CreateDeckCmd(m.app.Anki, deckName)
		}
	}

	return m, cmd
}
