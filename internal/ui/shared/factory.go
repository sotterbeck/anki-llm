package shared

import tea "github.com/charmbracelet/bubbletea"

// ScreenFactory constructs screen models without causing package cycles.
type ScreenFactory interface {
	NewPicker(app *AppState) tea.Model
	NewNotes(app *AppState) tea.Model
	NewDeckSelect(app *AppState) tea.Model
	NewDeckCreate(app *AppState) tea.Model
}
