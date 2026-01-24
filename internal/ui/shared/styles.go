package shared

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	ListStyle  = lipgloss.NewStyle().Padding(0, 1)
	SelStyle   = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230"))
	TitleStyle = lipgloss.NewStyle().Bold(true)
)

// RenderFooter renders the shared footer with hints and status.
func RenderFooter(app *AppState, hints string) string {
	status := app.Status
	sp := ""
	if app.Loading {
		status = "working"
		sp = " " + app.Spinner.View()
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(hints + "    " + status + sp)
}

// RenderError renders the error view, if any.
func RenderError(app *AppState) string {
	if app.Err == nil {
		return ""
	}
	return fmt.Sprintf("Error: %v\n", app.Err)
}
