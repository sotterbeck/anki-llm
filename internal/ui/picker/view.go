package picker

import "github.com/sotterbeck/anki-llm/internal/ui/shared"

func (m Model) View() string {
	if errView := shared.RenderError(m.app); errView != "" {
		return errView
	}
	hints := "up/down:move  enter:select  q:quit"
	return shared.TitleStyle.Render("Choose PDF") + "\n" + m.app.Picker.View() + "\n" + shared.RenderFooter(m.app, hints)
}
