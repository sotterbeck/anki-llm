package notes

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sotterbeck/anki-llm/internal/ui/shared"
)

func (m Model) View() string {
	if errView := shared.RenderError(m.app); errView != "" {
		return errView
	}
	left := m.renderList()
	right := m.renderPreview()
	cols := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	hints := fmt.Sprintf(
		"j/k:move  space:toggle  a:add  s:select-all d:change-deck (%s) r:regenerate  q:quit",
		m.app.DeckName,
	)
	return shared.TitleStyle.Render("Select Notes") + "\n" + cols + "\n" + shared.RenderFooter(m.app, hints)
}

func (m Model) renderList() string {
	var b strings.Builder
	for i, it := range m.app.Notes {
		cursor := " "
		if i == m.app.Cursor {
			cursor = ">"
		}
		chk := "[ ]"
		if m.app.Selected[i] {
			chk = "[x]"
		}
		line := fmt.Sprintf("%s %s %s", cursor, chk, it.Front)
		if i == m.app.Cursor {
			b.WriteString(shared.SelStyle.Render(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}
	return shared.ListStyle.Render(b.String())
}

func (m Model) renderPreview() string {
	if len(m.app.Notes) == 0 {
		return "(no notes)"
	}
	cur := m.app.Notes[m.app.Cursor]
	var b strings.Builder
	b.WriteString(shared.TitleStyle.Render("Front") + "\n")
	b.WriteString(cur.Front + "\n\n")
	b.WriteString(shared.TitleStyle.Render("Back") + "\n")
	b.WriteString(cur.Back + "\n")
	return lipgloss.NewStyle().Padding(0, 1).Render(b.String())
}
