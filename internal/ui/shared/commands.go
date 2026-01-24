package shared

import (
	"context"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// GenerateNotesCmd triggers background generation.
func GenerateNotesCmd(ctx context.Context, llm LLM, path, noteModel string) tea.Cmd {
	return func() tea.Msg {
		f, err := os.Open(path)
		if err != nil {
			return GenerateErrMsg{Err: err}
		}
		defer f.Close()

		cctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
		defer cancel()
		notes, err := llm.GenerateAnkiNotes(cctx, f, noteModel)
		if err != nil {
			return GenerateErrMsg{Err: err}
		}
		return GeneratedNotesMsg{Notes: notes}
	}
}

// AddNotesCmd triggers add-to-Anki.
func AddNotesCmd(anki AnkiAPI, deck, model string, notes []map[string]string) tea.Cmd {
	return func() tea.Msg {
		err := anki.AddNotes(deck, model, notes)
		return AnkiResultMsg{Err: err}
	}
}

// CreateDeckCmd triggers deck creation in Anki.
func CreateDeckCmd(anki AnkiAPI, deckName string) tea.Cmd {
	return func() tea.Msg {
		err := anki.CreateDeck(deckName)
		return DeckCreatedMsg{DeckName: deckName, Err: err}
	}
}
