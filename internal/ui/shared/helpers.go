package shared

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func notesToItems(raw []map[string]string) []NoteItem {
	var out []NoteItem
	for i, r := range raw {
		front := r["Front"]
		back := r["Back"]
		out = append(out, NoteItem{Index: i, Front: front, Back: back, Raw: r})
	}
	return out
}

// ResetNotes replaces the note list and clears selection.
func ResetNotes(app *AppState, raw []map[string]string) {
	app.Notes = notesToItems(raw)
	app.Selected = map[int]bool{}
	app.Cursor = 0
}

// MoveNotesCursor moves the cursor within the notes list.
func MoveNotesCursor(app *AppState, delta int) {
	if len(app.Notes) == 0 {
		return
	}
	next := app.Cursor + delta
	if next < 0 {
		next = 0
	}
	if next > len(app.Notes)-1 {
		next = len(app.Notes) - 1
	}
	app.Cursor = next
}

// ToggleCurrentNote toggles the selection for the current note.
func ToggleCurrentNote(app *AppState) {
	if len(app.Notes) == 0 {
		return
	}
	if _, ok := app.Selected[app.Cursor]; ok {
		delete(app.Selected, app.Cursor)
		return
	}
	app.Selected[app.Cursor] = true
}

// ToggleAllNotes selects or clears all notes.
func ToggleAllNotes(app *AppState) {
	if len(app.Notes) == 0 {
		return
	}
	if len(app.Selected) == len(app.Notes) {
		app.Selected = map[int]bool{}
		return
	}
	for i := range app.Notes {
		app.Selected[i] = true
	}
}

// SelectedNotes returns the raw note data for all selected notes.
func SelectedNotes(app *AppState) []map[string]string {
	var sel []map[string]string
	for i := range app.Selected {
		sel = append(sel, app.Notes[i].Raw)
	}
	return sel
}

// VisibleDecks returns the list of decks shown in the deck selector.
func VisibleDecks(app *AppState) []string {
	decks := make([]string, len(app.DeckList)+1)
	decks[0] = "+ Create new deck"
	for i, deck := range app.DeckList {
		decks[i+1] = deck
	}
	return decks
}

// HandleSpinnerTick updates the shared spinner state.
func HandleSpinnerTick(app *AppState, msg tea.Msg) (tea.Cmd, bool) {
	mt, ok := msg.(spinner.TickMsg)
	if !ok {
		return nil, false
	}
	var cmd tea.Cmd
	app.Spinner, cmd = app.Spinner.Update(mt)
	return cmd, true
}

// HandleAsyncMsg processes shared async messages.
func HandleAsyncMsg(app *AppState, msg tea.Msg) bool {
	switch mt := msg.(type) {
	case GeneratedNotesMsg:
		app.Loading = false
		app.Status = "generated"
		ResetNotes(app, mt.Notes)
		return true
	case GenerateErrMsg:
		app.Loading = false
		app.Err = mt.Err
		app.Status = "generation error"
		return true
	case AnkiResultMsg:
		app.Loading = false
		if mt.Err != nil {
			app.Status = "anki error: " + mt.Err.Error()
		} else {
			app.Status = "added to anki"
		}
		return true
	default:
		return false
	}
}
