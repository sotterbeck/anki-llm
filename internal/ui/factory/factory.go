package factory

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sotterbeck/anki-llm/internal/ui/deckcreate"
	"github.com/sotterbeck/anki-llm/internal/ui/deckselect"
	"github.com/sotterbeck/anki-llm/internal/ui/notes"
	"github.com/sotterbeck/anki-llm/internal/ui/picker"
	"github.com/sotterbeck/anki-llm/internal/ui/shared"
)

type Factory struct{}

func New() Factory {
	return Factory{}
}

func (Factory) NewPicker(app *shared.AppState) tea.Model {
	return picker.NewModel(app)
}

func (Factory) NewNotes(app *shared.AppState) tea.Model {
	return notes.NewModel(app)
}

func (Factory) NewDeckSelect(app *shared.AppState) tea.Model {
	return deckselect.NewModel(app)
}

func (Factory) NewDeckCreate(app *shared.AppState) tea.Model {
	return deckcreate.NewModel(app)
}
