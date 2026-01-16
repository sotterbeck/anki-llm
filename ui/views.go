package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	listStyle  = lipgloss.NewStyle().Padding(0, 1)
	selStyle   = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230"))
	titleStyle = lipgloss.NewStyle().Bold(true)
)

func (m *Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	if m.picking {
		return titleStyle.Render("Choose PDF") + "\n" + m.picker.View() + "\n" + m.renderFooter()
	}

	left := m.renderList()
	right := m.renderPreview()

	cols := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	footer := m.renderFooter()
	return titleStyle.Render("Select Notes") + "\n" + cols + "\n" + footer
}

func (m *Model) renderList() string {
	var b strings.Builder
	for i, it := range m.notes {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		chk := "[ ]"
		if m.selected[i] {
			chk = "[x]"
		}
		line := fmt.Sprintf("%s %s %s", cursor, chk, it.Front)
		if i == m.cursor {
			b.WriteString(selStyle.Render(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}
	return listStyle.Render(b.String())
}

func (m *Model) renderPreview() string {
	if len(m.notes) == 0 {
		return "(no notes)"
	}
	cur := m.notes[m.cursor]
	var b strings.Builder
	b.WriteString(titleStyle.Render("Front") + "\n")
	b.WriteString(cur.Front + "\n\n")
	b.WriteString(titleStyle.Render("Back") + "\n")
	b.WriteString(cur.Back + "\n")
	return lipgloss.NewStyle().Padding(0, 1).Render(b.String())
}

func (m *Model) renderFooter() string {
	hints := fmt.Sprintf("j/k:move  space:toggle  a:add  s:select-all d:change-deck (%s) r:regenerate  q:quit", m.deckName)
	if m.picking {
		hints = "up/down:move  enter:select  q:quit"
	}
	status := m.status
	sp := ""
	if m.loading {
		status = "working"
		sp = " " + m.spinner.View()
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(hints + "    " + status + sp)
}
