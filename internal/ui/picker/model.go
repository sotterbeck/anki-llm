package picker

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
	return tea.Batch(spinner.Tick, m.app.Picker.Init())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmd, handled := shared.HandleSpinnerTick(m.app, msg); handled {
		return m, cmd
	}
	if handled := shared.HandleAsyncMsg(m.app, msg); handled {
		return m, nil
	}

	pickerModel, cmd := m.app.Picker.Update(msg)
	m.app.Picker = pickerModel
	if cmd != nil {
		return m, cmd
	}

	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "q", "ctrl+c":
			m.app.Cancel()
			return m, tea.Quit
		}
	}

	if did, path := m.app.Picker.DidSelectFile(msg); did {
		m.app.PDFPath = path
		m.app.Loading = true
		m.app.Status = "generating notes..."
		next := m.app.Factory.NewNotes(m.app)
		return next, tea.Batch(shared.GenerateNotesCmd(m.app.Ctx, m.app.LLM, m.app.PDFPath, m.app.NoteModel), next.Init())
	}

	if didDisabled, _ := m.app.Picker.DidSelectDisabledFile(msg); didDisabled {
		m.app.Status = "cannot select that file"
	}

	return m, nil
}
