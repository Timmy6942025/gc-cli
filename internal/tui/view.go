package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *model) View() string {
	styles := defaultStyles()

	title := styles.title.Render("Google Classroom CLI")
	subtitle := styles.muted.Render("Bubble Tea + Bubbles | Focus: " + m.focusLabel())

	leftStyle := styles.panel
	if m.focus == focusGlobal {
		leftStyle = styles.panelFocused
	}
	middleStyle := styles.panel
	if m.focus == focusCourses {
		middleStyle = styles.panelFocused
	}
	rightStyle := styles.panel
	if m.focus == focusTabs || m.focus == focusContent {
		rightStyle = styles.panelFocused
	}

	left := leftStyle.Width(m.leftWidth).Render(m.globalList.View())
	middle := middleStyle.Width(m.middleWidth).Render(m.courseList.View())

	rightSections := make([]string, 0, 4)
	if m.classMode {
		rightSections = append(rightSections, m.classTabs.View())
	}
	rightSections = append(rightSections, m.courseInfo.View(), m.content.View())
	right := rightStyle.Width(m.rightWidth).Render(strings.Join(rightSections, "\n\n"))

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, middle, right)
	footer := styles.status.Render(m.status)
	kb := m.help.View(m.keys)
	if m.commandMode {
		prompt := styles.command.Render(":" + m.commandInput.View())
		return strings.Join([]string{title, subtitle, "", body, "", footer, kb, prompt}, "\n")
	}
	return strings.Join([]string{title, subtitle, "", body, "", footer, kb}, "\n")
}
