package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sotterbeck/anki-llm/internal/ui/factory"
	"github.com/sotterbeck/anki-llm/internal/ui/picker"
	"github.com/sotterbeck/anki-llm/internal/ui/shared"
)

type LLM = shared.LLM
type AnkiAPI = shared.AnkiAPI

// NewModel constructs the initial Bubble Tea model. Provide llm and anki implementations.
func NewModel(ctx context.Context, llm LLM, anki AnkiAPI) tea.Model {
	app := shared.NewAppState(ctx, llm, anki)
	app.Factory = factory.New()
	return picker.NewModel(app)
}
