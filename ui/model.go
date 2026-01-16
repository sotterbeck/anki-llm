package ui

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Interfaces expected from the application (imported from main package)
// We duplicate small interfaces here to avoid import cycles; callers will pass concrete implementations.

type LLM interface {
	GenerateAnkiNotes(ctx context.Context, r io.Reader, noteModel string) ([]map[string]string, error)
	Close() error
}

type AnkiAPI interface {
	AddNotes(deckName, modelName string, notes []map[string]string) error
}

// NoteItem represents a generated Anki note.
type NoteItem struct {
	Index int
	Front string
	Back  string
	Raw   map[string]string
}

// Model is the Bubble Tea model for the UI.
type Model struct {
	ctx       context.Context
	cancel    context.CancelFunc
	width     int
	height    int
	pdfPath   string
	pdfList   []string
	picking   bool
	picker    filepicker.Model
	noteModel string
	deckName  string
	modelName string
	notes     []NoteItem
	selected  map[int]bool
	cursor    int
	status    string
	spinner   spinner.Model
	llm       LLM
	anki      AnkiAPI
	err       error
	loading   bool
	showHelp  bool
	search    string
}

// NewModel constructs a UI Model. Provide llm and anki implementations.
func NewModel(ctx context.Context, llm LLM, anki AnkiAPI) *Model {
	cctx, cancel := context.WithCancel(ctx)
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	fp := filepicker.New()
	fp.AllowedTypes = []string{".pdf"}
	fp.DirAllowed = false
	fp.FileAllowed = true
	if wd, err := os.Getwd(); err == nil {
		fp.CurrentDirectory = wd
	}

	return &Model{
		ctx:       cctx,
		cancel:    cancel,
		noteModel: "Basic",
		deckName:  "Default",
		modelName: "Basic",
		selected:  map[int]bool{},
		spinner:   sp,
		llm:       llm,
		anki:      anki,
		picker:    fp,
		picking:   true,
	}
}

func (m *Model) Init() tea.Cmd {
	if m.picking {
		return tea.Batch(spinner.Tick, m.picker.Init())
	}
	if m.pdfPath != "" {
		m.loading = true
		return tea.Batch(spinner.Tick, generateNotesCmd(m.ctx, m.llm, m.pdfPath, m.noteModel))
	}
	return spinner.Tick
}

// helper to convert raw notes to NoteItem
func notesToItems(raw []map[string]string) []NoteItem {
	var out []NoteItem
	for i, r := range raw {
		front := r["Front"]
		back := r["Back"]
		out = append(out, NoteItem{Index: i, Front: front, Back: back, Raw: r})
	}
	return out
}

// generateNotes triggers background generation (returns a command)
func generateNotesCmd(ctx context.Context, llm LLM, path, noteModel string) tea.Cmd {
	return func() tea.Msg {
		f, err := os.Open(path)
		if err != nil {
			return generateErrMsg{err}
		}
		defer f.Close()
		// use a short timeout for safety
		cctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
		defer cancel()
		notes, err := llm.GenerateAnkiNotes(cctx, f, noteModel)
		if err != nil {
			return generateErrMsg{err}
		}
		return generatedNotesMsg{notes}
	}
}

// addNotesCmd triggers add-to-anki
func addNotesCmd(ctx context.Context, anki AnkiAPI, deck, model string, notes []map[string]string) tea.Cmd {
	return func() tea.Msg {
		err := anki.AddNotes(deck, model, notes)
		return ankiResultMsg{err}
	}
}

// Update processes incoming messages and key events
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch mt := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(mt)
		return m, cmd
	}

	model, cmd, done := m.updateFilePicker(msg)
	if done {
		return model, cmd
	}

	switch mt := msg.(type) {
	case tea.KeyMsg:
		switch mt.String() {
		case "q", "ctrl+c":
			m.cancel()
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.notes)-1 {
				m.cursor++
			}
		case " ": // space toggle selection
			if _, ok := m.selected[m.cursor]; ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = true
			}
		case "s": // select all
			if len(m.selected) == len(m.notes) {
				m.selected = map[int]bool{}
			} else {
				for i := range m.notes {
					m.selected[i] = true
				}
			}
		case "a": // add selected
			sel := m.getSelectedNotes()
			if len(sel) == 0 {
				m.status = "no notes selected"
				return m, nil
			}
			m.loading = true
			m.status = "adding to Anki..."
			return m, addNotesCmd(m.ctx, m.anki, m.deckName, m.modelName, sel)
		case "r":
			if m.pdfPath == "" {
				m.status = "no pdf selected"
				return m, nil
			}
			m.loading = true
			m.status = "regenerating..."
			return m, generateNotesCmd(m.ctx, m.llm, m.pdfPath, m.noteModel)
		}
	case generatedNotesMsg:
		m.loading = false
		m.status = "generated"
		m.notes = notesToItems(mt.Notes)
		m.selected = map[int]bool{}
		m.cursor = 0
		return m, nil
	case generateErrMsg:
		m.loading = false
		m.err = mt.Err
		m.status = "generation error"
		return m, nil
	case ankiResultMsg:
		m.loading = false
		if mt.Err != nil {
			m.status = "anki error: " + mt.Err.Error()
		} else {
			m.status = "added to anki"
		}
		return m, nil
	}
	return m, nil
}

// getSelectedNotes returns a slice of raw note data for all selected notes in the model.
func (m *Model) getSelectedNotes() []map[string]string {
	var sel []map[string]string
	for i := range m.selected {
		sel = append(sel, m.notes[i].Raw)
	}
	return sel
}

// updateFilePicker handles updates to the file picker state and processes user interactions during file selection.
func (m *Model) updateFilePicker(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	if m.picking {
		picker, cmd := m.picker.Update(msg)
		m.picker = picker
		if cmd != nil {
			return m, cmd, true
		}

		if km, ok := msg.(tea.KeyMsg); ok {
			switch km.String() {
			case "q", "ctrl+c":
				m.cancel()
				return m, tea.Quit, true
			}
		}

		if did, path := m.picker.DidSelectFile(msg); did {
			m.pdfPath = path
			m.picking = false
			m.loading = true
			m.status = "generating notes..."
			return m, generateNotesCmd(m.ctx, m.llm, m.pdfPath, m.noteModel), true
		}

		if didDisabled, _ := m.picker.DidSelectDisabledFile(msg); didDisabled {
			m.status = "cannot select that file"
		}

		return m, nil, true
	}
	return nil, nil, false
}
