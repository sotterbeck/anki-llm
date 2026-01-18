package ui

// Messages used by the TUI to communicate async results.

type generatedNotesMsg struct {
	Notes []map[string]string
}

type generateErrMsg struct {
	Err error
}

type ankiResultMsg struct {
	Err error
}

type deckCreatedMsg struct {
	DeckName string
	Err      error
}
