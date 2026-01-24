package shared

// Messages used by the TUI to communicate async results.

type GeneratedNotesMsg struct {
	Notes []map[string]string
}

type GenerateErrMsg struct {
	Err error
}

type AnkiResultMsg struct {
	Err error
}

type DeckCreatedMsg struct {
	DeckName string
	Err      error
}
