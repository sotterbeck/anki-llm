package notes

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

	switch mt := msg.(type) {
	case tea.KeyMsg:
		switch mt.String() {
		case "q", "ctrl+c":
			m.app.Cancel()
			return m, tea.Quit
		case "d":
			m.app.DeckCursor = 0
			m.app.Status = "select deck"
			next := m.app.Factory.NewDeckSelect(m.app)
			return next, next.Init()
		case "up", "k":
			shared.MoveNotesCursor(m.app, -1)
		case "down", "j":
			shared.MoveNotesCursor(m.app, 1)
		case " ":
			shared.ToggleCurrentNote(m.app)
		case "s":
			shared.ToggleAllNotes(m.app)
		case "a":
			sel := shared.SelectedNotes(m.app)
			if len(sel) == 0 {
				m.app.Status = "no notes selected"
				return m, nil
			}
			m.app.Loading = true
			m.app.Status = "adding to Anki..."
			return m, shared.AddNotesCmd(m.app.Anki, m.app.DeckName, m.app.NoteModel, sel)
		case "r":
			if m.app.PDFPath == "" {
				m.app.Status = "no pdf selected"
				return m, nil
			}
			m.app.Loading = true
			m.app.Status = "regenerating..."
			return m, shared.GenerateNotesCmd(m.app.Ctx, m.app.LLM, m.app.PDFPath, m.app.NoteModel)
		}
	}

	return m, nil
}
