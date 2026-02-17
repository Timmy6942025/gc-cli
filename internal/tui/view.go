package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *model) View() string {
	styles := defaultStyles()

	title := styles.title.Render("Google Classroom CLI")
	subtitle := styles.muted.Render("Bubble Tea + Bubbles: list, table, viewport, help, key")

	left := styles.panel.Width(m.leftWidth).Render(m.globalList.View())
	middle := styles.panel.Width(m.middleWidth).Render(m.courseList.View())

	rightSections := make([]string, 0, 4)
	if m.classMode {
		rightSections = append(rightSections, m.classTabs.View())
	}
	rightSections = append(rightSections, m.courseInfo.View(), m.content.View())
	right := styles.panel.Width(m.rightWidth).Render(strings.Join(rightSections, "\n\n"))

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, middle, right)
	footer := styles.status.Render(m.status)
	kb := m.help.View(m.keys)

	return strings.Join([]string{title, subtitle, "", body, "", footer, kb}, "\n")
}
