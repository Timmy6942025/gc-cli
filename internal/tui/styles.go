package tui

import "github.com/charmbracelet/lipgloss"

type uiStyles struct {
	title  lipgloss.Style
	panel  lipgloss.Style
	muted  lipgloss.Style
	status lipgloss.Style
}

func defaultStyles() uiStyles {
	return uiStyles{
		title: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")),
		panel: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1),
		muted: lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		status: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("236")).
			Padding(0, 1),
	}
}
