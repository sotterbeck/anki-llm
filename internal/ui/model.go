package ui

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
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
	ListDeckNames() ([]string, error)
	CreateDeck(deckName string) error
}

type AppState int

const (
	StatePickingPDF AppState = iota
	StateViewingNotes
	StateSelectingDeck
	StateCreatingDeck
)

// NoteItem represents a generated Anki note.
type NoteItem struct {
	Index int
	Front string
	Back  string
	Raw   map[string]string
}

// Model is the Bubble Tea model for the UI.
type Model struct {
	ctx          context.Context
	cancel       context.CancelFunc
	width        int
	height       int
	pdfPath      string
	pdfList      []string
	picker       filepicker.Model
	noteModel    string
	deckName     string
	deckList     []string
	deckCursor   int
	newDeckInput textinput.Model
	notes        []NoteItem
	selected     map[int]bool
	cursor       int
	status       string
	spinner      spinner.Model
	llm          LLM
	anki         AnkiAPI
	err          error
	loading      bool
	search       string
	state        AppState
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

	deckNames, _ := anki.ListDeckNames()
	var deck = "Default"

	if len(deckNames) > 0 {
		deck = deckNames[0]
	}

	ti := textinput.New()
	ti.Placeholder = "New deck name"

	return &Model{
		ctx:          cctx,
		cancel:       cancel,
		noteModel:    "Basic",
		deckName:     deck,
		deckList:     deckNames,
		deckCursor:   0,
		newDeckInput: ti,
		selected:     map[int]bool{},
		spinner:      sp,
		llm:          llm,
		anki:         anki,
		picker:       fp,
		state:        StatePickingPDF,
	}
}

func (m *Model) setState(newState AppState) tea.Cmd {
	m.state = newState

	switch newState {
	case StatePickingPDF:
		return tea.Batch(spinner.Tick, m.picker.Init())
	case StateCreatingDeck:
		m.newDeckInput.Reset()
		m.newDeckInput.Focus()
		return tea.Batch(spinner.Tick, textinput.Blink)
	case StateViewingNotes:
	case StateSelectingDeck:
	}
	return spinner.Tick
}

func (m *Model) Init() tea.Cmd {
	return m.setState(m.state)
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

// generateNotesCmd triggers background generation (returns a command)
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
func addNotesCmd(anki AnkiAPI, deck, model string, notes []map[string]string) tea.Cmd {
	return func() tea.Msg {
		err := anki.AddNotes(deck, model, notes)
		return ankiResultMsg{err}
	}
}

// createDeckCmd triggers deck creation in Anki
func createDeckCmd(anki AnkiAPI, deckName string) tea.Cmd {
	return func() tea.Msg {
		err := anki.CreateDeck(deckName)
		return deckCreatedMsg{DeckName: deckName, Err: err}
	}
}

// Update processes incoming messages and key events
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch mt := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(mt)
		cmds = append(cmds, cmd)
	case tea.KeyMsg:
		switch m.state {
		case StatePickingPDF:
			nm, cmd := m.handlePickerMsg(mt)
			return nm, cmd
		case StateSelectingDeck:
			return m.handleDeckSelection(mt)
		case StateCreatingDeck:
			return m.handleNewDeckCreation(mt)
		case StateViewingNotes:
			return m.handleViewingNotes(mt)
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
	case deckCreatedMsg:
		if mt.Err != nil {
			m.status = "error creating deck: " + mt.Err.Error()
		} else {
			m.deckName = mt.DeckName
			m.deckList = append(m.deckList, mt.DeckName)
			m.status = "deck created"
		}
		return m, m.setState(StateViewingNotes)
	default:
		// Send all other messages (including filepicker internal ones) to the active state handler
		if m.state == StatePickingPDF {
			return m.handlePickerMsg(msg)
		}
	}

	return m, tea.Batch(cmds...)
}

// handleViewingNotes processes key inputs to modify the viewing state of notes or perform actions like selection and regeneration.
func (m *Model) handleViewingNotes(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.cancel()
		return m, tea.Quit
	case "d":
		m.deckCursor = 0
		m.status = "select deck"
		return m, m.setState(StateSelectingDeck)
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.notes)-1 {
			m.cursor++
		}
	case " ":
		if _, ok := m.selected[m.cursor]; ok {
			delete(m.selected, m.cursor)
		} else {
			m.selected[m.cursor] = true
		}
	case "s":
		if len(m.selected) == len(m.notes) {
			m.selected = map[int]bool{}
		} else {
			for i := range m.notes {
				m.selected[i] = true
			}
		}
	case "a":
		sel := m.getSelectedNotes()
		if len(sel) == 0 {
			m.status = "no notes selected"
			return m, nil
		}
		m.loading = true
		m.status = "adding to Anki..."
		return m, addNotesCmd(m.anki, m.deckName, m.noteModel, sel)
	case "r":
		if m.pdfPath == "" {
			m.status = "no pdf selected"
			return m, nil
		}
		m.loading = true
		m.status = "regenerating..."
		return m, generateNotesCmd(m.ctx, m.llm, m.pdfPath, m.noteModel)
	}
	return m, nil
}

// handlePickerMsg handles updates to the file picker state and processes user interactions during file selection.
func (m *Model) handlePickerMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	picker, cmd := m.picker.Update(msg)
	m.picker = picker
	if cmd != nil {
		return m, cmd
	}

	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "q", "ctrl+c":
			m.cancel()
			return m, tea.Quit
		}
	}

	if did, path := m.picker.DidSelectFile(msg); did {
		m.pdfPath = path
		m.loading = true
		m.status = "generating notes..."
		m.state = StateViewingNotes
		return m, generateNotesCmd(m.ctx, m.llm, m.pdfPath, m.noteModel)
	}

	if didDisabled, _ := m.picker.DidSelectDisabledFile(msg); didDisabled {
		m.status = "cannot select that file"
	}

	return m, nil
}

// handleDeckSelection handles key events when selecting a deck.
func (m *Model) handleDeckSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	visibleDecks := m.getVisibleDecks()

	switch msg.String() {
	case "q", "ctrl+c", "esc":
		m.status = ""
		return m, m.setState(StateViewingNotes)
	case "up", "k":
		if m.deckCursor > 0 {
			m.deckCursor--
		}
	case "down", "j":
		if m.deckCursor < len(visibleDecks)-1 {
			m.deckCursor++
		}
	case "enter":
		selected := visibleDecks[m.deckCursor]
		if selected == "+ Create new deck" {
			return m, m.setState(StateCreatingDeck)
		}
		m.deckName = selected
		m.status = "deck changed to " + selected
		return m, m.setState(StateViewingNotes)
	}
	return m, nil
}

// handleNewDeckCreation handles key events when creating a new deck.
func (m *Model) handleNewDeckCreation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.newDeckInput, cmd = m.newDeckInput.Update(msg)

	switch msg.String() {
	case "esc":
		m.newDeckInput.Blur()
		return m, m.setState(StateViewingNotes)
	case "enter":
		deckName := m.newDeckInput.Value()
		if deckName == "" {
			m.status = "deck name cannot be empty"
			return m, nil
		}
		m.status = "creating deck..."
		m.newDeckInput.Blur()
		return m, createDeckCmd(m.anki, deckName)
	}

	return m, cmd
}

// getVisibleDecks returns the list of decks shown in the deck selector.
func (m *Model) getVisibleDecks() []string {
	decks := make([]string, len(m.deckList)+1)
	decks[0] = "+ Create new deck"
	for i, deck := range m.deckList {
		decks[i+1] = deck
	}
	return decks
}

// getSelectedNotes returns a slice of raw note data for all selected notes in the model.
func (m *Model) getSelectedNotes() []map[string]string {
	var sel []map[string]string
	for i := range m.selected {
		sel = append(sel, m.notes[i].Raw)
	}
	return sel
}
