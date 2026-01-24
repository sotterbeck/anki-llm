package shared

import (
	"context"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
)

// Interfaces expected from the application (imported from main package).
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

// NoteItem represents a generated Anki note.
type NoteItem struct {
	Index int
	Front string
	Back  string
	Raw   map[string]string
}

// AppState is shared across screen models.
type AppState struct {
	Ctx        context.Context
	Cancel     context.CancelFunc
	Width      int
	Height     int
	NoteModel  string
	Picker     filepicker.Model
	PDFPath    string
	DeckName   string
	DeckList   []string
	DeckCursor int
	DeckInput  textinput.Model
	Notes      []NoteItem
	Selected   map[int]bool
	Cursor     int
	Status     string
	Spinner    spinner.Model
	LLM        LLM
	Anki       AnkiAPI
	Factory    ScreenFactory
	Err        error
	Loading    bool
}

// NewAppState constructs the shared application state.
func NewAppState(ctx context.Context, llm LLM, anki AnkiAPI) *AppState {
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
	deck := "Default"
	if len(deckNames) > 0 {
		deck = deckNames[0]
	}

	ti := textinput.New()
	ti.Placeholder = "New deck name"

	return &AppState{
		Ctx:        cctx,
		Cancel:     cancel,
		NoteModel:  "Basic",
		Picker:     fp,
		PDFPath:    "",
		DeckName:   deck,
		DeckList:   deckNames,
		DeckCursor: 0,
		DeckInput:  ti,
		Notes:      nil,
		Selected:   map[int]bool{},
		Cursor:     0,
		Spinner:    sp,
		LLM:        llm,
		Anki:       anki,
	}
}
